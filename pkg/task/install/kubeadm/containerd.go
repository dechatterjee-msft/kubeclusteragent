package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"

	"kubeclusteragent/pkg/task"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
)

type ContainerdInstall struct{}

var _ task.Task = &ContainerdInstall{}

func NewInstallContainerd() *ContainerdInstall {
	t := &ContainerdInstall{}
	return t
}

func (t *ContainerdInstall) Name() string {
	return "containerd-install"
}

func (t *ContainerdInstall) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx)
	logger.Info("Running containerd task")

	if !ou.PackageManager().CheckInstalled(ctx, "containerd") {
		if err := ou.PackageManager().Update(ctx); err != nil {
			return fmt.Errorf("update packages: %w", err)
		}
		if err := ou.PackageManager().Install(ctx, "containerd"); err != nil {
			return fmt.Errorf("install containerd: %w", err)
		}
	}

	isRunning, err := ou.Systemd().IsRunning(ctx, "containerd")
	if err != nil {
		return fmt.Errorf("check if containerd is running: %w", err)
	}

	if !isRunning {
		if err := ou.Systemd().Start(ctx, "containerd"); err != nil {
			return fmt.Errorf("start containerd: %w", err)
		}
	}

	return nil
}

func (t *ContainerdInstall) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
