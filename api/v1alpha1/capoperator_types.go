/*
SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"fmt"
	"os"

	"github.com/sap/component-operator-runtime/pkg/component"
	runtimetypes "github.com/sap/component-operator-runtime/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=".status.state"

// CAPOperator is the Schema for the CAPOperators API
type CAPOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPOperatorSpec   `json:"spec,omitempty"`
	Status CAPOperatorStatus `json:"status,omitempty"`
}

type CAPOperatorStatus struct {
	// add other fields to status subresource here
	component.Status `json:",inline"`
}

// CAPOperatorSpec defines the desired state of CAPOperator
type CAPOperatorSpec struct {
	// SubscriptionServer specification
	SubscriptionServer SubscriptionServer `json:"subscriptionServer"`
	// +kubebuilder:validation:Pattern=^[a-z0-9-.]*$
	// Public ingress URL for the cluster Load Balancer
	DNSTarget string `json:"dnsTarget,omitempty"`
	// +kubebuilder:validation:MinItems=1
	// Labels used to identify the istio ingress-gateway component and its corresponding namespace. Usually {"app":"istio-ingressgateway","istio":"ingressgateway"}
	IngressGatewayLabels []NameValue `json:"ingressGatewayLabels,omitempty"`
	// Controller specification
	Controller Controller `json:"controller,omitempty"`
	// Monitoring specification
	Monitoring Monitoring `json:"monitoring,omitempty"`
	// Webhook specification
	Webhook Webhook `json:"webhook,omitempty"`
}

type Webhook struct {
	// Certificate manager which can be either `Default` or `CertManager`
	// +kubebuilder:validation:Enum=Default;CertManager
	CertificateManager CertificateManager `json:"certificateManager,omitempty"`
	// Certificate configuration
	CertificateConfig *CertificateConfig `json:"certificateConfig,omitempty"`
}

type Monitoring struct {
	// Optionally enable Prometheus monitoring for all components
	Enabled bool `json:"enabled"`
}

type SubscriptionServer struct {
	Subdomain string `json:"subDomain"`
	// Certificate manager which can be either `Gardener` or `CertManager`
	// +kubebuilder:validation:Enum=Gardener;CertManager
	CertificateManager CertificateManager `json:"certificateManager,omitempty"`
	// Certificate configuration
	CertificateConfig *CertificateConfig `json:"certificateConfig,omitempty"`
}

type Controller struct {
	// Optionally enable detailed opertational metrics for the controller by setting this to true
	DetailedOperationalMetrics bool `json:"detailedOperationalMetrics,omitempty"`
	// Version monitoring configuration
	VersionMonitoring *VersionMonitoring `json:"versionMonitoring,omitempty"`
	// Optionally specify list of additional volumes for the controller pod(s)
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// Optionally specify list of additional volumeMounts for the controller container(s)
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

type VersionMonitoring struct {
	// URL of the Prometheus server from which metrics related to managed application versions can be queried
	PrometheusAddress string `json:"prometheusAddress,omitempty"`
	// The duration (example 2h) after which versions are evaluated for deletion; based on specified workload metrics
	MetricsEvaluationInterval Duration `json:"metricsEvaluationInterval,omitempty"`
	// The duration (example 10m) to wait before retrying to acquire Prometheus client and verify connection, after a failed attempt
	PromClientAcquireRetryDelay Duration `json:"promClientAcquireRetryDelay,omitempty"`
}

type CertificateConfig struct {
	// Gardener configuration
	Gardener Gardener `json:"gardener,omitempty"`
	// CertManager configuration
	CertManager CertManager `json:"certManager,omitempty"`
}

type Gardener struct {
	// Issuer name
	IssuerName string `json:"issuerName,omitempty"`
	// Issuer namespace
	IssuerNamespace string `json:"issuerNamespace,omitempty"`
}

type CertManager struct {
	// Issuer name
	IssuerName string `json:"issuerName,omitempty"`
	// Issuer kind
	IssuerKind string `json:"issuerKind,omitempty"`
	// Issuer group
	IssuerGroup string `json:"issuerGroup,omitempty"`
}

// Supported values are Gardener, CertManager, or Default
// For the subscription server, it can be either Gardener or CertManager, while for the webhook, it can be either Default or CertManager
type CertificateManager string

// Duration is a valid time duration that can be parsed by Prometheus
// Supported units: y, w, d, h, m, s, ms
// Examples: `30s`, `1m`, `1h20m15s`, `15d`
// +kubebuilder:validation:Pattern:="^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"
type Duration string

// Generic Name/Value configuration
type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var _ component.Component = &CAPOperator{}

// +kubebuilder:object:root=true

// CAPOperatorList contains a list of CAPOperator
type CAPOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPOperator `json:"items"`
}

func (c *CAPOperatorSpec) ToUnstructured() map[string]any {
	result, err := runtime.DefaultUnstructuredConverter.ToUnstructured(c)
	if err != nil {
		fmt.Println("Error calling ToUnstructured: ", err.Error())
		return nil
	}
	return result
}

func (c *CAPOperator) GetDeploymentNamespace() string {
	return os.Getenv("POD_NAMESPACE")
}

func (c *CAPOperator) GetDeploymentName() string {
	return c.Name
}

func (c *CAPOperator) GetSpec() runtimetypes.Unstructurable {
	return &c.Spec
}

func (c *CAPOperator) GetStatus() *component.Status {
	return &c.Status.Status
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&CAPOperator{}, &CAPOperatorList{})
}
