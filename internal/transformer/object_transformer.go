package transformer

import (
	"context"
	"fmt"
	"strings"

	operatorv1alpha1 "github.com/sap/cap-operator-lifecycle/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Annotation to check for retaining CRDs.
	AnnotationRetainCRDs = "cap-operator.sme.sap.com/retain-crds"
	// Annotation key for setting the delete policy.
	AnnotationDeletePolicy = "cap-operator.sme.sap.com/delete-policy"

	// GVK for CustomResourceDefinition
	GVK_CRD_Group   = "apiextensions.k8s.io"
	GVK_CRD_Version = "v1"
	GVK_CRD_Kind    = "CustomResourceDefinition"
)

type objectTransformer struct {
	client client.Client
}

func NewObjectTransformer(client client.Client) *objectTransformer {
	return &objectTransformer{client: client}
}

func (ot *objectTransformer) TransformObjects(namespace string, name string, objects []client.Object) ([]client.Object, error) {
	// Step 1: Find the single CAPOperator resource.
	capOperator, err := ot.getCAPOperator()
	if err != nil {
		return objects, err
	}

	// Step 2: Check for the retain-crds annotation value.
	shouldRetain := ot.checkRetainCRDs(capOperator)

	// Step 3: Apply transformation logic based on the check.
	if shouldRetain {
		// If retain-crds="true", add delete-policy=orphan to all CRDs.
		return ot.addDeletePolicyOrphan(objects), nil
	}

	// If retain-crds is not "true", ensure the delete-policy annotation is removed from all CRDs.
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

// checkRetainCRDs checks if the CAPOperator is annotated to retain CRDs.
func (ot *objectTransformer) checkRetainCRDs(capOperator *operatorv1alpha1.CAPOperator) bool {
	retainCRDsValue, ok := capOperator.Annotations[AnnotationRetainCRDs]
	return ok && strings.ToLower(retainCRDsValue) == "true"
}

// addDeletePolicyOrphan iterates over objects and adds the orphan delete policy annotation to CRDs.
func (ot *objectTransformer) addDeletePolicyOrphan(objects []client.Object) []client.Object {
	crdGVK := schema.GroupVersionKind{Group: GVK_CRD_Group, Version: GVK_CRD_Version, Kind: GVK_CRD_Kind}

	for _, obj := range objects {
		if obj.GetObjectKind().GroupVersionKind() == crdGVK {
			annotations := obj.GetAnnotations()
			if annotations == nil {
				annotations = make(map[string]string)
			}

			// Add or overwrite the delete-policy annotation.
			annotations[AnnotationDeletePolicy] = "orphan"
			obj.SetAnnotations(annotations)
		}
	}
	return objects
}

// removeDeletePolicy iterates over objects and ensures the delete policy annotation is absent from CRDs.
func (ot *objectTransformer) removeDeletePolicy(objects []client.Object) []client.Object {
	crdGVK := schema.GroupVersionKind{Group: GVK_CRD_Group, Version: GVK_CRD_Version, Kind: GVK_CRD_Kind}

	for _, obj := range objects {
		if obj.GetObjectKind().GroupVersionKind() == crdGVK {
			annotations := obj.GetAnnotations()
			if annotations != nil {
				if _, found := annotations[AnnotationDeletePolicy]; found {
					delete(annotations, AnnotationDeletePolicy)
					// Only call SetAnnotations if an annotation was actually deleted
					obj.SetAnnotations(annotations)
				}
			}
		}
	}
	return objects
}
