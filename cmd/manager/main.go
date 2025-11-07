/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	elasticv1alpha1 "github.com/Carrefour-Group/elastic-phenix-operator/pkg/api/v1alpha1"
	"github.com/Carrefour-Group/elastic-phenix-operator/pkg/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	MetricsAddrFlag           = "metrics-addr"
	EnableLeaderElectionFlag  = "enable-leader-election"
	NamespacesFlag            = "namespaces"
	NamespacesRegexFilterFlag = "namespaces-regex-filter"
)

func init() {
	viper.AutomaticEnv()

	_ = clientgoscheme.AddToScheme(scheme)

	_ = elasticv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.String(MetricsAddrFlag, ":8080", "The address the metric endpoint binds to.")
	pflag.Bool(EnableLeaderElectionFlag, false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	pflag.StringSlice(NamespacesFlag, []string{}, "Comma-separated list of namespaces in which "+
		"this operator should manage resources (defaults to all namespaces)")
	pflag.String(NamespacesRegexFilterFlag, "", "Filter that will be applied before reconciliation "+
		"on namespaces managed by namespaces flag (defaults to no filter applied)")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	var metricsAddr = viper.GetString(MetricsAddrFlag)
	var enableLeaderElection = viper.GetBool(EnableLeaderElectionFlag)
	var namespaces []string = viper.GetStringSlice(NamespacesFlag)
	var namespacesRegexFilter string = viper.GetString(NamespacesRegexFilterFlag)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	opts := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "2da9ee94.carrefour.com",
	}

	switch {
	case len(namespaces) == 0:
		setupLog.Info("Phenix operator configured to manage all namespaces")
	case len(namespaces) == 1:
		setupLog.Info("Phenix operator configured to manage a single namespace", "namespace", namespaces[0])
		opts.Namespace = namespaces[0]
	default:
		setupLog.Info("Phenix operator configured to manage multiple namespaces", "namespaces", namespaces)
		opts.NewCache = cache.MultiNamespacedCacheBuilder(namespaces)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), opts)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	setupLog.Info("flags",
		MetricsAddrFlag, metricsAddr, EnableLeaderElectionFlag, enableLeaderElection,
		NamespacesFlag, namespaces, NamespacesRegexFilterFlag, namespacesRegexFilter)

	if err = (&controllers.ElasticIndexReconciler{
		Client:                mgr.GetClient(),
		Log:                   ctrl.Log.WithName("controllers").WithName("ElasticIndex"),
		Scheme:                mgr.GetScheme(),
		NamespacesRegexFilter: namespacesRegexFilter,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ElasticIndex")
		os.Exit(1)
	}
	if err = (&elasticv1alpha1.ElasticIndex{}).SetupWebhookWithManager(mgr, namespaces); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ElasticIndex")
		os.Exit(1)
	}
	if err = (&controllers.ElasticTemplateReconciler{
		Client:                mgr.GetClient(),
		Log:                   ctrl.Log.WithName("controllers").WithName("ElasticTemplate"),
		Scheme:                mgr.GetScheme(),
		NamespacesRegexFilter: namespacesRegexFilter,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ElasticTemplate")
		os.Exit(1)
	}
	if err = (&elasticv1alpha1.ElasticTemplate{}).SetupWebhookWithManager(mgr, namespaces); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ElasticTemplate")
		os.Exit(1)
	}

	if err = (&controllers.ElasticPipelineReconciler{
		Client:                mgr.GetClient(),
		Log:                   ctrl.Log.WithName("controllers").WithName("ElasticPipeline"),
		Scheme:                mgr.GetScheme(),
		NamespacesRegexFilter: namespacesRegexFilter,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ElasticPipeline")
		os.Exit(1)
	}
	if err = (&elasticv1alpha1.ElasticPipeline{}).SetupWebhookWithManager(mgr, namespaces); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ElasticPipeline")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
