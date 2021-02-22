// Copyright 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sidecar

import (
	"emperror.dev/errors"
	"github.com/banzaicloud/operator-tools/pkg/merge"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/thanos-operator/pkg/sdk/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type endpointService struct {
	*v1alpha1.StoreEndpoint
}

func (e *endpointService) sidecarService() (runtime.Object, reconciler.DesiredState, error) {
	if e.Spec.Selector != nil {
		var grpcPort int32 = 10901
		var httpPort int32 = 10902
		labels := map[string]string{
			"app": "prometheus",
		}
		if e.Spec.Selector.GRPCPort != 0 {
			grpcPort = e.Spec.Selector.GRPCPort
		}
		if e.Spec.Selector.HTTPPort != 0 {
			httpPort = e.Spec.Selector.HTTPPort
		}
		if e.Spec.Selector.Labels != nil {
			labels = e.Spec.Selector.Labels
		}
		service := &corev1.Service{
			ObjectMeta: e.getMeta(),
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:     "grpc",
						Protocol: corev1.ProtocolTCP,
						Port:     grpcPort,
						TargetPort: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "grpc",
						},
					},
					{
						Name:     "http",
						Protocol: corev1.ProtocolTCP,
						Port:     httpPort,
						TargetPort: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "http",
						},
					},
				},
				Selector:  labels,
				Type:      corev1.ServiceTypeClusterIP,
				ClusterIP: corev1.ClusterIPNone,
			},
		}
		if e.Spec.ServiceOverrides != nil {
			err := merge.Merge(service, e.Spec.ServiceOverrides)
			if err != nil {
				return service, reconciler.StatePresent, errors.WrapIf(err, "unable to merge overrides to base object")
			}
		}

		return service, reconciler.StatePresent, nil
	}
	delete := &corev1.Service{
		ObjectMeta: e.getMeta(),
	}
	return delete, reconciler.StateAbsent, nil
}
