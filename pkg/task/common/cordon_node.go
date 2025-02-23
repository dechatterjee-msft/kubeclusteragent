package common

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/k8s"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
)

type CordonNode struct{}

var _ task.Task = &CordonNode{}

func NewCordonNode() *CordonNode {
	t := &CordonNode{}
	return t
}

func (t *CordonNode) Name() string {
	return "cordon-node"
}

func (t *CordonNode) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	k8sUtility := k8s.K8sUtil{}
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	err := k8sUtility.NodeWorkloadScheduler(ctx, "cordon")
	if err != nil {
		logger.Error(err, "failed to cordon node")
		return err
	}
	logger.Info("successfully cordon node")
	return nil
}

func (t *CordonNode) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
