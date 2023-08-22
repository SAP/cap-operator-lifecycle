/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package manifests

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sap/component-operator-runtime/pkg/manifests"
	"github.com/sap/component-operator-runtime/pkg/types"

	"github.com/sap/cap-operator-lifecycle/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
)

const (
	kubeSystemNamespace     = "kube-system"
	shootInfoConfigMap      = "shoot-info"
	istioIngressGWNamespace = "istio-system"
	annotationDNSNames      = "dns.gardener.cloud/dnsnames"
)

type HelmGenerator struct {
	resourceGenerator *manifests.HelmGenerator
	client            client.Client
}

type unstructurableMap struct {
	data map[string]any
}

func (m *unstructurableMap) ToUnstructured() map[string]any {
	return runtime.DeepCopyJSON(m.data)
}

var _ Generator = &HelmGenerator{}

func NewHelmGenerator(name string, fsys fs.FS, chartPath string, client client.Client, discoveryClient discovery.DiscoveryInterface) (*HelmGenerator, error) {
	resourceGenerator, err := manifests.NewHelmGenerator(name, fsys, chartPath, client, discoveryClient)
	if err != nil {
		return nil, err
	}
	g := HelmGenerator{resourceGenerator: resourceGenerator, client: client}
	return &g, nil
}

func (g *HelmGenerator) Generate(namespace string, name string, parameters types.Unstructurable) ([]client.Object, error) {
	parameterMap := parameters.ToUnstructured()

	if err := g.fillDomain(parameterMap); err != nil {
		return nil, err
	}

	if err := g.fillDNSTarget(parameterMap); err != nil {
		return nil, err
	}

	return g.resourceGenerator.Generate(namespace, name, &unstructurableMap{data: parameterMap})
}

func trimDNSTarget(dnsTarget string) string {
	// Trim dnsTarget to under 64 chars
	for len(dnsTarget) > 64 {
		dnsTarget = dnsTarget[strings.Index(dnsTarget, ".")+1:]
	}
	// Fix for domain gw/creds secret name (Replace *.domain with x.domain for secret name)
	return strings.ReplaceAll(dnsTarget, "*", "x")
}

func (g *HelmGenerator) fillDNSTarget(parameters map[string]any) error {
	// get DNSTarget
	subscriptionServer := parameters["subscriptionServer"].(map[string]interface{})
	if subscriptionServer["dnsTarget"] != nil { // already filled in CRO
		subscriptionServer["dnsTarget"] = trimDNSTarget(subscriptionServer["dnsTarget"].(string))
		return nil
	}

	// DNSTarget not given - read it from the load balancer service in istio namespace
	if parameters["ingressGatewayLabels"] == nil {
		return fmt.Errorf("cannot get dnsTarget; provide either dnsTarget or ingressGatewayLabels in the CRO")
	}

	ingressGatewayLabels := parameters["ingressGatewayLabels"].(map[string]interface{})
	if ingressGatewayLabels["app"] == nil || ingressGatewayLabels["istio"] == nil {
		return fmt.Errorf("cannot get dnsTarget; provide ingressGatewayLabels/app and ingressGatewayLabels/istio values in the CRO")
	}

	dnsTarget, err := g.getDNSTarget(&v1alpha1.IngressGatewayLabels{App: ingressGatewayLabels["app"].(string), Istio: ingressGatewayLabels["istio"].(string)})
	if err != nil {
		return err
	}

	subscriptionServer["dnsTarget"] = trimDNSTarget(dnsTarget)

	return nil
}

