/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package operator

import (
	"flag"
	"os"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sap/component-operator-runtime/pkg/component"
	"github.com/sap/component-operator-runtime/pkg/manifests"
	"github.com/sap/component-operator-runtime/pkg/operator"

	operatorv1alpha1 "github.com/sap/cap-operator-lifecycle/api/v1alpha1"
	"github.com/sap/cap-operator-lifecycle/internal/transformer"
)

const Name = "cap-operator.sme.sap.com"
const FlagPrefix = "manifest-directory"

type Options struct {
	Name       string
	FlagPrefix string
}

type Operator struct {
	options Options
}

var defaultOperator operator.Operator = New()

func GetName() string {
	return defaultOperator.GetName()
}

func InitScheme(scheme *runtime.Scheme) {
	defaultOperator.InitScheme(scheme)
}

func InitFlags(flagset *flag.FlagSet) {
	defaultOperator.InitFlags(flagset)
}

func ValidateFlags() error {
	return defaultOperator.ValidateFlags()
}

func GetUncacheableTypes() []client.Object {
	return defaultOperator.GetUncacheableTypes()
}

func Setup(mgr ctrl.Manager, discoveryClient discovery.DiscoveryInterface) error {
	return defaultOperator.Setup(mgr, discoveryClient)
}

func New() *Operator {
	return NewWithOptions(Options{})
}

func NewWithOptions(options Options) *Operator {
	operator := &Operator{options: options}
	if operator.options.Name == "" {
		operator.options.Name = Name
	}
	if operator.options.FlagPrefix == "" {
		operator.options.FlagPrefix = FlagPrefix
	}
	return operator
}

func (o *Operator) GetName() string {
	return o.options.Name
}

func (o *Operator) InitScheme(scheme *runtime.Scheme) {
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
}

var chartFlag *flag.Flag

func (o *Operator) InitFlags(flagset *flag.FlagSet) {
	chartFlag = flagset.Lookup(o.options.FlagPrefix)
}

func (o *Operator) ValidateFlags() error {
	// Add logic to validate flags (if running in a combined controller you might want to evaluate o.options.FlagPrefix).
	return nil
}

func (o *Operator) GetUncacheableTypes() []client.Object {
	// Add types which should bypass informer caching.
	return []client.Object{&operatorv1alpha1.CAPOperator{}}
}

func (o *Operator) Setup(mgr ctrl.Manager, discoveryClient discovery.DiscoveryInterface) error {
	chartDir := chartFlag.Value.String()

	if err := checkDirectoryExists(chartDir); err != nil {
		return errors.Wrap(err, "error checking manifest directory")
	}

	resourceGenerator, err := manifests.NewHelmGeneratorWithParameterTransformer(
		o.options.Name,
		nil,
		chartDir,
		mgr.GetClient(),
		discoveryClient,
		transformer.NewParameterTransformer(mgr.GetClient()),
	)
	if err != nil {
		return errors.Wrap(err, "error initializing resource generator")
	}

	if err := component.NewReconciler[*operatorv1alpha1.CAPOperator](
		o.options.Name,
		mgr.GetClient(),
		discoveryClient,
		mgr.GetEventRecorderFor(o.options.Name),
		mgr.GetScheme(),
		resourceGenerator,
	).SetupWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create controller")
	}

	return nil
}

func checkDirectoryExists(path string) error {
	fsinfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fsinfo.IsDir() {
		return errors.Errorf("not a directory: %s", path)
	}
	return nil
}
