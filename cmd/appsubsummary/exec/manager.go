// Copyright 2021 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"fmt"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	appsubapi "open-cluster-management.io/multicloud-operators-subscription/pkg/apis"
	"open-cluster-management.io/multicloud-operators-subscription/pkg/controller"
	"open-cluster-management.io/multicloud-operators-subscription/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost             = "0.0.0.0"
	metricsPort         int = 8391
	operatorMetricsPort int = 8691
)

// RunManager starts the actual manager.
func RunManager() {
	enableLeaderElection := false

	if _, err := rest.InClusterConfig(); err == nil {
		klog.Info("LeaderElection enabled as running in a cluster")

		enableLeaderElection = true
	} else {
		klog.Info("LeaderElection disabled as not running in a cluster")
	}

	klog.Info("kubeconfig:" + options.KubeConfig)

	var err error

	cfg := ctrl.GetConfigOrDie()

	if options.KubeConfig != "" {
		cfg, err = utils.GetClientConfigFromKubeConfig(options.KubeConfig)

		if err != nil {
			klog.Error(err, "")
			os.Exit(1)
		}
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		MetricsBindAddress:         fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		Port:                       operatorMetricsPort,
		LeaderElection:             enableLeaderElection,
		LeaderElectionID:           "multicloud-operators-appsubsummary-leader.open-cluster-management.io",
		LeaderElectionResourceLock: "configmaps",
	})
	if err != nil {
		klog.Error(err, "")
		os.Exit(1)
	}

	klog.Info("Registering AppSubSummary component.")

	// Setup Scheme for all resources
	if err := appsubapi.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers.
	if err := controller.AddAppSubSummaryToManager(mgr, options.SyncInterval); err != nil {
		klog.Error(err, "")
		os.Exit(1)
	}

	sig := signals.SetupSignalHandler()

	klog.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(sig); err != nil {
		klog.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}
