/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/
package transformer

import (
	"testing"

	componentoperatorruntimetypes "github.com/sap/component-operator-runtime/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestTransformer(t *testing.T) {
	tests := []struct {
		name                       string
		dnsTargetFilled            bool
		ingressGatewayLabelsFilled bool
		expectError                bool
	}{
		{
			name:                       "Test with all fields filled",
			dnsTargetFilled:            true,
			ingressGatewayLabelsFilled: true,
			expectError:                false,
		},
		{
			name:                       "Test without dnsTarget",
			dnsTargetFilled:            false,
			ingressGatewayLabelsFilled: true,
			expectError:                false,
		},
		{
			name:                       "Test without dnsTarget and ingress gateway labels",
			dnsTargetFilled:            false,
			ingressGatewayLabelsFilled: false,
			expectError:                true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clientBuilder := fake.NewClientBuilder()
			clientBuilder.WithObjects(
				&corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      shootInfoConfigMap,
						Namespace: kubeSystemNamespace,
					},
					Data: map[string]string{
						"domain": "some.cluster.sap",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingressGw",
						Namespace: "istio-system",
						Labels: map[string]string{
							"istio": "ingress",
							"app":   "istio-ingress",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "istioingress-gateway",
						Namespace: "istio-system",
						Annotations: map[string]string{
							"dns.gardener.cloud/dnsnames": "public-ingress.some.cluster.sap",
						},
					},
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeLoadBalancer,
						Selector: map[string]string{
							"istio": "ingress",
							"app":   "istio-ingress",
						},
					},
				})

			kubeClient := clientBuilder.Build()

			transformer := NewParameterTransformer(kubeClient)

			parameter := make(map[string]interface{})

			parameter["spec"] = map[string]interface{}{}

			parameter["subscriptionServer"] = map[string]interface{}{}
			subscriptionServer := parameter["subscriptionServer"].(map[string]interface{})
			subscriptionServer["subDomain"] = "cop"
			if tt.dnsTargetFilled {
				subscriptionServer["dnsTarget"] = "public-ingress.some.cluster.sap"
			}

			if tt.ingressGatewayLabelsFilled {
				parameter["ingressGatewayLabels"] = map[string]interface{}{
					"istio": "ingress",
					"app":   "istio-ingress",
				}
			}

			transformedParameters, err := transformer.TransformParameters("cap-operator-system", "cap-operator.sme.sap.com", componentoperatorruntimetypes.UnstructurableMap(parameter))
			if !tt.expectError && err != nil {
				t.Error(err)
			}

			if tt.expectError && err == nil {
				t.Error("error expected but not returned")
			}

			if tt.expectError && err != nil {
				t.Log(err)
				return
			}
			transformedParametersMap := transformedParameters.ToUnstructured()
			if transformedParametersMap["subscriptionServer"].(map[string]interface{})["dnsTarget"].(string) != "public-ingress.some.cluster.sap" {
				t.Error("unexpected value returned")
			}
			if transformedParametersMap["subscriptionServer"].(map[string]interface{})["domain"].(string) != "cop.some.cluster.sap" {
				t.Error("unexpected value returned")
			}
		})
	}
}
