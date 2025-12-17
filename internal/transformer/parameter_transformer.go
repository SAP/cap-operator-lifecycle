/*
SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package transformer

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	apitypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	componentoperatorruntimetypes "github.com/sap/component-operator-runtime/pkg/types"
)

const (
	kubeSystemNamespace     = "kube-system"
	shootInfoConfigMap      = "shoot-info"
	istioIngressGWNamespace = "istio-system"
	annotationDNSNames      = "dns.gardener.cloud/dnsnames"
)

var setupLog = ctrl.Log.WithName("transformer")

type parameterTransformer struct {
	client client.Client
}

func NewParameterTransformer(client client.Client) *parameterTransformer {
	return &parameterTransformer{client: client}
}

func (t *parameterTransformer) TransformParameters(_, _ string, parameters componentoperatorruntimetypes.Unstructurable) (componentoperatorruntimetypes.Unstructurable, error) {
	parameterMap := parameters.ToUnstructured()

	if err := t.fillDomain(parameterMap); err != nil {
		return nil, err
	}

	if err := t.fillDNSTarget(parameterMap); err != nil {
		return nil, err
	}

	return componentoperatorruntimetypes.UnstructurableMap(parameterMap), nil
}

func replaceAsteriskDNSTarget(dnsTarget string) string {
	// Fix for domain gw/creds secret name (Replace *.domain with x.domain for secret name)
	return strings.ReplaceAll(dnsTarget, "*", "x")
}

func (t *parameterTransformer) fillDNSTarget(parameters map[string]any) error {
	subscriptionServer := parameters["subscriptionServer"].(map[string]interface{})

	if parameters["controller"] == nil {
		parameters["controller"] = map[string]interface{}{}
	}
	controller := parameters["controller"].(map[string]interface{})

	// DNSTarget given - use it
	if parameters["dnsTarget"] != nil {
		replacedDnsTarget := replaceAsteriskDNSTarget(parameters["dnsTarget"].(string))

		// set the dnsTarget in the subscriptionServer
		subscriptionServer["dnsTarget"] = replacedDnsTarget
		// set the dnsTarget in the controller
		controller["dnsTarget"] = replacedDnsTarget

		delete(parameters, "dnsTarget")
		return nil
	}

	// DNSTarget not given - read it from the load balancer service in istio namespace
	if parameters["ingressGatewayLabels"] == nil {
		return fmt.Errorf("unable to retrieve dnsTarget; please specify either dnsTarget or ingressGatewayLabels in the CAP Operator CRO")
	}

	dnsTarget, err := t.getDNSTargetUsingIngressGatewayLabels(parameters["ingressGatewayLabels"].([]interface{}))
	if err != nil {
		setupLog.Info("dnsTarget not found using ingressGatewayLabels", "error", err)

		// default the dnsTarget to the x.<cluster-domain>
		dnsTarget, err = t.getDomain("x")
		if err != nil {
			return err
		}

		setupLog.Info("defaulting dnsTarget to " + dnsTarget)
	}

	replacedDnsTarget := replaceAsteriskDNSTarget(dnsTarget)
	// set the dnsTarget in the subscriptionServer
	subscriptionServer["dnsTarget"] = replacedDnsTarget
	// set the dnsTarget in the controller
	controller["dnsTarget"] = replacedDnsTarget
	delete(parameters, "ingressGatewayLabels")
	return nil
}

func (t *parameterTransformer) getDNSTargetUsingIngressGatewayLabels(ingressGatewayLabels []interface{}) (dnsTarget string, err error) {

	ctx := context.TODO()

	// create ingress gateway selector from labels
	ingressLabelSelector, err := labels.ValidatedSelectorFromSet(convertIngressGatewayLabelsToMap(ingressGatewayLabels))
	if err != nil {
		return "", err
	}

	// Get relevant Ingress Gateway pods
	ingressPods := &corev1.PodList{TypeMeta: metav1.TypeMeta{Kind: "Pod"}}
	err = t.client.List(ctx, ingressPods, &client.ListOptions{Namespace: metav1.NamespaceAll, LabelSelector: ingressLabelSelector})
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
	ingressGWSvc, err := t.getIngressGatewayService(ctx, relevantsPodsNames)
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

	return dnsTarget, nil
}

func convertIngressGatewayLabelsToMap(ingressGatewayLabels []interface{}) map[string]string {
	ingressLabels := map[string]string{}

	for _, label := range ingressGatewayLabels {
		labelMap := label.(map[string]interface{})
		ingressLabels[labelMap["name"].(string)] = labelMap["value"].(string)
	}

	return ingressLabels
}

func (t *parameterTransformer) getLoadBalancerServices(ctx context.Context) ([]corev1.Service, error) {
	// List all services in the same namespace as the istio-ingressgateway pod namespace
	svcList := &corev1.ServiceList{TypeMeta: metav1.TypeMeta{Kind: "Service"}}
	if err := t.client.List(ctx, svcList, &client.ListOptions{Namespace: istioIngressGWNamespace}); err != nil {
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

func (t *parameterTransformer) getIngressGatewayService(ctx context.Context, relevantPodNames map[string]struct{}) (*corev1.Service, error) {
	loadBalancerSvcs, err := t.getLoadBalancerServices(ctx)
	if err != nil {
		return nil, err
	}
	// Get Relevant services that match the ingress gw pod via selectors
	var ingressGwSvc corev1.Service
	podList := &corev1.PodList{TypeMeta: metav1.TypeMeta{Kind: "Pod"}}
	for _, svc := range loadBalancerSvcs {
		// Get all matching ingress GW pods in the ingress gw namespace via ingress gw service selectors
		err := t.client.List(ctx, podList, &client.ListOptions{LabelSelector: labels.SelectorFromValidatedSet(svc.Spec.Selector)})
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

func (t *parameterTransformer) fillDomain(parameters map[string]any) error {
	// get domain
	subscriptionServer := parameters["subscriptionServer"].(map[string]interface{})
	domain, err := t.getDomain(subscriptionServer["subDomain"].(string))
	if err != nil {
		return err
	}

	if len(domain) > 64 {
		return fmt.Errorf("subscription server domain '%s' is longer than 64 characters; use a smaller subDomain", domain)
	}

	subscriptionServer["domain"] = domain
	delete(subscriptionServer, "subDomain")

	return nil
}

func (t *parameterTransformer) getDomain(subDomain string) (string, error) {
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
	err := t.client.Get(ctx, apitypes.NamespacedName{Namespace: configMapObj.GetNamespace(), Name: configMapObj.GetName()}, configMapObj)
	if err != nil {
		return "", err
	}

	return subDomain + "." + configMapObj.Data["domain"], nil
}
