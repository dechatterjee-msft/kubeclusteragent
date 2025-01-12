package common

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/k8s"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"os"
)

type CoreDNSBackup struct{}

var _ task.Task = &CoreDNSBackup{}

func NewCoreDNSBackup() *CoreDNSBackup {
	t := &CoreDNSBackup{}
	return t
}

func (t *CoreDNSBackup) Name() string {
	return "coredns-backup"
}

func (t *CoreDNSBackup) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	kc, err := os.ReadFile(constants.KubeadmKubeconfigPath)
	if err != nil {
		logger.Error(err, "unable to read kubeconfig file from path", "path", constants.KubeadmKubeconfigPath)
		return err
	}
	client, err := k8s.GetKubeClientFromKubeconfig(string(kc))
	if err != nil {
		logger.Error(err, "unable to make connection with kubernetes api server")
		return err
	}
	corednsConfigMap, err, isNotFound := k8s.CopyConfigMap(client, ctx, "kube-system", "coredns")
	if isNotFound {
		logger.Info("coredns config-map is not found in the cluster,checking for backup")
		configMap, err := status.GetConfigMap(ctx, "coredns")
		if configMap != nil {
			return nil
		}
		logger.Error(err, "coredns config is not found in the backup")
		return err
	}
	if err != nil {
		logger.Error(err, "unable to copy core-dns config")
		return err
	}
	err = status.StoreConfigMap(ctx, corednsConfigMap, "coredns")
	if err != nil {
		logger.Error(err, "unable to store coredns config-map")
		return err
	}
	err = k8s.DeleteConfigMap(client, ctx, "kube-system", "coredns")
	if err != nil {
		logger.Error(err, "unable remove coredns-config map")
		return err
	}
	logger.Info("core dns config map backup successful")
	return nil
}

func (t *CoreDNSBackup) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
