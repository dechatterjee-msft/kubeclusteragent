package common

import (
	"context"
	"fmt"
	"go.uber.org/multierr"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"time"
)

type RestartK8sControlplane struct{}

var _ task.Task = &RestartK8sControlplane{}

func NewRestartK8sControlplane() *RestartK8sControlplane {
	t := &RestartK8sControlplane{}
	return t
}

func (t *RestartK8sControlplane) Name() string {
	return "restart-k8s-controlplane"
}

func (t *RestartK8sControlplane) Run(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("running k8s controlplane restart")
	code, _, err := ou.Exec().Command(ctx, "mv", nil, []string{constants.StaticPodManifests, constants.StaticPodManifestsBkp}...)
	if err != nil || code != 0 {
		currErr := fmt.Errorf("error occouered while configuring the static pods %d", code)
		err = multierr.Append(err, currErr)
		return err
	}
	// wait for kubelet to detect the change
	time.Sleep(30 * time.Second)
	code, _, err = ou.Exec().Command(ctx, "mv", nil, []string{constants.StaticPodManifestsBkp, constants.StaticPodManifests}...)
	if err != nil {
		if err != nil || code != 0 {
			currErr := fmt.Errorf("error occouered while configuring the static pods code:%d,error:%v", code, err)
			err = multierr.Append(err, currErr)
			return err
		}
		currErr := fmt.Errorf("error occouered while configuring the static pods code:%d,error:%v", code, err)
		err = multierr.Append(err, currErr)
		return err
	}
	// waiting for controlplane to come-up
	time.Sleep(20 * time.Second)
	return nil
}

func (t *RestartK8sControlplane) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
