package patch

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/k8s"
	"kubeclusteragent/pkg/util/osutility/linux"
)

type UpdateNodeWorkloadSchedule struct{}

var _ task.Task = &UpdateNodeWorkloadSchedule{}

func UpdateWorkloadScheduler() *UpdateNodeWorkloadSchedule {
	t := &UpdateNodeWorkloadSchedule{}
	return t
}

func (t *UpdateNodeWorkloadSchedule) Name() string {
	return "workload-schedule"
}

func (t *UpdateNodeWorkloadSchedule) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	// This is only applicable for Patch request , it may happen user wants to disable the workload during cluster creation or upgrade
	currentClusterSpec := status.GetSpec(ctx)
	k8sUtility := k8s.K8sUtil{}
	if currentClusterSpec.DisableWorkloads != nil {
		if *currentClusterSpec.DisableWorkloads {
			return k8sUtility.NodeWorkloadScheduler(ctx, "cordon")
		}
		return k8sUtility.NodeWorkloadScheduler(ctx, "uncordon")
	}
	return nil
}

func (t *UpdateNodeWorkloadSchedule) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
