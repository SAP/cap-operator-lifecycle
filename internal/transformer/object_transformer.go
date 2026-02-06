/*
SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package transformer

import (
	"context"
	"fmt"
	"strings"

	operatorv1alpha1 "github.com/sap/cap-operator-lifecycle/api/v1alpha1"
	"github.com/sap/component-operator-runtime/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Annotation to check for retaining resources.
	AnnotationRetainResources = "operator.sme.sap.com/retain-resources"
	// Annotation key for setting the delete policy.
	AnnotationDeletePolicySuffix = "/delete-policy"
)

type objectTransformer struct {
	client client.Client
	name   string
}

func NewObjectTransformer(client client.Client, name string) *objectTransformer {
	return &objectTransformer{client: client, name: name}
}

func (ot *objectTransformer) TransformObjects(_, _ string, objects []client.Object) ([]client.Object, error) {
	// Step 1: Find the single CAPOperator resource.
	capOperator, err := ot.getCAPOperator()
	if err != nil {
		return objects, err
	}

	// Step 2: Check for the retain-resources annotation value.
	shouldRetain := ot.checkRetainResources(capOperator)

	// Step 3: Apply transformation logic based on the check.
	if shouldRetain {
		// If retain-resources="true", add delete-policy=orphan to all resources.
		return ot.addDeletePolicy(objects), nil
	}

	// If retain-resources is not "true", ensure the delete-policy annotation is removed from all resources.
	return ot.removeDeletePolicy(objects), nil
}

// getCAPOperator fetches the single CAPOperator instance in the cluster.
func (ot *objectTransformer) getCAPOperator() (*operatorv1alpha1.CAPOperator, error) {
	capOperatorList := &operatorv1alpha1.CAPOperatorList{}

	// List all CAPOperator resources across all namespaces.
	err := ot.client.List(context.TODO(), capOperatorList, &client.ListOptions{Namespace: corev1.NamespaceAll})
	if err != nil {
		return nil, fmt.Errorf("failed to list CAPOperator resources: %w", err)
	}

	if len(capOperatorList.Items) == 0 {
		return nil, fmt.Errorf("no CAPOperator resource found")
	}

	if len(capOperatorList.Items) > 1 {
		return nil, fmt.Errorf("more than one CAPOperator resource found")
	}

	return &capOperatorList.Items[0], nil
}

// checkRetainResources checks if the CAPOperator is annotated to retain resources.
func (ot *objectTransformer) checkRetainResources(capOperator *operatorv1alpha1.CAPOperator) bool {
	retainResourcesValue, ok := capOperator.Annotations[AnnotationRetainResources]
	return ok && strings.ToLower(retainResourcesValue) == "true"
}

// addDeletePolicy iterates over objects and adds orphan delete policy annotation to resources.
func (ot *objectTransformer) addDeletePolicy(objects []client.Object) []client.Object {
	for _, obj := range objects {
		annotations := obj.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}

		// Set the delete-policy annotation to orphan
		annotations[ot.name+AnnotationDeletePolicySuffix] = string(reconciler.DeletePolicyOrphan)
		obj.SetAnnotations(annotations)
	}
	return objects
}

func (ot *objectTransformer) removeDeletePolicy(objects []client.Object) []client.Object {
	for _, obj := range objects {
		annotations := obj.GetAnnotations()
		if annotations != nil {
			if _, found := annotations[ot.name+AnnotationDeletePolicySuffix]; found {
				delete(annotations, ot.name+AnnotationDeletePolicySuffix)
				obj.SetAnnotations(annotations)
			}
		}
	}
	return objects
}
