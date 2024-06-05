/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"encoding/json"
	"fmt"
	"strings"

	i2alb "k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb"
	"k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb/albconfig"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ToAlb converts the received ingresses to i2alb.AlbResources,
// without taking into consideration any provider specific logic.
func ToAlbIngress(ingresses []networkingv1.Ingress) (i2alb.AlbResources, field.ErrorList) {
	aggregator := ingressAggregator{ingressGroups: map[ingressClassKey][]ingressRuleGroup{}}

	var errs field.ErrorList
	for _, ingress := range ingresses {
		aggregator.addIngress(ingress)
	}
	if len(errs) > 0 {
		return i2alb.AlbResources{}, errs
	}

	albIngresses, albconfigs, errs := aggregator.toAlbIngressAndConfig()

	if len(errs) > 0 {
		return i2alb.AlbResources{}, errs
	}

	ingressByKey := make(map[types.NamespacedName]*networkingv1.Ingress)
	for _, albIngress := range albIngresses {
		key := types.NamespacedName{Namespace: albIngress.Namespace, Name: albIngress.Name}
		ingressByKey[key] = albIngress
	}

	ingressClasses := []networkingv1.IngressClass{}
	for _, albconfig := range albconfigs {
		ic := networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: albconfig.Name,
			},
			Spec: networkingv1.IngressClassSpec{
				Controller: "ingress.k8s.alibabacloud/alb",
				Parameters: &networkingv1.IngressClassParametersReference{
					APIGroup: &AlbConfigGVK.Group,
					Kind:     AlbConfigGVK.Kind,
					Name:     albconfig.Name,
				},
			},
		}
		ic.SetGroupVersionKind(networkingv1.SchemeGroupVersion.WithKind("IngressClass"))
		ingressClasses = append(ingressClasses, ic)
	}

	return i2alb.AlbResources{
		AlbConfigs:     albconfigs,
		IngressClasses: ingressClasses,
		Ingresses:      ingressByKey,
	}, nil
}

var (
	AlbConfigGVK = schema.GroupVersionKind{
		Group:   "alibabacloud.com",
		Version: "v1",
		Kind:    "AlbConfig",
	}
)

type ingressClassKey string

type ingressAggregator struct {
	ingressGroups   map[ingressClassKey][]ingressRuleGroup
	defaultBackends []ingressDefaultBackend
}

type ingressRuleGroup struct {
	namespace    string
	name         string
	ingressClass string
	tls          []networkingv1.IngressTLS
	ingress      networkingv1.Ingress
}

type ingressDefaultBackend struct {
	name         string
	namespace    string
	ingressClass string
	ingress      networkingv1.Ingress
	backend      networkingv1.IngressBackend
}

func (a *ingressAggregator) addIngress(ingress networkingv1.Ingress) {
	var ingressClass string
	if ingress.Spec.IngressClassName != nil && *ingress.Spec.IngressClassName != "" {
		ingressClass = *ingress.Spec.IngressClassName
	} else if _, ok := ingress.Annotations[networkingv1beta1.AnnotationIngressClass]; ok {
		ingressClass = ingress.Annotations[networkingv1beta1.AnnotationIngressClass]
	} else {
		ingressClass = ingress.Name
	}
	rgsKey := ingressClassKey(acov(ingressClass))
	rgs, ok := a.ingressGroups[rgsKey]
	if !ok {
		rgs = []ingressRuleGroup{}
	}
	rg := ingressRuleGroup{
		namespace:    ingress.Namespace,
		name:         ingress.Name,
		ingressClass: ingressClass,
	}

	rg.tls = ingress.Spec.TLS
	rg.ingress = ingress
	rgs = append(rgs, rg)
	a.ingressGroups[rgsKey] = rgs

	if ingress.Spec.DefaultBackend != nil {
		a.defaultBackends = append(a.defaultBackends, ingressDefaultBackend{
			name:         ingress.Name,
			namespace:    ingress.Namespace,
			ingressClass: ingressClass,
			backend:      *ingress.Spec.DefaultBackend,
			ingress:      ingress,
		})
	}
}

