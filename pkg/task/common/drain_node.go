package common

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	osutil2 "kubeclusteragent/pkg/util/osutility"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
)

type DrainNode struct{}

var _ task.Task = &DrainNode{}

func NewDrainNode() *DrainNode {
	t := &DrainNode{}
	return t
}

func (t *DrainNode) Name() string {
	return "drain-node"
}

func (t *DrainNode) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutil2.OSUtil) error {
	var hostUitl osutil2.Host = &osutil2.LiveHost{}
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("Running drain node task")
	nodeName, err := hostUitl.GetHostname()
	logger.Info("draining node before upgrade", "node", nodeName)
	if err != nil {
		return fmt.Errorf("kubectl run:  %w", err)
	}
	err = ou.Kubectl().Run(ctx, []string{"drain", nodeName, "--ignore-daemonsets"}...)
	if err != nil {
		return fmt.Errorf("kubectl run  : %w", err)
	}
	return nil
}

func (t *DrainNode) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutil2.OSUtil) error {
	return nil
}
