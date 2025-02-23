package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"

	"kubeclusteragent/pkg/task"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
)

type Reset struct {
}

var _ task.Task = &Reset{}

func NewKubeadmReset() *Reset {
	t := &Reset{}
	return t
}

func (r *Reset) Name() string {
	return "kubeadm-reset"
}

func (r *Reset) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(r.Name())
	logger.Info("resetting kubeadm")
	code, data, err := ou.Exec().Command(ctx, "kubeadm", nil, "reset", "-f")
	if err != nil {
		return fmt.Errorf("resetting cluster with kubeadm: %w", err)
	}
	if code != 0 {
		resetError := fmt.Errorf("reset did not return a 0 error code")
		logger.Info("Failed reset output", "output", string(data))
		return resetError
	}
	return nil
}

func (r *Reset) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
