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
	"fmt"
	"regexp"

	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

func GetIngressClass(ingress networkingv1.Ingress) string {
	var ingressClass string

	if ingress.Spec.IngressClassName != nil && *ingress.Spec.IngressClassName != "" {
		ingressClass = *ingress.Spec.IngressClassName
	} else if _, ok := ingress.Annotations[networkingv1beta1.AnnotationIngressClass]; ok {
		ingressClass = ingress.Annotations[networkingv1beta1.AnnotationIngressClass]
	} else {
		ingressClass = ingress.Name
	}

	return ingressClass
}

type IngressRuleGroup struct {
	Namespace    string
	Name         string
	IngressClass string
	Host         string
	TLS          []networkingv1.IngressTLS
	Rules        []Rule
}

type Rule struct {
	Ingress     networkingv1.Ingress
	IngressRule networkingv1.IngressRule
}

func GetRuleGroups(ingresses []networkingv1.Ingress) map[string]IngressRuleGroup {
	ruleGroups := make(map[string]IngressRuleGroup)

	for _, ingress := range ingresses {
		ingressClass := GetIngressClass(ingress)

		for _, rule := range ingress.Spec.Rules {

			rgKey := ingressClass
			rg, ok := ruleGroups[rgKey]
			if !ok {
				rg = IngressRuleGroup{
					Namespace:    ingress.Namespace,
					Name:         ingress.Name,
					IngressClass: ingressClass,
					Host:         rule.Host,
				}
				ruleGroups[rgKey] = rg
			}
			rg.TLS = append(rg.TLS, ingress.Spec.TLS...)
			rg.Rules = append(rg.Rules, Rule{
				Ingress:     ingress,
				IngressRule: rule,
			})

			ruleGroups[rgKey] = rg
		}

	}

	return ruleGroups
}

func NameFromHost(host string) string {
	// replace all special chars with -
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	step1 := reg.ReplaceAllString(host, "-")
	// remove all - at start of string
	reg2, _ := regexp.Compile("^[^a-zA-Z0-9]+")
	step2 := reg2.ReplaceAllString(step1, "")
	// if nothing left, return "all-hosts"
	if len(host) == 0 || host == "*" {
		return "all-hosts"
	}
	return step2
}

func RouteName(ingressName, host string) string {
	return fmt.Sprintf("%s-%s", ingressName, NameFromHost(host))
}

func PtrTo[T any](a T) *T {
	return &a
}
