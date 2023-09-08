/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/
package operator

import (
	"flag"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

var (
	scheme = runtime.NewScheme()
)

func TestCheckDirectoryExists(t *testing.T) {

	// invalid directory
	if err := checkDirectoryExists("invalid"); err == nil {
		t.Error("error expected but not returned")
		return
	}

	// valid directory
	if err := checkDirectoryExists("../../chart"); err != nil {
		t.Error("error not expected but returned")
		return
	}

	// File path passed instead of a directory
	if err := checkDirectoryExists("../../chart/values.yaml"); err == nil {
		t.Error("error expected but not returned")
		return
	}
}

func TestOperator(t *testing.T) {
	{
		operator := New()

		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
		utilruntime.Must(apiregistrationv1.AddToScheme(scheme))

		operator.InitScheme(scheme)

		opts := zap.Options{
			Development: false,
		}
		opts.BindFlags(flag.CommandLine)
		flag.Parse()

		ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

		var chartDir string
		var enableLeaderElection bool

		flag.BoolVar(&enableLeaderElection, "leader-elect", false,
			"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
		flag.StringVar(&chartDir, "manifest-directory", "../../chart",
			"The directory containing the deployment manifests for the managed operator.")

		operator.InitFlags(flag.CommandLine)

		if err := operator.ValidateFlags(); err != nil {
			t.Error("Flag validation failed")
			return
		}

		if name := operator.GetName(); name != "cap-operator.sme.sap.com" {
			t.Error("Invalid name")
			return
		}

		if uncacheTypes := operator.GetUncacheableTypes(); len(uncacheTypes) != 1 {
			t.Error("Returned wrong number of types")
			return
		}

		// TODO - Add unit tests for setup
	}
}
