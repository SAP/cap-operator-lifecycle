/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/
package manifests

import (
	"testing"

	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
)

func ContainsFunc(clientObjects []client.Object, name string, kind string) bool {
	return slices.ContainsFunc(clientObjects, func(object client.Object) bool {
		return object.GetName() == name && object.GetObjectKind().GroupVersionKind().Kind == kind
	})
}

func TestHelmResourceGenerator(t *testing.T) {
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

			fakeclientset := fakeclientset.NewSimpleClientset()
			fakeDiscovery, ok := fakeclientset.Discovery().(*fakediscovery.FakeDiscovery)
			if !ok {
				t.Error("fake discovery client creation failed")
			}
			fakeDiscovery.Resources = []*metav1.APIResourceList{
				{
					GroupVersion: "cert.gardener.cloud/v1alpha1",
					TypeMeta: metav1.TypeMeta{
						Kind:       "APIResourceList",
						APIVersion: "v1",
					},
					APIResources: []metav1.APIResource{
						{
							Name:         "certificates",
							SingularName: "certificate",
							Namespaced:   true,
							Kind:         "Certificate",
						},
					},
				},
				{
					GroupVersion: "dns.gardener.cloud/v1alpha1",
					TypeMeta: metav1.TypeMeta{
						Kind:       "APIResourceList",
						APIVersion: "v1",
					},
					APIResources: []metav1.APIResource{
						{
							Name:         "dnsentries",
							SingularName: "dnsentry",
							Namespaced:   true,
							Kind:         "DNSEntry",
						},
					},
				},
			}

			helmGenerator, err := NewHelmGenerator("cap-operator.sme.sap.com", nil, "../../chart", kubeClient, fakeDiscovery)
			if err != nil {
				t.Error(err)
			}

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

			clientObjects, err := helmGenerator.Generate("cap-operator-system", "cap-operator", &unstructurableMap{parameter})
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

			if len(clientObjects) != 24 {
				t.Error("wrong number of returned client objects")
			}

			if !ContainsFunc(clientObjects, "cap-operator-controller", "Deployment") {
				t.Error("controller deployment not found")
			}

			if !ContainsFunc(clientObjects, "cap-operator-subscription-server", "Deployment") {
				t.Error("subscription-server deployment not found")
			}

			if !ContainsFunc(clientObjects, "cap-operator-webhook", "Deployment") {
				t.Error("webhook deployment not found")
			}

			if !ContainsFunc(clientObjects, "cap-operator-subscription-server", "VirtualService") {
				t.Error("virutal service not found")
			}

			if !ContainsFunc(clientObjects, "cap-operator-subscription-server", "Gateway") {
				t.Error("gateway not found")
			}

			if !ContainsFunc(clientObjects, "cap-operator-subscription-server", "DNSEntry") {
				t.Error("DNSEntry not found")
			}
		})
	}
}
