package common

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/k8s"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"os"
)

type CoreDNSRestore struct{}

var _ task.Task = &CoreDNSRestore{}

func NewCoreDNSRestore() *CoreDNSRestore {
	t := &CoreDNSRestore{}
	return t
}

func (t *CoreDNSRestore) Name() string {
	return "coredns-restore"
}

func (t *CoreDNSRestore) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
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
	configMap, err := status.GetConfigMap(ctx, "coredns")
	if err != nil {
		logger.Error(err, "unable to get core-dns config")
		return err
	}
	err = k8s.CreateConfigMap(client, ctx, "kube-system", configMap)
	if err != nil {
		logger.Error(err, "unable to create coredns config-map")
		return err
	}
	logger.Info("core-dns restore successful")
	return nil
}

func (t *CoreDNSRestore) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
