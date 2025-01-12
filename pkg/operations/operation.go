package operations

import (
	"context"
	"fmt"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
)

type Operation struct {
	name          string
	preTasks      []task.Task
	tasks         []task.Task
	postTasks     []task.Task
	osUtil        osutility.OSUtil
	clusterStatus cluster.Status
	clusterSpec   *v1alpha1.ClusterSpec
}

func NewOperation(name string, clusterStatus cluster.Status, clusterSpec *v1alpha1.ClusterSpec, taskDetails TaskDetails) *Operation {
	o := &Operation{
		name:          name,
		clusterStatus: clusterStatus,
		clusterSpec:   clusterSpec,
		preTasks:      taskDetails.PreTasks,
		tasks:         taskDetails.Tasks,
		postTasks:     taskDetails.PostTasks,
		osUtil:        taskDetails.OsUtil,
	}
	return o
}

func (o *Operation) Run(ctx context.Context) error {
	logger := log.From(ctx).WithName(o.name).WithValues("ClusterType", o.clusterSpec.ClusterType, "version", o.clusterSpec.Version)
	logger.Info("Starting operation:", "name", o.name)
	if err := o.runTasks(ctx); err != nil {
		return err
	}
	logger.Info("Operation completed:", "name", o.name)
	return nil
}

func (o *Operation) runTasks(ctx context.Context) error {
	for _, t := range o.preTasks {
		if err := o.runTask(ctx, t); err != nil {
			return fmt.Errorf("failed pre-task (%s): %w", t.Name(), err)
		}
	}

	for _, t := range o.tasks {
		if err := o.runTask(ctx, t); err != nil {
			return fmt.Errorf("failed install task (%s): %w", t.Name(), err)
		}
	}

	for _, t := range o.postTasks {
		if err := o.runTask(ctx, t); err != nil {
			return fmt.Errorf("failed install post-task (%s): %w", t.Name(), err)
		}
	}

	return nil
}

func (o *Operation) runTask(ctx context.Context, t task.Task) error {
	logger := log.From(ctx).WithName(o.name).WithValues("task", t.Name())
	ctx = log.WithExistingLogger(ctx, logger)
	return t.Run(ctx, o.clusterStatus, o.clusterSpec, o.osUtil)
}
