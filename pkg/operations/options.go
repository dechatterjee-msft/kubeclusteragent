package operations

import (
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/osutility/linux"
)

type TaskDetails struct {
	PreTasks  []task.Task
	Tasks     []task.Task
	PostTasks []task.Task
	OsUtil    linux.OSUtil
}

type Option func(o *TaskDetails)

func DryRun() Option {
	return func(o *TaskDetails) {
		o.OsUtil = linux.NewDryRun()
	}
}
