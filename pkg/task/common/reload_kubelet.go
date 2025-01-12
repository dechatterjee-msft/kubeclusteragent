package common

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
)

type KubeletReload struct{}

var _ task.Task = &KubeletReload{}

func NewKubeletReload() *KubeletReload {
	t := &KubeletReload{}
	return t
}

func (t *KubeletReload) Name() string {
	return "kubelet-reload"
}

func (t *KubeletReload) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("Running kubelet reload task")
	err := ou.Systemd().DaemonReload(ctx)
	if err != nil {
		return fmt.Errorf("systemd run:  %w", err)
	}
	err = ou.Systemd().Restart(ctx, "kubelet")
	if err != nil {
		return fmt.Errorf("systemctl run  : %w", err)
	}
	return nil
}

func (t *KubeletReload) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
