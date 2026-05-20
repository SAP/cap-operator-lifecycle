/*
SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package transformer

import (
	"strings"

	operatorv1alpha1 "github.com/sap/cap-operator-lifecycle/api/v1alpha1"
	"github.com/sap/cap-operator-lifecycle/internal/util"
	"github.com/sap/component-operator-runtime/pkg/reconciler"
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
	capOperator, err := util.GetCAPOperator(ot.client)
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
