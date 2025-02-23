// nolint
package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/cni/cilium"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
)

type Cni struct{}

var _ task.Task = &Cni{}

func NewInstallCNI() *Cni {
	t := &Cni{}
	return t
}

func (t *Cni) Name() string {
	return "install-cni"
}

func (t *Cni) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("installing cni")
	if clusterSpec.Networking.CniName == "Calico" {
		// TODO install calico
		logger.Info("cni is calico")
	}
	if clusterSpec.Networking.CniName == "Cilium" {
		return cilium.Install(ctx, logger, clusterSpec.Networking.Cilium.CliVersion, ou)
	}
	response, err := ou.Kubectl().RunWithResponse(ctx, "apply", "-f", clusterSpec.Networking.CniManifestURL)
	if err != nil {
		logger.Error(err, "unable to apply CNI present in the give location", "path", clusterSpec.Networking.CniManifestURL)
		return err
	}
	logger.Info("cni installation response", "response", response)
	return nil
}

func (t *Cni) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
