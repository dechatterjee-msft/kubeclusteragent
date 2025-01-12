// nolint
package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
)

type Cni struct{}

var _ task.Task = &Cni{}

func NewInstallCNI() *Cni {
	t := &Cni{}
	return t
}

func (t *Cni) Name() string {
	return "install-cni"
}

func (t *Cni) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("installing cni")
	//cniPath := fmt.Sprintf("%s/%s/%s_%s.yaml", constants.RootCNIPath, clusterSpec.Networking.ClusterCni.Name,
	//	clusterSpec.Networking.ClusterCni.Name,
	//	clusterSpec.Networking.ClusterCni.Version)
	//exists, err := ou.Filesystem().Exists(ctx, cniPath)
	//if err != nil || !exists {
	//	logger.Error(err, "cni doesn't exits in the given path or the path is not correct", "path", cniPath)
	//	return err
	//}
	response, err := ou.Kubectl().RunWithResponse(ctx, "apply", "-f", clusterSpec.Networking.CniManifestURL)
	if err != nil {
		logger.Error(err, "unable to apply CNI present in the give location", "path", clusterSpec.Networking.CniManifestURL)
		return err
	}
	logger.Info("cni installation response", "response", response)
	//if clusterSpec.Networking.ClusterCni.Name == "calico" {
	//	response, err := calico.ConfigurePodIPReservation(ctx, ou)
	//	if err != nil {
	//		logger.Error(err, "error occurred while applying calico IP reservation")
	//		return err
	//	}
	//	logger.Info("calico IP reservation installation response", "Response", response)
	//}

	return nil
}

func (t *Cni) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
