/*
SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and cap-operator contributors
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

const (
	mockPrometheusAddress = "http://prometheus.server.local:9090"
)

func TestTransformer(t *testing.T) {
	tests := []struct {
		name                               string
		dnsTargetFilled                    bool
		ingressGatewayLabelsFilled         bool
		longDomain                         bool
		expectError                        bool
		withoutIngressGatewaySvcAnnotation bool
		withVersionMonitoring              bool
		omitVersionMonitoringDurations     bool
		withControllerVolumes              bool
	}{
		{
			name:                       "With dnsTarget and without ingress gateway labels",
			dnsTargetFilled:            true,
			ingressGatewayLabelsFilled: false,
			expectError:                false,
		},
		{
			name:                       "Without dnsTarget and with ingress gateway labels",
			dnsTargetFilled:            false,
			ingressGatewayLabelsFilled: true,
			expectError:                false,
		},
		{
			name:                       "Without dnsTarget and ingress gateway labels",
			dnsTargetFilled:            false,
			ingressGatewayLabelsFilled: false,
			expectError:                true,
		},
		{
			name:                       "With more than 64 character domain",
			ingressGatewayLabelsFilled: true,
			longDomain:                 true,
			expectError:                true,
		},
		{
			name:                               "Without annotation in ingress gateway labels",
			ingressGatewayLabelsFilled:         true,
			longDomain:                         false,
			withoutIngressGatewaySvcAnnotation: true,
		},
		{
			name:                       "With version monitoring and dnsTarget filled",
			dnsTargetFilled:            true,
			ingressGatewayLabelsFilled: false,
			expectError:                false,
			withVersionMonitoring:      true,
		},
		{
			name:                           "With version monitoring and ingress labels filled",
			dnsTargetFilled:                false,
			ingressGatewayLabelsFilled:     true,
			expectError:                    false,
			withVersionMonitoring:          true,
			omitVersionMonitoringDurations: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clientBuilder := fake.NewClientBuilder()

			istioSvc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istioingress-gateway",
					Namespace: "istio-system",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Selector: map[string]string{
						"istio": "ingress",
						"app":   "istio-ingress",
					},
				},
			}

			if !tt.withoutIngressGatewaySvcAnnotation {
				istioSvc.ObjectMeta.Annotations = map[string]string{
					"dns.gardener.cloud/dnsnames": "public-ingress.some.cluster.sap",
				}
			}

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
				istioSvc)

			kubeClient := clientBuilder.Build()

			transformer := NewParameterTransformer(kubeClient)

			parameter := make(map[string]interface{})

			parameter["subscriptionServer"] = map[string]interface{}{}
			subscriptionServer := parameter["subscriptionServer"].(map[string]interface{})

			if tt.longDomain {
				subscriptionServer["subDomain"] = "long-subdomain-for-the-test-to-fail-to-check-the-error-case"
			} else {
				subscriptionServer["subDomain"] = "cop"
			}
			if tt.dnsTargetFilled {
				parameter["dnsTarget"] = "public-ingress.some.cluster.sap"
			}

			if tt.ingressGatewayLabelsFilled {
				parameter["ingressGatewayLabels"] = []interface{}{
					map[string]interface{}{
						"name":  "app",
						"value": "istio-ingress",
					},
					map[string]interface{}{
						"name":  "istio",
						"value": "ingress",
					},
				}
			}

			if tt.withVersionMonitoring {
				parameter["controller"] = map[string]any{
					"versionMonitoring": map[string]any{
						"prometheusAddress": mockPrometheusAddress,
					},
				}
				if !tt.omitVersionMonitoringDurations {
					controller := parameter["controller"].(map[string]any)
					vm := controller["versionMonitoring"].(map[string]any)
					vm["metricsEvaluationInterval"] = "5m"
					vm["promClientAcquireRetryDelay"] = "5h"
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

			var expectedDnsTarget string
			if tt.withoutIngressGatewaySvcAnnotation {
				expectedDnsTarget = "x.some.cluster.sap"
			} else {
				expectedDnsTarget = "public-ingress.some.cluster.sap"
			}

			transformedParametersMap := transformedParameters.ToUnstructured()
			if transformedParametersMap["subscriptionServer"].(map[string]interface{})["dnsTarget"].(string) != expectedDnsTarget {
				t.Error("unexpected value returned for subscriptionServer.dnsTarget")
			}
			if transformedParametersMap["subscriptionServer"].(map[string]interface{})["domain"].(string) != "cop.some.cluster.sap" {
				t.Error("unexpected value returned for subscriptionServer.domain")
			}
			transformedController := transformedParametersMap["controller"].(map[string]interface{})
			if transformedController["dnsTarget"].(string) != expectedDnsTarget {
				t.Error("unexpected value returned for controller.dnsTarget")
			}
			if tt.withVersionMonitoring {
				if transformedController["versionMonitoring"] == nil {
					t.Error("expected controller.versionMonitoring to be filled")
				} else {
					tvm := transformedController["versionMonitoring"].(map[string]any)
					if tvm["prometheusAddress"] != mockPrometheusAddress {
						t.Error("expected controller.versionMonitoring.prometheusAddress to be set")
					}
					if tt.omitVersionMonitoringDurations {
						if tvm["metricsEvaluationInterval"] != nil || tvm["promClientAcquireRetryDelay"] != nil {
							t.Error("expected version monitoring durations to be unset")
						}
					}
				}
			}
		})
	}
}
