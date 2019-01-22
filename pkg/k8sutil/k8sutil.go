/*
Copyright © 2018 inwinSTACK.inc

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

package k8sutil

import (
	"fmt"
	"reflect"

	inwinv1 "github.com/inwinstack/blended/apis/inwinstack/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("master", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func FilterIPsByPool(ips *inwinv1.IPList, pool *inwinv1.Pool) {
	var items []inwinv1.IP
	for _, ip := range ips.Items {
		if ip.Spec.PoolName == pool.Name {
			items = append(items, ip)
		}
	}
	ips.Items = items
}

func NewIPWithNamespace(ns *v1.Namespace, poolName string) *inwinv1.IP {
	return &inwinv1.IP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s", uuid.NewUUID()),
			Namespace: ns.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ns, schema.GroupVersionKind{
					Group:   v1.SchemeGroupVersion.Group,
					Version: v1.SchemeGroupVersion.Version,
					Kind:    reflect.TypeOf(v1.Namespace{}).Name(),
				}),
			},
		},
		Spec: inwinv1.IPSpec{
			PoolName:             poolName,
			MarkNamespaceRefresh: true,
		},
	}
}

func NewPool(name string, addrs, namespaces []string) *inwinv1.Pool {
	return &inwinv1.Pool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: inwinv1.PoolSpec{
			Addresses:                 addrs,
			IgnoreNamespaces:          namespaces,
			IgnoreNamespaceAnnotation: false,
			AssignToNamespace:         true,
			AvoidBuggyIPs:             true,
			AvoidGatewayIPs:           false,
		},
	}
}
