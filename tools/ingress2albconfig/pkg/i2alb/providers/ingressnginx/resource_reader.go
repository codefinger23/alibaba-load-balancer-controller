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

package ingressnginx

import (
	"context"

	i2alb "k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb"
	"k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb/providers/common"
)

// converter implements the i2gw.CustomResourceReader interface.
type resourceReader struct {
	conf *i2alb.ProviderConf
}

// newResourceReader returns a resourceReader instance.
func newResourceReader(conf *i2alb.ProviderConf) *resourceReader {
	return &resourceReader{
		conf: conf,
	}
}

func (r *resourceReader) readResourcesFromCluster(ctx context.Context) (*storage, error) {
	storage := newResourcesStorage()

	ingresses, err := common.ReadIngressesFromCluster(ctx, r.conf.Client, NginxIngressClass)
	if err != nil {
		return nil, err
	}
	storage.Ingresses = ingresses
	return storage, nil
}

func (r *resourceReader) readResourcesFromFile(filename string) (*storage, error) {
	storage := newResourcesStorage()

	ingresses, err := common.ReadIngressesFromFile(filename, r.conf.Namespace, NginxIngressClass)
	if err != nil {
		return nil, err
	}
	storage.Ingresses = ingresses
	return storage, nil
}
