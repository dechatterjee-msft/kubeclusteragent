package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
)

type Coredns struct{}

var _ task.Task = &Coredns{}

func NewCoredns() *Coredns {
	t := &Coredns{}

	return t
}

func (t *Coredns) Name() string {
	return "scale-coredns"
}

func (t *Coredns) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	response, err := ou.Kubectl().RunWithResponse(ctx, []string{"scale", "deployment", "coredns", "-n", "kube-system", "--replicas=1"}...)
	if err != nil {
		logger.Error(err, "error occurred during coredns scale down")
		return err
	}
	logger.Info("scale in conredns deployment output", response)
	return nil
}

func (t *Coredns) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
