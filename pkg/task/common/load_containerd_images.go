package common

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/cri"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
)

type LoadContainerdImages struct{}

var _ task.Task = &LoadContainerdImages{}

func NewLoadContainerdImages() *LoadContainerdImages {
	t := &LoadContainerdImages{}
	return t
}

func (t *LoadContainerdImages) Name() string {
	return "load-containerd-images"
}

func (t *LoadContainerdImages) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	err := cri.LoadContainerdImages(ctx, ou, logger)
	if err != nil {
		logger.Error(err, "error occurred on loading containerd images")
	}
	return err
}

func (t *LoadContainerdImages) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
