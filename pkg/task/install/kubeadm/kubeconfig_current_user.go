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
	"os/user"
)

type CurrentUserKubeconfig struct{}

var _ task.Task = &CurrentUserKubeconfig{}

func NewCurrentUserKubeconfig() *CurrentUserKubeconfig {
	t := &CurrentUserKubeconfig{}
	return t
}

func (t CurrentUserKubeconfig) Name() string {
	return "current-user-kubeconfig"
}

func (t CurrentUserKubeconfig) Run(ctx context.Context, status cluster.Status, clusterSpec *v1alpha1.ClusterSpec, ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	currentUser, err := user.Current()
	if err != nil {
		logger.Error(err, "failed to get current user")
		return err
	}
	err = ou.Filesystem().MkdirAll(ctx, fmt.Sprintf("%s/.kube", currentUser.HomeDir), constants.DirPerm)
	code, _, err := ou.Exec().Command(ctx, "cp", nil, []string{constants.KubeadmKubeconfigPath, fmt.Sprintf("%s/.kube/config", currentUser.HomeDir)}...)
	if err != nil {
		return err
	}
	if code != 0 {
		logger.Error(err, "failed to copy the current user's kubeconfig")
	}
	return nil
}

func (t CurrentUserKubeconfig) Rollback(ctx context.Context, status cluster.Status, clusterSpec *v1alpha1.ClusterSpec, ou osutility.OSUtil) error {
	return nil
}