func (g *HelmGenerator) getDNSTarget(ingressGatewayLabels *v1alpha1.IngressGatewayLabels) (dnsTarget string, err error) {

	ctx := context.TODO()

	// create ingress gateway selector from labels
	ingressLabelSelector, err := labels.ValidatedSelectorFromSet(getIngressGatewayLabels(ingressGatewayLabels))
	if err != nil {
		return "", err
	}

	// Get relevant Ingress Gateway pods
	ingressPods := &corev1.PodList{TypeMeta: metav1.TypeMeta{Kind: "Pod"}}
	err = g.client.List(ctx, ingressPods, &client.ListOptions{Namespace: metav1.NamespaceAll, LabelSelector: ingressLabelSelector})
	if err != nil {
		return "", err
	}

	// Determine relevant istio-ingressgateway namespace
	namespace := ""
	// Create a dummy lookup map for determining relevant pods
	relevantsPodsNames := map[string]struct{}{}
	for _, pod := range ingressPods.Items {
		// We only support 1 ingress gateway pod namespace as of now! (Multiple pods e.g. replicas can exist in the same namespace)
		if namespace == "" {
			namespace = pod.Namespace
		} else if namespace != pod.Namespace {
			return "", fmt.Errorf("more than one matching ingress gateway pod namespaces found")
		}
		relevantsPodsNames[pod.Name] = struct{}{}
	}
	if namespace == "" {
		return "", fmt.Errorf("no matching ingress gateway pod found")
	}

	// Get dnsTarget
	ingressGWSvc, err := g.getIngressGatewayService(ctx, relevantsPodsNames)
	if err != nil {
		return "", err
	}
	if ingressGWSvc != nil {
		dnsTarget = ingressGWSvc.Annotations[annotationDNSNames]
	}

	// No DNS Target --> Error
	if dnsTarget == "" {
		return "", fmt.Errorf("ingress gateway service not annotated with dns target name")
	}

	// Return ingress Gateway info (Namespace and DNS target)
	return dnsTarget, nil
}

func getIngressGatewayLabels(ingressGatewayLabels *v1alpha1.IngressGatewayLabels) map[string]string {
	ingressLabels := map[string]string{}

	ingressLabels["app"] = ingressGatewayLabels.App
	ingressLabels["istio"] = ingressGatewayLabels.Istio

	return ingressLabels
}

func (g *HelmGenerator) getLoadBalancerSvcs(ctx context.Context) ([]corev1.Service, error) {
	// List all services in the same namespace as the istio-ingressgateway pod namespace
	svcList := &corev1.ServiceList{TypeMeta: metav1.TypeMeta{Kind: "Service"}}
	if err := g.client.List(ctx, svcList, &client.ListOptions{Namespace: istioIngressGWNamespace}); err != nil {
		return nil, err
	}

	// Filter out LoadBalancer services
	loadBalancerSvcs := []corev1.Service{}
	for _, svc := range svcList.Items {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			loadBalancerSvcs = append(loadBalancerSvcs, svc)
		}
	}
	return loadBalancerSvcs, nil
}

func (g *HelmGenerator) getIngressGatewayService(ctx context.Context, relevantPodNames map[string]struct{}) (*corev1.Service, error) {
	loadBalancerSvcs, err := g.getLoadBalancerSvcs(ctx)
	if err != nil {
		return nil, err
	}
	// Get Relevant services that match the ingress gw pod via selectors
	var ingressGwSvc corev1.Service
	podList := &corev1.PodList{TypeMeta: metav1.TypeMeta{Kind: "Pod"}}
	for _, svc := range loadBalancerSvcs {
		// Get all matching ingress GW pods in the ingress gw namespace via ingress gw service selectors
		err := g.client.List(ctx, podList, &client.ListOptions{LabelSelector: labels.SelectorFromValidatedSet(svc.Spec.Selector)})
		if err != nil {
			return nil, err
		}
		for _, pod := range podList.Items {
			if _, ok := relevantPodNames[pod.Name]; ok {
				if ingressGwSvc.Name == "" {
					// we only expect 1 ingress gateway service in the cluster
					ingressGwSvc = svc
					break
				} else if ingressGwSvc.Name != svc.Name {
					return nil, fmt.Errorf("more than one matching ingress gateway service found")
				}
			}
		}
	}

	if ingressGwSvc.Name == "" {
		return nil, fmt.Errorf("unable to find a matching ingress gateway service")
	}
	return &ingressGwSvc, nil
}

func (g *HelmGenerator) fillDomain(parameters map[string]any) error {
	// get domain
	subscriptionServer := parameters["subscriptionServer"].(map[string]interface{})
	domain, err := g.getDomain(subscriptionServer["subDomain"].(string))
	if err != nil {
		return err
	}

	subscriptionServer["domain"] = domain
	delete(subscriptionServer, "subDomain")

	return nil
}

func (g *HelmGenerator) getDomain(subDomain string) (string, error) {
	configMapObj := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        shootInfoConfigMap,
			Namespace:   kubeSystemNamespace,
			Labels:      nil,
			Annotations: nil,
		},
	}

	ctx := context.TODO()
	err := g.client.Get(ctx, apitypes.NamespacedName{Namespace: configMapObj.GetNamespace(), Name: configMapObj.GetName()}, configMapObj)
	if err != nil {
		return "", err
	}

	return subDomain + "." + configMapObj.Data["domain"], nil
}
