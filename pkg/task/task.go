package task

import (
	"context"
	"kubeclusteragent/pkg/util/osutility"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
)

var (
	DefaultKubernetesModules = []string{"br_netfilter", "overlay"}
	DefaultKernelSettings    = map[string]string{
		"net.ipv4.ip_forward":                 "1",
		"net.bridge.bridge-nf-call-ip6tables": "1",
		"net.bridge.bridge-nf-call-iptables":  "1",
	}
)

type Task interface {
	Name() string
	Run(
		ctx context.Context,
		status cluster.Status,
		clusterSpec *v1alpha1.ClusterSpec,
		ou osutility.OSUtil) error
	Rollback(ctx context.Context,
		status cluster.Status,
		clusterSpec *v1alpha1.ClusterSpec,
		ou osutility.OSUtil) error
}
