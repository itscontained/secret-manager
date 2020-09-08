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

package app

import (
	"fmt"
	"time"

	"github.com/itscontained/secret-manager/cmd/controller/app/options"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	esctrl "github.com/itscontained/secret-manager/pkg/controller/externalsecret"
	"github.com/itscontained/secret-manager/pkg/util"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// kubernetes import to support cloud provider auth
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	log "k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

type Controller struct {
	options options.ControllerOptions
	manager ctrl.Manager
}

func NewController(opts *options.ControllerOptions) (*Controller, error) {
	c := &Controller{
		options: *opts,
	}

	ctrl.SetLogger(klogr.New())
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = smv1alpha1.AddToScheme(scheme)
	config, err := clientcmd.BuildConfigFromFlags(c.options.APIServerHost, c.options.Kubeconfig)
	if err != nil {
		return nil, err
	}

	gracefulShutdown := time.Second * 30
	c.manager, err = ctrl.NewManager(config, ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      fmt.Sprintf(":%d", c.options.MetricPort),
		Port:                    c.options.WebhookPort,
		Namespace:               c.options.Namespace,
		CertDir:                 c.options.TLSCertDir,
		LeaderElection:          c.options.LeaderElect,
		LeaderElectionNamespace: c.options.LeaderElectionNamespace,
		LeaderElectionID:        "secret-manager-controller",
		LeaseDuration:           &c.options.LeaderElectionLeaseDuration,
		RenewDeadline:           &c.options.LeaderElectionRenewDeadline,
		RetryPeriod:             &c.options.LeaderElectionRetryPeriod,
		ReadinessEndpointName:   "/ready",
		LivenessEndpointName:    "/health",
		HealthProbeBindAddress:  fmt.Sprintf(":%d", c.options.HealthPort),
		GracefulShutdownTimeout: &gracefulShutdown,
	})

	if err != nil {
		return nil, err
	}
	if err = (&esctrl.ExternalSecretReconciler{
		Client: c.manager.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ExternalSecret"),
		Scheme: c.manager.GetScheme(),
		Reader: c.manager.GetAPIReader(),
	}).SetupWithManager(c.manager); err != nil {
		log.Errorf("Unable to create ExternalSecret controller: %v", err.Error())
		return nil, err
	}

	err = c.manager.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		log.Errorf("Unable add a readiness check to controller: %v", err.Error())
		return nil, err
	}

	err = c.manager.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		log.Errorf("Unable add a liveness check to controller: %v", err.Error())
		return nil, err
	}

	return c, nil
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	log.Info("Starting manager")
	if err := c.manager.Start(stopCh); err != nil {
		log.Errorf("Error while running manager: %v", err.Error())
		return err
	}
	return nil
}

func NewControllerCmd(stopCh <-chan struct{}) *cobra.Command {
	opts := &options.ControllerOptions{}

	cmd := &cobra.Command{
		Use:   "secret-manager-controller",
		Short: fmt.Sprintf("Automated secret controller for Kubernetes (%s) (%s)", util.VersionInfo().GitVersion, util.VersionInfo().GitCommit),
		Long: `
secret-manager is a Kubernetes addon to automate the management and issuance of
secret sources from various external secret systems.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Validate(); err != nil {
				return fmt.Errorf("error validating options: %s", err)
			}

			log.Infof("Starting secret manager controller: version (%s) (%s)", util.VersionInfo().GitVersion, util.VersionInfo().GitCommit)
			controller, err := NewController(opts)
			if err != nil {
				log.Fatalf("Failed to start secret manager controller: %v", err.Error())
			}
			return controller.Run(stopCh)
		},
	}

	flags := cmd.Flags()
	opts.InitFlags(flags)

	return cmd
}
