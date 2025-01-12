package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"strings"
)

type RemoveTaint struct{}

var _ task.Task = &RemoveTaint{}

func NewRemoveTaint() *RemoveTaint {
	t := &RemoveTaint{}
	return t
}

func (t *RemoveTaint) Name() string {
	return "remove-controlplane-taint"
}

func (t *RemoveTaint) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	output, err := ou.Kubectl().RunWithResponse(ctx, "taint", "nodes", "--all", "node-role.kubernetes.io/control-plane-")
	if err != nil {
		logger.Error(err, "remove controlplane node taint error", "Taint", "node-role.kubernetes.io/control-plane-")
		return fmt.Errorf("remove master node taint: %w", err)
	}
	if !strings.Contains(output, "untainted") {
		err = fmt.Errorf("error while removing taint %v", err)
		logger.Error(err, "failed to remove taint", "Taint", "node-role.kubernetes.io/control-plane-", "output", output)
		return err
	}
	logger.Info("Remove taint output form controlplane", "output", output)
	return nil
}

func (t *RemoveTaint) Rollback(ctx context.Context, // nolint
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
