package k3s

import (
	"context"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	osutil2 "kubeclusteragent/pkg/util/osutility"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
)

type ClusterUpgradeController struct{}

var _ task.Task = &ClusterUpgradeController{}

var defaultUpgradeController = "https://github.com/rancher/system-upgrade-controller/releases/latest/download/system-upgrade-controller.yaml"

func NewK3sClusterUpgradeController() *ClusterUpgradeController {
	t := &ClusterUpgradeController{}
	return t
}

func (t *ClusterUpgradeController) Name() string {
	return "install-cluster-upgrade-controller"
}

func (t *ClusterUpgradeController) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutil2.OSUtil) error {
	logger := log.From(ctx).WithValues(
		"Cluster Type", clusterSpec.ClusterType,
		"Version", clusterSpec.Version,
		"Operation", "Install", "Task", t.Name())
	logger.Info("installing k3s cluster upgrade controller")
	var kubectlClient osutil2.Kubectl = &osutil2.K3sLiveKubectl{}
	logger.Info("creating system-upgrade namespace for upgrade controller")
	err := kubectlClient.Run(ctx, []string{"create", "ns", "system-upgrade"}...)
	if err != nil {
		logger.Error(err, "unable to create system upgrade namespace")
		return err
	}
	err = kubectlClient.Run(ctx, []string{"apply", "-f", defaultUpgradeController}...)
	if err != nil {
		logger.Error(err, "unable to create upgrade controller")
		return err
	}
	return nil
}

func (t *ClusterUpgradeController) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutil2.OSUtil) error {
	return nil
}
