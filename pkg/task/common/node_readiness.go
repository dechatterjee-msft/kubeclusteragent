package common

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"strings"
	"time"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
)

type NodeReady struct{}

var _ task.Task = &NodeReady{}

func NewNodeReady() *NodeReady {
	n := &NodeReady{}
	return n
}

func (n *NodeReady) Name() string {
	return "node-readiness"
}

func (n *NodeReady) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(n.Name())
	logger.Info("checking node readiness")
	var retryCount = 0
nodeReady:
	data, err := ou.Kubectl().RunWithResponse(ctx, []string{"get", "nodes"}...)
	if err != nil {
		return fmt.Errorf("kubectl run  : %w", err)
	}
	if !strings.Contains(data, " Ready") {
		if strings.Contains(data, "NotReady") ||
			strings.Contains(data, "Unknown") ||
			strings.Contains(data, "did you specify the right host or port?") {
			time.Sleep(constants.NodeReadinessRetryInterval)
			retryCount++
			if retryCount <= constants.NodeReadinessMaxRetryCount {
				logger.Info("retrying node readiness")
				goto nodeReady
			} else {
				logger.Error(fmt.Errorf("node is  not ready"), "node is not ready after %vs , failing the cluster installation", constants.NodeReadinessMaxRetryCount*constants.NodeReadinessRetryInterval.Seconds())
				return fmt.Errorf("node is  not ready")
			}
		}
	}
	return nil
}

func (n *NodeReady) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
