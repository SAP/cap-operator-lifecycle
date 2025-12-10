package transformer

import (
	"testing"

	operatorv1alpha1 "github.com/sap/cap-operator-lifecycle/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// mockObject is a mock implementation of client.Object for testing
type mockObject struct {
	metav1.ObjectMeta
	metav1.TypeMeta
}

func (m *mockObject) DeepCopyObject() runtime.Object {
	return &mockObject{
		ObjectMeta: *m.ObjectMeta.DeepCopy(),
		TypeMeta:   m.TypeMeta,
	}
}

func (m *mockObject) GetObjectKind() schema.ObjectKind {
	return m
}

func (m *mockObject) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(m.APIVersion, m.Kind)
}

func (m *mockObject) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	m.APIVersion, m.Kind = gvk.ToAPIVersionAndKind()
}

func TestNewObjectTransformer(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	transformer := NewObjectTransformer(fakeClient)

	if transformer == nil {
		t.Fatal("transformer is nil")
	}
	if transformer.client == nil {
		t.Fatal("transformer.client is nil")
	}
}

func TestGetCAPOperator(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = operatorv1alpha1.AddToScheme(scheme)

	tests := []struct {
		name          string
		capOperators  []operatorv1alpha1.CAPOperator
		expectedError string
	}{
		{
			name:          "No CAPOperator found",
			capOperators:  []operatorv1alpha1.CAPOperator{},
			expectedError: "no CAPOperator resource found",
		},
		{
			name: "Single CAPOperator found",
			capOperators: []operatorv1alpha1.CAPOperator{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-operator",
						Namespace: "default",
					},
				},
			},
			expectedError: "",
		},
		{
			name: "Multiple CAPOperators found",
			capOperators: []operatorv1alpha1.CAPOperator{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-operator-1",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-operator-2",
						Namespace: "default",
					},
				},
			},
			expectedError: "more than one CAPOperator resource found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)
			for i := range tt.capOperators {
				clientBuilder = clientBuilder.WithObjects(&tt.capOperators[i])
			}
			fakeClient := clientBuilder.Build()

			transformer := NewObjectTransformer(fakeClient)
			capOperator, err := transformer.getCAPOperator()

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedError)
				}
				if capOperator != nil {
					t.Fatal("expected capOperator to be nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if capOperator == nil {
					t.Fatal("expected capOperator to be non-nil")
				}
			}
		})
	}
}

func TestCheckRetainCRDs(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
	}{
		{
			name:        "No annotations",
			annotations: nil,
			expected:    false,
		},
		{
			name:        "Annotation not present",
			annotations: map[string]string{"other-key": "value"},
			expected:    false,
		},
		{
			name:        "Annotation set to true",
			annotations: map[string]string{AnnotationRetainCRDs: "true"},
			expected:    true,
		},
		{
			name:        "Annotation set to True (case insensitive)",
			annotations: map[string]string{AnnotationRetainCRDs: "True"},
			expected:    true,
		},
		{
			name:        "Annotation set to TRUE (case insensitive)",
			annotations: map[string]string{AnnotationRetainCRDs: "TRUE"},
			expected:    true,
		},
		{
			name:        "Annotation set to false",
			annotations: map[string]string{AnnotationRetainCRDs: "false"},
			expected:    false,
		},
		{
			name:        "Annotation set to invalid value",
			annotations: map[string]string{AnnotationRetainCRDs: "invalid"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capOperator := &operatorv1alpha1.CAPOperator{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			scheme := runtime.NewScheme()
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			transformer := NewObjectTransformer(fakeClient)

			result := transformer.checkRetainCRDs(capOperator)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAddDeletePolicyOrphan(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	transformer := NewObjectTransformer(fakeClient)

	crdObject := &mockObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-crd",
		},
	}

	nonCRDObject := &mockObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-configmap",
		},
	}

	tests := []struct {
		name     string
		objects  []client.Object
		expected map[string]string
	}{
		{
			name:    "Add annotation to CRD without existing annotations",
			objects: []client.Object{crdObject},
			expected: map[string]string{
				AnnotationDeletePolicy: "orphan",
			},
		},
		{
			name: "Add annotation to CRD with existing annotations",
			objects: []client.Object{
				&mockObject{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test-crd-with-annotations",
						Annotations: map[string]string{"existing": "value"},
					},
				},
			},
			expected: map[string]string{
				"existing":             "value",
				AnnotationDeletePolicy: "orphan",
			},
		},
		{
			name:     "Non-CRD object should not be modified",
			objects:  []client.Object{nonCRDObject},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.addDeletePolicyOrphan(tt.objects)
			if len(result) != len(tt.objects) {
				t.Fatalf("expected %d objects, got %d", len(tt.objects), len(result))
			}

			if tt.expected != nil {
				annotations := result[0].GetAnnotations()
				if len(annotations) != len(tt.expected) {
					t.Fatalf("expected %d annotations, got %d", len(tt.expected), len(annotations))
				}
				for k, v := range tt.expected {
					if annotations[k] != v {
						t.Errorf("expected annotation %s=%s, got %s", k, v, annotations[k])
					}
				}
			}
		})
	}
}

