/*
SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package main

// +kubebuilder:rbac:groups=operator.sme.sap.com,resources=capoperators;capoperators/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert.gardener.cloud,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways;virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dns.gardener.cloud,resources=dnsentries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations;mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups="sme.sap.com",resources=capapplications;capapplicationversions;captenants;captenantoperations;captenantoutputs;clusterdomains;domains,verbs=get;list;watch
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets;configmaps;pods;services;serviceaccounts;namespaces;events,verbs=get;list;watch;create;update;patch;delete

import (
	"flag"
	"os"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/sap/cap-operator-lifecycle/pkg/operator"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(apiregistrationv1.AddToScheme(scheme))

	operator.InitScheme(scheme)
}

func main() {
	var metricsAddr, probeAddr, chartDir string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&chartDir, "manifest-directory", "./chart",
		"The directory containing the deployment manifests for the managed operator.")
	operator.InitFlags(flag.CommandLine)
	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	// Set klog to use the same logger as controller-runtime, so that logs from all libraries are consistent.
	// There seems to be an issue with leader election logs being emitted which uses klog. See: https://github.com/kubernetes-sigs/controller-runtime/issues/2656
	klog.SetLogger(logger)

	if chartDir == "" {
		setupLog.Error(nil, "command line flag missing or empty: --chart-directory")
		os.Exit(1)
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
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := operator.Setup(mgr); err != nil {
		setupLog.Error(err, "error registering controller with manager")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
