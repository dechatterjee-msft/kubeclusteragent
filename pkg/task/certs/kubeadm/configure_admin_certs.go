package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
)

type AdminCertsRotation struct{}

func (u AdminCertsRotation) Name() string {
	return "configure-admin-certs"
}

var _ task.Task = &AdminCertsRotation{}

func NewRotateAdminCerts() *AdminCertsRotation {
	t := &AdminCertsRotation{}
	return t
}

func (u AdminCertsRotation) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(u.Name())
	logger.Info("configuring admin kube-config certs")
	contents, err := ou.Filesystem().ReadFile(ctx, constants.KubeadmKubeconfigPath)
	if err != nil {
		return fmt.Errorf("error while getting kube-config contents: %w", err)
	}
	err = ou.Filesystem().WriteFile(ctx, constants.AdminKubeconfigPath, contents, constants.FileReadWriteAccess)
	if err != nil {
		return fmt.Errorf("eror while write kube-config contents to admin kube-config: %w", err)
	}
	return nil
}

func (u AdminCertsRotation) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}
