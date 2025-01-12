// nolint
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
)

type Csi struct{}

var _ task.Task = &Csi{}

func NewInstallCSI() *Csi {
	t := &Csi{}
	return t
}

func (t *Csi) Name() string {
	return "install-csi"
}

func (t *Csi) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("installing csi", "CSI", clusterSpec.Storage.ClusterCsi.Name, "Version", clusterSpec.Storage.ClusterCsi.Version)
	csiPath := fmt.Sprintf("%s/%s/%s_%s.yaml", constants.RootCSIPath, clusterSpec.Storage.ClusterCsi.Name,
		clusterSpec.Storage.ClusterCsi.Name,
		clusterSpec.Storage.ClusterCsi.Version)
	exists, err := ou.Filesystem().Exists(ctx, csiPath)
	if err != nil || !exists {
		logger.Error(err, "csi doesn't exits in the given path or the path is not correct", "path", csiPath)
		return err
	}
	response, err := ou.Kubectl().RunWithResponse(ctx, "apply", "-f", csiPath)
	if err != nil {
		logger.Error(err, "unable to apply CSI present in the give location", "path", csiPath,
			"Name", clusterSpec.Storage.ClusterCsi.Name,
			"Version", clusterSpec.Storage.ClusterCsi.Version)
		return err
	}
	logger.Info("csi installation response", "response", response, "CSI", clusterSpec.Storage.ClusterCsi.Name, "Version", clusterSpec.Storage.ClusterCsi.Version)
	return nil
}

func (t *Csi) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
