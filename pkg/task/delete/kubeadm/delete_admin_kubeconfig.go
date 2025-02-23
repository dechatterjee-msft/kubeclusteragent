package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
)

type DeleteAdminKubeConfig struct {
}

var _ task.Task = &DeleteAdminKubeConfig{}

func NewDeleteAdminKubeConfig() *DeleteAdminKubeConfig {
	t := &DeleteAdminKubeConfig{}
	return t
}

func (r *DeleteAdminKubeConfig) Name() string {
	return "delete-admin-kube-config"
}

func (r *DeleteAdminKubeConfig) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(r.Name())
	logger.Info("deleting current user kubeconfig")
	err := ou.Filesystem().RemoveAll(ctx, constants.AdminKubeconfigDirPath)
	if err != nil {
		return fmt.Errorf("error while deleting admin kube-config: %w", err)
	}
	return nil
}

func (r *DeleteAdminKubeConfig) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
