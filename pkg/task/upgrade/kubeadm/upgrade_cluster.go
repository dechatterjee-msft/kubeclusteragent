package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"strings"

	"kubeclusteragent/pkg/constants"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
)

type Cluster struct{}

func (u Cluster) Name() string {
	return "upgrade-cluster"
}

var _ task.Task = &Cluster{}

func NewUpgradeCluster() *Cluster {
	t := &Cluster{}
	return t
}

func (u Cluster) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(u.Name())
	logger.Info("upgrading Kubernetes cluster using kubeadm tool", "version", clusterSpec.Version)
	out, err := ou.Kubeadm().Upgrade(ctx, clusterSpec.Version, "all")
	if err != nil {
		return err
	}
	upgradeLogs := out
	banner := fmt.Sprintf("%s \"%s\"", constants.KubeadmClusterSuccessfulUpgradeBanner, clusterSpec.Version)
	if strings.Contains(upgradeLogs, banner) {
		logger.Info("cluster upgrade successfully", "version", clusterSpec.Version)
	} else {
		err := fmt.Errorf("%s", upgradeLogs)
		logger.Error(err, "upgrade failed with error")
		return err
	}
	return nil
}

func (u Cluster) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