func UnMatchTLS(tls []networkingv1.IngressTLS, rule *networkingv1.IngressRule) bool {
	for _, tls := range tls {
		for _, host := range tls.Hosts {
			if rule.Host == host {
				return false
			}
		}
	}
	return true
}

func (a *ingressAggregator) toAlbIngressAndConfig() ([]*networkingv1.Ingress, []*albconfig.AlbConfig, field.ErrorList) {
	var errors field.ErrorList
	var albIngresses []*networkingv1.Ingress

	listenersByName := map[string]map[string]albconfig.ListenerSpec{}

	for _, rgs := range a.ingressGroups {
		for _, rg := range rgs {
			albconfigKey := rg.ingressClass
			if listenersByName[albconfigKey] == nil {
				listenersByName[albconfigKey] = map[string]albconfig.ListenerSpec{
					"80": {Port: intstr.FromInt(80), Protocol: "HTTP"},
				}
			}
			options := &i2alb.AlbImplement{
				AliasTls: false,
			}
			for _, rule := range rg.ingress.Spec.Rules {
				if !UnMatchTLS(rg.ingress.Spec.TLS, &rule) {
					options.AliasTls = true
					break
				}
			}
			if options.AliasTls {
				listenersByName[albconfigKey]["443"] = albconfig.ListenerSpec{Port: intstr.FromInt(443), Protocol: "HTTPS"}
			}
			albIngress, errs := rg.convertAlbIngress(&rg.ingress, options)
			if len(errs) > 0 {
				errors = append(errors, errs...)
				continue
			}
			albIngresses = append(albIngresses, albIngress)
		}
	}

	for _, db := range a.defaultBackends {
		options := &i2alb.AlbImplement{
			AliasTls: false,
		}
		for _, rule := range db.ingress.Spec.Rules {
			if !UnMatchTLS(db.ingress.Spec.TLS, &rule) {
				options.AliasTls = true
				break
			}
		}
		albIngress, errs := db.convertAlbIngress(options)
		if len(errs) > 0 {
			errors = append(errors, errs...)
			continue
		}
		albIngresses = append(albIngresses, albIngress)
	}

	albconfigByKey := map[string]*albconfig.AlbConfig{}
	for albconfigKey, listeners := range listenersByName {
		albc := albconfigByKey[albconfigKey]
		if albc == nil {
			albc = &albconfig.AlbConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: albconfigKey,
				},
				Spec: albconfig.AlbConfigSpec{
					LoadBalancer: &albconfig.LoadBalancerSpec{
						Name:    albconfigKey,
						Edition: "Standard",
						Tags: []albconfig.Tag{
							{
								Key:   "converted/ingress2albconfig",
								Value: "true",
							},
						},
					},
					Listeners: []*albconfig.ListenerSpec{},
				},
			}
			albc.SetGroupVersionKind(AlbConfigGVK)
			albconfigByKey[albconfigKey] = albc
		}
		for _, listener := range listeners {
			albc.Spec.Listeners = append(albc.Spec.Listeners, &albconfig.ListenerSpec{
				Port:     listener.Port,
				Protocol: listener.Protocol,
			})
		}

	}

	var albconfigs []*albconfig.AlbConfig
	for _, alb := range albconfigByKey {
		albconfigs = append(albconfigs, alb)
	}

	return albIngresses, albconfigs, errors
}

