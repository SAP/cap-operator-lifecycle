/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"fmt"
	"os"

	"github.com/sap/component-operator-runtime/pkg/component"
	runtimetypes "github.com/sap/component-operator-runtime/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=".status.state"

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
	// SubscriptionServer info
	SubscriptionServer SubscriptionServer `json:"subscriptionServer"`
	// IngressGatewayLabels
	IngressGatewayLabels IngressGatewayLabels `json:"ingressGatewayLabels"`
}

type SubscriptionServer struct {
	DNSTarget string `json:"dnsTarget,omitempty"`
	Subdomain string `json:"subDomain"`
}

type IngressGatewayLabels struct {
	Istio string `json:"istio"`
	App   string `json:"app"`
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
