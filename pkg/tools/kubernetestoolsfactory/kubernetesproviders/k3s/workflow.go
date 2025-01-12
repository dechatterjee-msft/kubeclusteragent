package k3s

import (
	"kubeclusteragent/pkg/operations"
	"kubeclusteragent/pkg/task"
	k3sInstall "kubeclusteragent/pkg/task/install/k3s"
	"kubeclusteragent/pkg/util/osutility"
)

func buildInstallOptions(options ...operations.Option) operations.TaskDetails {
	current := operations.TaskDetails{
		Tasks:     []task.Task{k3sInstall.NewInstallCluster()},
		PostTasks: []task.Task{k3sInstall.NewInstallCNI()},
		OsUtil:    osutility.New(),
	}
	for _, o := range options {
		o(&current)
	}

	return current
}

func buildUpgradeOptions(options ...operations.Option) operations.TaskDetails {
	current := operations.TaskDetails{
		Tasks: []task.Task{
			k3sInstall.NewInstallCluster(),
		},
		OsUtil: osutility.New(),
	}
	for _, o := range options {
		o(&current)
	}

	return current
}

//func buildResetOptions(options ...operations.Option) operations.TaskDetails {
//	current := operations.TaskDetails{
//		PreTasks: []task.Task{
//			k3sDelete.NewK3sKillServer(),
//		},
//		Tasks: []task.Task{
//			k3sDelete.NewResetCluster(),
//		},
//		PostTasks: []task.Task{
//			k3sDelete.NewK3sCleanup(),
//		},
//		OsUtil: osutil.New(),
//	}
//	for _, o := range options {
//		o(&current)
//	}
//
//	return current
//}
//
//func buildCertsRotationOptions(options ...operations.Option) operations.TaskDetails {
//	current := operations.TaskDetails{
//		Tasks: []task.Task{
//			k3sCerts.NewRotateCerts(),
//		},
//		OsUtil: osutil.New(),
//	}
//	for _, o := range options {
//		o(&current)
//	}
//
//	return current
//}
