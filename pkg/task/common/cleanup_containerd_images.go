package common

import (
	"context"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/cri"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
)

type CleanUpK8sControlPlaneContainerdImages struct{}

var _ task.Task = &CleanUpK8sControlPlaneContainerdImages{}

func NewCleanUpK8sControlPlaneContainerdImages() *CleanUpK8sControlPlaneContainerdImages {
	t := &CleanUpK8sControlPlaneContainerdImages{}
	return t
}

func (t *CleanUpK8sControlPlaneContainerdImages) Name() string {
	return "k8s-controlplane-images-cleanup"
}

func (t *CleanUpK8sControlPlaneContainerdImages) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	containerdClient, err := cri.NewConnection(constants.ContainerdAddress, constants.ContainerdKubernetesNamespace)
	defer func(containerdClient cri.Client, ctx context.Context) {
		err := containerdClient.Close(ctx)
		if err != nil {
			logger.Error(err, "error occurred while closing the containerd connection",
				"address", constants.ContainerdAddress)
			return
		}
	}(containerdClient, ctx)
	if err != nil {
		logger.Error(err, "error occurred while making containerd connection",
			"address", constants.ContainerdAddress)
		return err
	}
	logger.Info("kubernetes version during clean-up",
		"version", cri.GetImageVersionForCleanup(),
		"namespace", constants.ContainerdKubernetesNamespace)
	images, err := containerdClient.ListK8sControlplaneImages(ctx, cri.GetImageVersionForCleanup())
	if err != nil {
		logger.Error(err, "error occurred while listing containerd images",
			"namespace", constants.ContainerdKubernetesNamespace)
		return err
	}
	logger.Info("list of images to be cleaned-up",
		"images", images,
		"namespace", constants.ContainerdKubernetesNamespace)
	for _, image := range images {
		err := containerdClient.DeleteImage(ctx, image)
		if err != nil {
			logger.Error(err, "error occurred while deleting the image",
				"image", image,
				"namespace", constants.ContainerdKubernetesNamespace)
		}
	}
	return nil
}

func (t *CleanUpK8sControlPlaneContainerdImages) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
