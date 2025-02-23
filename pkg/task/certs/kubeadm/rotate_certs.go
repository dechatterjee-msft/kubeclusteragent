package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
)

type CertsRotation struct{}

func (u CertsRotation) Name() string {
	return "kubeadm-rotate-certs"
}

var _ task.Task = &CertsRotation{}

func NewRotateCerts() *CertsRotation {
	t := &CertsRotation{}
	return t
}

func (u CertsRotation) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(u.Name())
	all, err := ou.Kubeadm().CertsRotateAll(ctx)
	if err != nil {
		return fmt.Errorf("run kubeadm certs renew: %w", err)
	}
	logger.Info("certs rotation logs", "info", all)
	return nil
}

func (u CertsRotation) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
