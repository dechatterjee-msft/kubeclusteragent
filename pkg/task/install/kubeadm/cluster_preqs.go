package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"os"
)

type ClusterPrerequisites struct{}

var _ task.Task = &ClusterPrerequisites{}

func NewClusterPrerequisites() *ClusterPrerequisites {
	t := &ClusterPrerequisites{}
	return t
}

func (c ClusterPrerequisites) Name() string {
	return "cluster-prerequisites"
}

func (c ClusterPrerequisites) Run(ctx context.Context, status cluster.Status, clusterSpec *v1alpha1.ClusterSpec, ou osutility.OSUtil) error {
	logger := log.From(ctx)
	logger.Info("running kubernetes prerequisites")
	code, _, err := ou.Exec().CommandWithNoLogging(ctx, "swapoff", nil, []string{"-a"}...)
	if code != 0 {
		err = fmt.Errorf("failed to swapoff kubernetes prerequisites")
		logger.Error(err, "failed to switch off swap", "code", code)
		return err
	}
	if err != nil {
		logger.Error(err, "failed to switch off swap")
		return err
	}
	// backup /etc/fstab
	code, _, err = ou.Exec().CommandWithNoLogging(ctx, "cp", nil, []string{"/etc/fstab", "/etc/fstab.bak"}...)
	if code != 0 {
		err = fmt.Errorf("failed to backup /etc/fstab.bak")
		logger.Error(err, "failed to backup /etc/fstab.bak")
		return err
	}
	code, _, err = ou.Exec().CommandWithNoLogging(ctx, "sed", nil, []string{"-i.bak", "-e", "/\\sswap\\s/s/^/#/", "/etc/fstab"}...)
	if code != 0 {
		err = fmt.Errorf("failed to update swappiness in /etc/fstab")
		logger.Error(err, "failed to switch off swap", "code", code)
		return err
	}
	if err != nil {
		logger.Error(err, "failed to update swap/d in /etc/fstab")
		return err
	}
	file, err := ou.Filesystem().Open(ctx, constants.KubernetesKernelModuleFile)
	if err != nil {
		logger.Error(err, "unable to read or open", "filename", constants.KubernetesKernelModuleFile)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)
	content := `overlay
br_netfilter
`
	// Write content to the file
	_, err = file.WriteString(content)
	if err != nil {
		logger.Error(err, "unable to load kernel content to the file /etc/modules-load.d/k8s.conf")
		return err
	}
	_, _, err = ou.Exec().Command(ctx, "modprobe", []string{"overlay"})
	if err != nil {
		logger.Error(err, "unable to load kernel module overlay")
		return err
	}
	_, _, err = ou.Exec().Command(ctx, "modprobe", []string{"br_netfilter"})
	if err != nil {
		logger.Error(err, "unable to load kernel module br_netfilter")
		return err
	}

	content = `net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables  = 1
net.ipv4.ip_forward                 = 1
`
	file, err = ou.Filesystem().Open(ctx, constants.KubernetesSysctlModuleFile)
	if err != nil {
		logger.Error(err, "unable to read or open",
			"filename", constants.KubernetesSysctlModuleFile)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	// Write content to the file
	_, err = file.WriteString(content)
	if err != nil {
		logger.Error(err, "unable to load kernel content to the file", "filename", constants.KubernetesSysctlModuleFile)
		return err
	}
	err = ou.Sysctl().Reload(ctx)
	if err != nil {
		logger.Error(err, "unable to reload sysctl")
		return err
	}

	return nil
}

func (c ClusterPrerequisites) Rollback(ctx context.Context, status cluster.Status, clusterSpec *v1alpha1.ClusterSpec, ou osutility.OSUtil) error {
	return nil
}
