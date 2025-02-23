package common

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/k8s"
	"kubeclusteragent/pkg/util/osutility/linux"
)

type UnCordonNode struct{}

var _ task.Task = &UnCordonNode{}

func NewUnCordonNode() *UnCordonNode {
	t := &UnCordonNode{}
	return t
}

func (t *UnCordonNode) Name() string {
	return "uncordon-node"
}

func (t *UnCordonNode) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	k8sUtility := k8s.K8sUtil{}
	return k8sUtility.NodeWorkloadScheduler(ctx, "uncordon")
}

func (t *UnCordonNode) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
