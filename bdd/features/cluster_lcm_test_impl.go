package features

import (
	"context"
	"kubeclusteragent/stf/clustertesttoolfacotry"
	"testing"
)

var clusterToolFactory clustertesttoolfacotry.ClusterTestToolFactoryInterface = &clustertesttoolfacotry.ClusterTest{}

func LcmOfSingleNodeKubeadmCluster(t *testing.T) error {
	clusterToolFactory.ClusterLifeCycleTest(context.Background(), "kubeadm", "1.23.8", "1.24.3", t)
	return nil
}