func (rg *ingressRuleGroup) convertAlbIngress(ing *networkingv1.Ingress, options *i2alb.AlbImplement) (*networkingv1.Ingress, field.ErrorList) {

	annotations := ing.DeepCopyObject().(metav1.Object).GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	albIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        rg.name,
			Namespace:   rg.namespace,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{},
		},
	}
	albIngress.SetGroupVersionKind(networkingv1.SchemeGroupVersion.WithKind("Ingress"))

	if rg.ingressClass != "" {
		albIngress.Spec.IngressClassName = &rg.ingressClass
	}

	var errors field.ErrorList
	for i, rule := range ing.Spec.Rules {
		newRule := networkingv1.IngressRule{
			Host: rule.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{},
				},
			},
		}
		if rule.HTTP == nil {
			errors = append(errors, field.Invalid(field.NewPath("Spec", "Rules", fmt.Sprint(i)), rule.Host,
				fmt.Sprintf("empty http rule find: %s/%s", ing.Namespace, ing.Name)))
			return nil, errors
		}
		for _, path := range rule.HTTP.Paths {
			newPath := networkingv1.HTTPIngressPath{
				Backend: path.Backend,
			}
			switch *path.PathType {
			case networkingv1.PathTypePrefix:
			case networkingv1.PathTypeExact:
				newPath.Path = path.Path
				newPath.PathType = path.PathType
			case networkingv1.PathTypeImplementationSpecific:
				normalizedPath := strings.TrimSuffix(path.Path, "/")
				newPath.Path = normalizedPath + "/*"
				newPath.PathType = path.PathType
			default:
				err := field.Invalid(field.NewPath("spec", "rule", "http", "path", "pathType"), path.PathType, fmt.Sprintf("unsupported path match type: %s", *path.PathType))
				errors = append(errors, err)
			}
			newRule.HTTP.Paths = append(newRule.HTTP.Paths, newPath)
		}

		albIngress.Spec.Rules = append(albIngress.Spec.Rules, newRule)
	}

	listenPorts := []map[string]int{}
	listenPorts = append(listenPorts, map[string]int{"HTTP": 80})

	if options.AliasTls {
		listenPorts = append(listenPorts, map[string]int{"HTTPS": 443})
	}

	listenPortsStr, err := json.Marshal(listenPorts)
	if err != nil {
		errors = append(errors, field.Invalid(field.NewPath("ObjectMeta", "Annotations"), listenPorts,
			fmt.Sprintf("error json Marshal A for key: %s/%s(%s)", ing.Namespace, ing.Name, err.Error())))
		return nil, errors
	}

	albIngress.Annotations["alb.ingress.kubernetes.io/listen-ports"] = string(listenPortsStr)

	return albIngress, errors
}

func (db *ingressDefaultBackend) convertAlbIngress(options *i2alb.AlbImplement) (*networkingv1.Ingress, field.ErrorList) {
	var errors field.ErrorList
	if db.backend.Service == nil {
		errors = append(errors, field.Invalid(field.NewPath("Spec", "defaultBackend"), db.backend,
			fmt.Sprintf("not support non-service defaultBackend: %s/%s", db.namespace, db.name)))
		return nil, errors
	}
	name := fmt.Sprintf("%s__default", db.name)
	pathType := networkingv1.PathTypePrefix
	albIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   db.namespace,
			Annotations: map[string]string{},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: db.backend.Service.Name,
											Port: db.backend.Service.Port,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	albIngress.SetGroupVersionKind(networkingv1.SchemeGroupVersion.WithKind("Ingress"))

	if db.ingressClass != "" {
		albIngress.Spec.IngressClassName = &db.ingressClass
	}

	listenPorts := []map[string]int{}
	listenPorts = append(listenPorts, map[string]int{"HTTP": 80})

	if options.AliasTls {
		listenPorts = append(listenPorts, map[string]int{"HTTPS": 443})
	}

	listenPortsStr, err := json.Marshal(listenPorts)
	if err != nil {
		errors = append(errors, field.Invalid(field.NewPath("ObjectMeta", "Annotations"), listenPorts,
			fmt.Sprintf("error json Marshal A for key: %s/%s(%s)", db.namespace, db.name, err.Error())))
		return nil, errors
	}

	albIngress.Annotations["alb.ingress.kubernetes.io/listen-ports"] = string(listenPortsStr)

	return albIngress, errors
}

func acov(key string) string {
	prefix := "alb__"
	return fmt.Sprintf("%s%s", prefix, key)
}
