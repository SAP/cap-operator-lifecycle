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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
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
	InitFlags(flag.CommandLine)

	if err := operator.ValidateFlags(); err != nil {
		t.Error("validate flag failed")
		return
	}

	if name := operator.GetName(); name != "cap-operator.sme.sap.com" {
		t.Error("invalid name")
		return
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: append(operator.GetUncacheableTypes(), &apiextensionsv1.CustomResourceDefinition{}, &apiregistrationv1.APIService{}),
			},
		},
		LeaderElection:                enableLeaderElection,
		LeaderElectionID:              operator.GetName(),
		LeaderElectionReleaseOnCancel: true,
	})

	if err != nil {
		t.Error("unable to start manager")
		return
	}

	fakeclientset := fakeclientset.NewSimpleClientset()
	fakeDiscovery, ok := fakeclientset.Discovery().(*fakediscovery.FakeDiscovery)
	if !ok {
		t.Error("fake discovery client creation failed")
	}

	if err := operator.Setup(mgr, fakeDiscovery); err != nil {
		t.Error("error registering controller with manager")
		return
	}
}
