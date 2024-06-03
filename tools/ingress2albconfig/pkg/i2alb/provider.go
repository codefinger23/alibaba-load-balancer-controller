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

package i2alb

import (
	"context"

	"k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb/albconfig"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ProviderConstructorByName is a map of ProviderConstructor functions by a
// provider name. Different Provider implementations should add their construction
// func at startup.
var ProviderConstructorByName = map[ProviderName]ProviderConstructor{}

// ProviderName is a string alias that stores the concrete Provider name.
type ProviderName string

// ProviderConstructor is a construction function that constructs concrete
// implementations of the Provider interface.
type ProviderConstructor func(conf *ProviderConf) Provider

// ProviderConf contains all the configuration required for every concrete
// Provider implementation.
type ProviderConf struct {
	Client    client.Client
	Namespace string
}

// The Provider interface specifies the required functionality which needs to be
// implemented by every concrete Ingress/Gateway-API provider, in order for it to
// be used.
type Provider interface {
	CustomResourceReader
	ResourceConverter
}

type CustomResourceReader interface {

	// ReadResourcesFromCluster reads custom resources associated with
	// the underlying Provider implementation from the kubernetes cluster.
	ReadResourcesFromCluster(ctx context.Context) error

	// ReadResourcesFromFile reads custom resources associated with
	// the underlying Provider implementation from the file.
	ReadResourcesFromFile(ctx context.Context, filename string) error
}

// The ResourceConverter interface specifies all the implemented Gateway API resource
// conversion functions.
type ResourceConverter interface {

	// ToGatewayAPIResources converts the received InputResources associated
	// with the Provider into GatewayResources.
	ToAlbConfig() (AlbResources, field.ErrorList)
}

type AlbImplement struct {
	AliasTls bool
}

// GatewayResources contains all Gateway-API objects.
type AlbResources struct {
	AlbConfigs     []*albconfig.AlbConfig
	IngressClasses []networkingv1.IngressClass
	Ingresses      map[types.NamespacedName]*networkingv1.Ingress
}

// FeatureParser is a function that reads the InputResources, and applies
// the appropriate modifications to the AlbResources.
//
// Different FeatureParsers will run in undetermined order. The function must
// modify / create only the required fields of the albconfig resources and nothing else.
type FeatureParser func([]networkingv1.Ingress, *AlbResources) field.ErrorList
