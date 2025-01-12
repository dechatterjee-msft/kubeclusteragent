package common // nolint

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"

	"kubeclusteragent/pkg/task"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
)

type PurgeFiles struct {
}

var _ task.Task = &PurgeFiles{}

func NewPurgeFiles() *PurgeFiles {
	k := &PurgeFiles{}
	return k
}

func (k *PurgeFiles) Name() string {
	return "purge-files"
}

func (k *PurgeFiles) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(k.Name())
	logger.Info("Removing cluster related files")
	if err := ou.Filesystem().RemoveAll(ctx, "/etc/cni/net.d"); err != nil {
		return fmt.Errorf("removing cni file(/etc/cni/net.d): %w", err)
	}
	return nil
}

func (k *PurgeFiles) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
