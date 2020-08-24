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

package options

import (
	"time"

	"github.com/spf13/pflag"
)

type ControllerOptions struct {
	APIServerHost string
	Kubeconfig    string
	Namespace     string

	LeaderElect                 bool
	LeaderElectionNamespace     string
	LeaderElectionLeaseDuration time.Duration
	LeaderElectionRenewDeadline time.Duration
	LeaderElectionRetryPeriod   time.Duration

	EnabledControllers []string

	WebhookPort int
	HealthPort  int
	// The host and port address, separated by a ':', that the Prometheus server
	// should expose metrics on.
	MetricsListenAddress string

	// Path to TLS certificate and private key on disk.
	// The server key and certificate must be named tls.key and tls.crt, respectively.
	TLSCertDir string

	// TLSCipherSuites is the list of allowed cipher suites for the server.
	// Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants).
	TLSCipherSuites []string

	// MinTLSVersion is the minimum TLS version supported.
	// Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants).
	MinTLSVersion string
}

var (
	defaultEnabledControllers = make([]string, 0)
)

func (s *ControllerOptions) InitFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.APIServerHost, "master", "", ""+
		"Optional apiserver host address to connect to. If not specified, autoconfiguration "+
		"will be attempted.")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", "", ""+
		"Paths to a kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&s.Namespace, "namespace", "", ""+
		"If set, this limits the scope of secret-manager to a single namespace and ClusterIssuers are disabled. "+
		"If not specified, all namespaces will be watched")

	fs.BoolVar(&s.LeaderElect, "leader-elect", true, ""+
		"If true, secret-manager will perform leader election between instances to ensure no more "+
		"than one instance of secret-manager operates at a time")
	fs.StringVar(&s.LeaderElectionNamespace, "leader-election-namespace", "kube-system", ""+
		"Namespace used to perform leader election. Only used if leader election is enabled")
	fs.DurationVar(&s.LeaderElectionLeaseDuration, "leader-election-lease-duration", 60*time.Second, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&s.LeaderElectionRenewDeadline, "leader-election-renew-deadline", 45*time.Second, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&s.LeaderElectionRetryPeriod, "leader-election-retry-period", 15*time.Second, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")

	fs.StringSliceVar(&s.EnabledControllers, "controllers", defaultEnabledControllers, ""+
		"The set of controllers to enable.")

	fs.StringVar(&s.MetricsListenAddress, "metrics-listen-address", "0.0.0.0:9321", ""+
		"The host and port that the metrics endpoint should listen on.")
	fs.IntVar(&s.WebhookPort, "webhook-port", 8443, ""+
		"The port number to listen on for webhook connections.")
	fs.IntVar(&s.HealthPort, "health-port", 8400, ""+
		"The port number to listen on for health connections.")

	fs.StringVar(&s.TLSCertDir, "tls-cert-dir", "", ""+
		"The path to TLS certificate and private key on disk.")
}

func (s *ControllerOptions) Validate() error {
	return nil
}
