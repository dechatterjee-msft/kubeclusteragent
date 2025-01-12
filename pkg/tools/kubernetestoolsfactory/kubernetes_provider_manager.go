package kubernetestoolsfactory

import (
	"context"
	"kubeclusteragent/pkg/util/log/log"
	"time"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	kubernetestool "kubeclusteragent/pkg/tools/kubernetestoolsfactory/kubernetesproviders"
	k3sTool "kubeclusteragent/pkg/tools/kubernetestoolsfactory/kubernetesproviders/k3s"
	kubeadmtool "kubeclusteragent/pkg/tools/kubernetestoolsfactory/kubernetesproviders/kubeadm"
)

type KubeManager struct{}

type KubeToolsFactory interface {
	GetKubernetesProviderOnStartup(ctx context.Context) kubernetestool.KubernetesProviderFactory
	GetKubernetesToolByProvider(ctx context.Context, clusterType string) kubernetestool.KubernetesProviderFactory
	KubernetesClusterUpgradeManager(ctx context.Context, version string) error
}

func (k *KubeManager) GetKubernetesProviderOnStartup(ctx context.Context) kubernetestool.KubernetesProviderFactory {
	var clusterStatus cluster.Status = &cluster.LiveStatus{}
	existingClusterSpec := clusterStatus.GetSpec(ctx)
	if existingClusterSpec != nil {
		return k.GetKubernetesToolByProvider(ctx, existingClusterSpec.ClusterType)
	}
	return kubeadmtool.NewKubeadmInstallTool(clusterStatus, false)
}

func (k *KubeManager) GetKubernetesToolByProvider(ctx context.Context, clusterType string) kubernetestool.KubernetesProviderFactory {
	var clusterStatus cluster.Status = &cluster.LiveStatus{}
	switch clusterType {
	case "k3s":
		return k3sTool.NewK3sInstallTool(clusterStatus, false)
	default:
		return kubeadmtool.NewKubeadmInstallTool(clusterStatus, false)
	}
}

func (k *KubeManager) KubernetesClusterUpgradeManager(ctx context.Context, version string) error {
	logger := log.From(ctx).WithName("cluster-upgrade-manager")
	status := &cluster.LiveStatus{}
	clusterStatus := status.GetStatus(ctx)
	spec := status.GetSpec(ctx)
	kubeadmTool := kubeadmtool.NewKubeadmInstallTool(status, false)
	upgradeCluster := &v1alpha1.UpgradeClusterRequest{
		ApiVersion: "v1alpha1",
		Kind:       "Cluster",
		Spec: &v1alpha1.ClusterSpec{
			ClusterType: "kubeadm",
			ClusterName: spec.ClusterName,
			Version:     version,
		},
	}
	err := kubeadmTool.Upgrade(ctx, upgradeCluster)
	if err != nil {
		logger.Error(err, "cluster upgrade failed")
		return err
	}
	// wait for upgrade to be completed
	waitCount := 0
	for status.GetStatus(ctx).GetPhase() != constants.ClusterPhaseProvisioned && waitCount < constants.ClusterUpgradeWaitDuration {
		logger.Info("cluster upgrade is in progress,reconciler will wait for the upgrade to be completed")
		time.Sleep(1 * time.Minute)
		waitCount++
	}
	if waitCount >= constants.ClusterUpgradeWaitDuration && clusterStatus.KubernetesVersion != version {
		logger.Info("unable to upgrade cluster in 10 minutes window, agent will retry in next reconciliation loop")
	} else {
		logger.Info("cluster successfully upgraded", "version", clusterStatus.KubernetesVersion)
	}
	return nil
}
