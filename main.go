/*
SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package main

// +kubebuilder:rbac:groups=operator.sme.sap.com,resources=capoperators;capoperators/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert.gardener.cloud,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways;virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dns.gardener.cloud,resources=dnsentries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups="sme.sap.com",resources=capapplications;capapplicationversions;captenants;captenantoperations;captenantoutputs,verbs=get;list;watch
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets;configmaps;pods;services;serviceaccounts;namespaces;events,verbs=get;list;watch;create;update;patch;delete

import (
	"flag"
	"os"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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
	var metricsAddr string
	var probeAddr string
	// Uncomment the following lines to enable webhooks.
	// var webhookAddr string
	// var webhookCertDir string
	var enableLeaderElection bool
	var chartDir string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.")
	// Uncomment the following lines to enable webhooks.
	// flag.StringVar(&webhookAddr, "webhook-bind-address", ":2443",
	//	"The address the webhooks endpoint binds to.")
	// flag.StringVar(&webhookCertDir, "webhook-tls-directory", "",
	//	"The directory containing tls server key and certificate, as tls.key and tls.crt; defaults to $TMPDIR/k8s-webhook-server/serving-certs")
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

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if chartDir == "" {
		setupLog.Error(nil, "command line flag missing or empty: --chart-directory")
		os.Exit(1)
	}

	// Uncomment the following lines to enable webhooks.
	// webhookHost, webhookPort, err := parseAddress(webhookAddr)
	// if err != nil {
	//	setupLog.Error(err, "error parsing webhook address")
	//	os.Exit(1)
	// }

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
		// Uncomment the following lines to enable webhooks.
		// WebhookServer: webhook.NewServer(webhook.Options{
		// 	Host:    webhookHost,
		//	Port:    webhookPort,
		//	CertDir: webhookCertDir,
		// }),
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Uncomment to enable conversion webhook (in case additional api versions are added in ./api).
	// Note: to make conversion work, additional changes are necessary:
	// - additional api versions have to be added to InitScheme() in pkg/operator/operator.go
	// - one of the api versions has to marked as Hub, all other versions need to implement the
	//   conversion.Convertible interface (see https://book.kubebuilder.io/multiversion-tutorial/conversion.html)
	// - one of the api versions has to be marked as storage version (+kubebuilder:storageversion)
	// - the crd resource has to be enhanced with a conversion section, telling the Kubernetes API server how to
	//   connect to the conversion endpoint.
	// mgr.GetWebhookServer().Register("/convert", conversion.NewWebhookHandler(mgr.GetScheme()))

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

// Uncomment the following lines to enable webhooks.
// func parseAddress(address string) (string, int, error) {
//	host, p, err := net.SplitHostPort(address)
//	if err != nil {
//		return "", -1, err
//	}
//	port, err := strconv.Atoi(p)
//	if err != nil {
//		return "", -1, err
//	}
//	return host, port, nil
// }