func TestRemoveDeletePolicy(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	transformer := NewObjectTransformer(fakeClient)

	tests := []struct {
		name     string
		objects  []client.Object
		expected map[string]string
	}{
		{
			name: "Remove delete-policy annotation from CRD",
			objects: []client.Object{
				&mockObject{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-crd",
						Annotations: map[string]string{
							AnnotationDeletePolicy: "orphan",
						},
					},
				},
			},
			expected: map[string]string{},
		},
		{
			name: "Remove delete-policy but keep other annotations",
			objects: []client.Object{
				&mockObject{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-crd",
						Annotations: map[string]string{
							AnnotationDeletePolicy: "orphan",
							"other-annotation":     "value",
						},
					},
				},
			},
			expected: map[string]string{
				"other-annotation": "value",
			},
		},
		{
			name: "CRD without delete-policy annotation",
			objects: []client.Object{
				&mockObject{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test-crd",
						Annotations: map[string]string{"other": "value"},
					},
				},
			},
			expected: map[string]string{"other": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.removeDeletePolicy(tt.objects)
			if len(result) != len(tt.objects) {
				t.Fatalf("expected %d objects, got %d", len(tt.objects), len(result))
			}

			annotations := result[0].GetAnnotations()
			if len(annotations) != len(tt.expected) {
				t.Fatalf("expected %d annotations, got %d", len(tt.expected), len(annotations))
			}
			for k, v := range tt.expected {
				if annotations[k] != v {
					t.Errorf("expected annotation %s=%s, got %s", k, v, annotations[k])
				}
			}
		})
	}
}

func TestTransformObjects(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = operatorv1alpha1.AddToScheme(scheme)

	tests := []struct {
		name              string
		capOperator       *operatorv1alpha1.CAPOperator
		objects           []client.Object
		expectError       bool
		expectedAnnotated bool
	}{
		{
			name: "Retain CRDs annotation is true",
			capOperator: &operatorv1alpha1.CAPOperator{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-operator",
					Namespace: "default",
					Annotations: map[string]string{
						AnnotationRetainCRDs: "true",
					},
				},
			},
			objects: []client.Object{
				&mockObject{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-crd",
					},
				},
			},
			expectError:       false,
			expectedAnnotated: true,
		},
		{
			name: "Retain CRDs annotation is false",
			capOperator: &operatorv1alpha1.CAPOperator{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-operator",
					Namespace: "default",
					Annotations: map[string]string{
						AnnotationRetainCRDs: "false",
					},
				},
			},
			objects: []client.Object{
				&mockObject{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-crd",
						Annotations: map[string]string{
							AnnotationDeletePolicy: "orphan",
						},
					},
				},
			},
			expectError:       false,
			expectedAnnotated: false,
		},
		{
			name:        "No CAPOperator found",
			capOperator: nil,
			objects: []client.Object{
				&mockObject{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-crd",
					},
				},
			},
			expectError:       true,
			expectedAnnotated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)
			if tt.capOperator != nil {
				clientBuilder = clientBuilder.WithObjects(tt.capOperator)
			}
			fakeClient := clientBuilder.Build()

			transformer := NewObjectTransformer(fakeClient)
			result, err := transformer.TransformObjects("default", "test", tt.objects)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result to be non-nil")
				}

				if tt.expectedAnnotated {
					annotations := result[0].GetAnnotations()
					if annotations[AnnotationDeletePolicy] != "orphan" {
						t.Errorf("expected annotation %s=orphan, got %s", AnnotationDeletePolicy, annotations[AnnotationDeletePolicy])
					}
				} else {
					annotations := result[0].GetAnnotations()
					if _, exists := annotations[AnnotationDeletePolicy]; exists {
						t.Error("expected annotation to not exist")
					}
				}
			}
		})
	}
}
