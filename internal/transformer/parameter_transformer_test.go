/*
SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package transformer

import (
	"testing"

	"github.com/sap/cap-operator-lifecycle/api/v1alpha1"
	component "github.com/sap/component-operator-runtime/pkg/component"
	componentoperatorruntimetypes "github.com/sap/component-operator-runtime/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	mockPrometheusAddress = "http://prometheus.server.local:9090"
)

func TestTransformer(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)

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
		withCertManager                    bool
		withMonitoringGrafanaDashboard     bool
		withMaxConcurrentReconciles        bool
		// cache-related fields
		preResolvedDNSTarget      string // pre-populate status.ResolvedDNSTarget
		specGeneration            int64  // metadata.generation for CAPOperator (0 = default)
		appliedGeneration         int64  // status.appliedGeneration for CAPOperator (0 = default)
		skipPod                   bool   // omit the ingress gateway pod (for cache hit tests)
		expectedDNSTargetOverride string // if set, overrides the default expectedDnsTarget in assertions
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
			name:                        "With max concurrent reconciles",
			dnsTargetFilled:             true,
			withMaxConcurrentReconciles: true,
			expectError:                 false,
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
		{
			name:                       "With cert-manager",
			dnsTargetFilled:            false,
			ingressGatewayLabelsFilled: true,
			expectError:                false,
			withCertManager:            true,
		},
		{
			name:                           "With monitoring and grafana dashboard",
			dnsTargetFilled:                false,
			ingressGatewayLabelsFilled:     true,
			expectError:                    false,
			withCertManager:                true,
			withMonitoringGrafanaDashboard: true,
		},
		{
			name:                       "Cache hit - spec generation matches applied generation",
			ingressGatewayLabelsFilled: true,
			preResolvedDNSTarget:       "public-ingress.some.cluster.sap",
			specGeneration:             1,
			appliedGeneration:          1,
			skipPod:                    true, // pod not needed; cache hit must bypass getDNSTargetUsingIngressGatewayLabels
		},
		{
			name:                       "Cache miss on spec change - re-resolves DNS target",
			ingressGatewayLabelsFilled: true,
			preResolvedDNSTarget:       "old-dns-target.should-not-be-used",
			specGeneration:             2,
			appliedGeneration:          1,
		},
		{
			name:                      "Cache hit with explicit dnsTarget - uses persisted value over spec",
			dnsTargetFilled:           true,
			preResolvedDNSTarget:      "cached-ingress.some.cluster.sap",
			specGeneration:            1,
			appliedGeneration:         1,
			expectedDNSTargetOverride: "cached-ingress.some.cluster.sap",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)

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
				istioSvc,
			)

			if !tt.skipPod {
				clientBuilder.WithObjects(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingressGw",
						Namespace: "istio-system",
						Labels: map[string]string{
							"istio": "ingress",
							"app":   "istio-ingress",
						},
					},
				})
			}

			capOperator := &v1alpha1.CAPOperator{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "cap-operator",
					Namespace:  "cap-operator-system",
					Generation: tt.specGeneration,
				},
				Status: v1alpha1.CAPOperatorStatus{
					Status: component.Status{
						AppliedGeneration: tt.appliedGeneration,
					},
					ResolvedDNSTarget: tt.preResolvedDNSTarget,
				},
			}
			clientBuilder.WithStatusSubresource(capOperator).WithObjects(capOperator)

			kubeClient := clientBuilder.Build()

			parameterTransformer := NewParameterTransformer(kubeClient)

			capOperatorSpec := &v1alpha1.CAPOperatorSpec{}

			if tt.longDomain {
				capOperatorSpec.SubscriptionServer.Subdomain = "long-subdomain-for-the-test-to-fail-to-check-the-error-case"
			} else {
				capOperatorSpec.SubscriptionServer.Subdomain = "cop"
			}

			if tt.dnsTargetFilled {
				capOperatorSpec.DNSTarget = "public-ingress.some.cluster.sap"
			}

			if tt.withCertManager {
				certificateConfig := &v1alpha1.CertificateConfig{
					CertManager: v1alpha1.CertManager{
						IssuerName:  "abc",
						IssuerKind:  "abcKind",
						IssuerGroup: "abcGroup",
					},
				}

				capOperatorSpec.SubscriptionServer.CertificateManager = "CertManager"
				capOperatorSpec.Webhook.CertificateManager = "CertManager"
				capOperatorSpec.SubscriptionServer.CertificateConfig = certificateConfig
				capOperatorSpec.Webhook.CertificateConfig = certificateConfig
			} else {
				certificateConfig := &v1alpha1.CertificateConfig{
					Gardener: v1alpha1.Gardener{
						IssuerName:      "abc",
						IssuerNamespace: "abcNamespace",
					},
				}

				capOperatorSpec.SubscriptionServer.CertificateManager = "Gardener"
				capOperatorSpec.Webhook.CertificateManager = "Gardener"
				capOperatorSpec.SubscriptionServer.CertificateConfig = certificateConfig
				capOperatorSpec.Webhook.CertificateConfig = certificateConfig
			}

			if tt.ingressGatewayLabelsFilled {
				capOperatorSpec.IngressGatewayLabels = []v1alpha1.NameValue{
					{
						Name:  "app",
						Value: "istio-ingress",
					},
					{
						Name:  "istio",
						Value: "ingress",
					},
				}
			}

			if tt.withMaxConcurrentReconciles {
				capOperatorSpec.Controller.MaxConcurrentReconciles = &v1alpha1.MaxConcurrentReconciles{
					CAPApplication:        "2",
					CAPApplicationVersion: "5",
					CAPTenant:             "15",
					CAPTenantOperation:    "20",
					Domain:                "1",
				}
			}

			if tt.withVersionMonitoring {
				capOperatorSpec.Controller.VersionMonitoring = &v1alpha1.VersionMonitoring{
					PrometheusAddress: mockPrometheusAddress,
				}

				if !tt.omitVersionMonitoringDurations {
					capOperatorSpec.Controller.VersionMonitoring.MetricsEvaluationInterval = "5m"
					capOperatorSpec.Controller.VersionMonitoring.PromClientAcquireRetryDelay = "5h"
				}
			}

			if tt.withMonitoringGrafanaDashboard {
				capOperatorSpec.Monitoring = v1alpha1.Monitoring{
					Enabled: true,
					ServiceMonitorSelectorLabels: map[string]string{
						"release": "prometheus-operator",
					},
					Grafana: &v1alpha1.Grafana{
						Dashboard: &v1alpha1.GrafanaDashboard{
							ConfigMapLabels: map[string]string{
								"grafana_dashboard": "1",
							},
						},
					},
				}
			}

			transformedParameters, err := parameterTransformer.TransformParameters("cap-operator-system", "cap-operator.sme.sap.com", componentoperatorruntimetypes.UnstructurableMap(capOperatorSpec.ToUnstructured()))
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
			if tt.expectedDNSTargetOverride != "" {
				expectedDnsTarget = tt.expectedDNSTargetOverride
			} else if tt.withoutIngressGatewaySvcAnnotation {
				expectedDnsTarget = "x.some.cluster.sap"
			} else {
				expectedDnsTarget = "public-ingress.some.cluster.sap"
			}

			transformedParametersMap := transformedParameters.ToUnstructured()
			transformedSubscriptionServer := transformedParametersMap["subscriptionServer"].(map[string]any)
			transformedWebhook := transformedParametersMap["webhook"].(map[string]any)

			if transformedSubscriptionServer["dnsTarget"].(string) != expectedDnsTarget {
				t.Error("unexpected value returned for subscriptionServer.dnsTarget")
			}

			if transformedSubscriptionServer["domain"].(string) != "cop.some.cluster.sap" {
				t.Error("unexpected value returned for subscriptionServer.domain")
			}

			if tt.withCertManager {
				// subscription server checks
				if transformedSubscriptionServer["certificateManager"].(string) != "CertManager" {
					t.Error("expected subscriptionServer.certificateManager to be `CertManager`")
				}
				certificateConfig := transformedSubscriptionServer["certificateConfig"].(map[string]any)
				if len(certificateConfig["certManager"].(map[string]any)) == 0 {
					t.Error("expected subscriptionServer.certificateConfig.certManager not to be empty")
				}
				if len(certificateConfig["gardener"].(map[string]any)) != 0 {
					t.Error("expected subscriptionServer.certificateConfig.gardener to be empty")
				}

				// webhook checks
				if transformedWebhook["certificateManager"].(string) != "CertManager" {
					t.Error("expected webhook.certificateManager to be `CertManager`")
				}
				certificateConfig = transformedWebhook["certificateConfig"].(map[string]any)
				if len(certificateConfig["certManager"].(map[string]any)) == 0 {
					t.Error("expected webhook.certificateConfig.certManager not to be empty")
				}
				if len(certificateConfig["gardener"].(map[string]any)) != 0 {
					t.Error("expected webhook.certificateConfig.gardener to be empty")
				}
			} else {
				// subscription server checks
				if transformedSubscriptionServer["certificateManager"].(string) != "Gardener" {
					t.Error("expected subscriptionServer.certificateManager to be `Gardener`")
				}
				certificateConfig := transformedSubscriptionServer["certificateConfig"].(map[string]any)
				if len(certificateConfig["gardener"].(map[string]any)) == 0 {
					t.Error("expected subscriptionServer.certificateConfig.gardener not to be empty")
				}
				if len(certificateConfig["certManager"].(map[string]any)) != 0 {
					t.Error("expected subscriptionServer.certificateConfig.certManager to be empty")
				}

				// webhook checks
				if transformedWebhook["certificateManager"].(string) != "Gardener" {
					t.Error("expected webhook.certificateManager to be `Gardener`")
				}
				certificateConfig = transformedWebhook["certificateConfig"].(map[string]any)
				if len(certificateConfig["gardener"].(map[string]any)) == 0 {
					t.Error("expected webhook.certificateConfig.gardener not to be empty")
				}
				if len(certificateConfig["certManager"].(map[string]any)) != 0 {
					t.Error("expected webhook.certificateConfig.certManager to be empty")
				}
			}

			transformedController := transformedParametersMap["controller"].(map[string]any)
			if transformedController["dnsTarget"].(string) != expectedDnsTarget {
				t.Error("unexpected value returned for controller.dnsTarget")
			}

			if tt.withMaxConcurrentReconciles {
				if transformedController["maxConcurrentReconciles"] == nil {
					t.Error("expected controller.maxConcurrentReconciles to be filled")
				} else {
					mcr := transformedController["maxConcurrentReconciles"].(map[string]any)
					if mcr["capApplication"] != "2" || mcr["capApplicationVersion"] != "5" || mcr["capTenant"] != "15" || mcr["capTenantOperation"] != "20" || mcr["domain"] != "1" {
						t.Error("unexpected values returned for controller.maxConcurrentReconciles")
					}
				}
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

			if tt.withMonitoringGrafanaDashboard {
				dashboard := transformedParametersMap["monitoring"].(map[string]any)["grafana"].(map[string]any)["dashboard"].(map[string]any)
				if dashboard["configMapLabels"].(map[string]any)["grafana_dashboard"] != "1" {
					t.Error("expected monitoring.grafana.dashboard.configMapLabels to be set")
				}
				serviceMonitorSelectorLabels := transformedParametersMap["monitoring"].(map[string]any)["serviceMonitorSelectorLabels"].(map[string]any)
				if serviceMonitorSelectorLabels["release"] != "prometheus-operator" {
					t.Error("expected monitoring.serviceMonitorSelectorLabels to be set")
				}
			}
		})
	}
}
