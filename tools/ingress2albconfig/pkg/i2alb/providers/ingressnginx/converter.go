package ingressnginx

import (
	i2alb "k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb"
	"k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb/providers/common"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// converter implements the ToGatewayAPI function of i2alb.ResourceConverter interface.
type converter struct {
	featureParsers []i2alb.FeatureParser
}

// newConverter returns an ingress-nginx converter instance.
func newConverter() *converter {
	return &converter{
		featureParsers: []i2alb.FeatureParser{
			rewriteFeature,
		},
	}
}

func (c *converter) convert(storage *storage) (i2alb.AlbResources, field.ErrorList) {

	// TODO(liorliberman) temporary until we decide to change ToGateway and featureParsers to get a map of [types.NamespacedName]*networkingv1.Ingress instead of a list
	ingressList := []networkingv1.Ingress{}
	for _, ing := range storage.Ingresses {
		ingressList = append(ingressList, *ing)
	}

	// Convert plain ingress resources to gateway resources, ignoring all provider-specific features.
	albResources, errs := common.ToAlbIngress(ingressList)

	if len(errs) > 0 {
		return i2alb.AlbResources{}, errs
	}

	for _, parseFeatureFunc := range c.featureParsers {
		// Apply the feature parsing function to the gateway resources, one by one.
		parseErrs := parseFeatureFunc(ingressList, &albResources)

		// Append the parsing errors to the error list.
		errs = append(errs, parseErrs...)
	}

	return albResources, errs
}
