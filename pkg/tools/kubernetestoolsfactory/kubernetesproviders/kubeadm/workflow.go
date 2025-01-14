package kubeadm

import (
	"kubeclusteragent/pkg/operations"
	"kubeclusteragent/pkg/task"
	kubeadmCerts "kubeclusteragent/pkg/task/certs/kubeadm"
	"kubeclusteragent/pkg/task/common"
	kubeadmReset "kubeclusteragent/pkg/task/delete/kubeadm"
	kubeadmCreate "kubeclusteragent/pkg/task/install/kubeadm"
	"kubeclusteragent/pkg/task/patch"
	kubeadmUpgrade "kubeclusteragent/pkg/task/upgrade/kubeadm"
	"kubeclusteragent/pkg/util/osutility"
)

func buildInstallOptions(options ...operations.Option) operations.TaskDetails {
	current := operations.TaskDetails{
		PreTasks: []task.Task{
			kubeadmCreate.NewClusterPrerequisites(),
			kubeadmCreate.NewInstallContainerd(),
			kubeadmCreate.NewInstallBinaries(),
			kubeadmCreate.NewPrepareContainerd()},

		Tasks: []task.Task{kubeadmCreate.NewInstallCluster()},

		PostTasks: []task.Task{
			kubeadmCreate.NewRemoveTaint(),
			patch.UpdateWorkloadScheduler(),
			kubeadmCreate.NewInstallCNI(),
			common.NewNodeReady(),
			//	kubeadmCreate.NewCoredns(),
			//	kubeadmCreate.NewInstallCSI(),
			kubeadmCreate.NewCurrentUserKubeconfig(),
		},
		OsUtil: osutility.New(),
	}
	for _, o := range options {
		o(&current)
	}

	return current
}

func buildUpgradeOptions(options ...operations.Option) operations.TaskDetails {
	current := operations.TaskDetails{
		PreTasks: []task.Task{
			common.NewCordonNode(),
			common.NewLoadContainerdImages(),
			common.NewCoreDNSBackup(),
		},
		Tasks: []task.Task{
			common.NewNodeReady(),
			kubeadmUpgrade.NewUpgradeCluster()},

		PostTasks: []task.Task{
			common.NewKubeletReload(),
			common.NewNodeReady(),
			common.NewCoreDNSRestore(),
			common.NewUnCordonNode(),
			common.NewCleanUpK8sControlPlaneContainerdImages(),
			kubeadmCerts.NewRotateAdminCerts(),
		},
		OsUtil: osutility.New(),
	}
	for _, o := range options {
		o(&current)
	}

	return current
}

func buildResetOptions(options ...operations.Option) operations.TaskDetails {
	current := operations.TaskDetails{
		Tasks: []task.Task{
			kubeadmReset.NewKubeadmReset(),
			common.NewPurgeFiles(),
		},
		OsUtil: osutility.New(),
	}
	for _, o := range options {
		o(&current)
	}

	return current
}

func buildCertsRotationOptions(options ...operations.Option) operations.TaskDetails {
	current := operations.TaskDetails{
		Tasks: []task.Task{
			kubeadmCerts.NewRotateCerts(),
			common.NewRestartK8sControlplane(),
		},
		OsUtil: osutility.New(),
	}
	for _, o := range options {
		o(&current)
	}

	return current
}
