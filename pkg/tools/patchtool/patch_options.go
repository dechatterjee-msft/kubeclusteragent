package patchtool

import (
	"kubeclusteragent/pkg/operations"
	"kubeclusteragent/pkg/task"
	kubeadmCreate "kubeclusteragent/pkg/task/install/kubeadm"
	"kubeclusteragent/pkg/task/patch"
	"kubeclusteragent/pkg/util/osutility"
)

func buildPatchOptions(options ...operations.Option) operations.TaskDetails {
	current := operations.TaskDetails{
		Tasks: []task.Task{
			kubeadmCreate.NewInstallCNI(),
		},
		OsUtil: osutility.New(),
		PostTasks: []task.Task{
			patch.UpdateWorkloadScheduler(),
		},
	}
	for _, o := range options {
		o(&current)
	}

	return current
}
