package statusreconciler

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/multierr"
	"kubeclusteragent/pkg/util/conditions"
	"kubeclusteragent/pkg/util/heartbeat"
	"kubeclusteragent/pkg/util/k8s"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"os"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
)

const (
	ClusterStatusReconcilerName = "cluster-status-reconciler"
)

type ClusterStatusReconciler struct {
	stopped  chan struct{}
	quit     chan bool
	context  context.Context
	client   *kubernetes.Clientset
	interval time.Duration
	log      logr.Logger
	osUtil   osutility.OSUtil
}

func NewClusterStatusReconciler(ctx context.Context, kubeconfig string) (*ClusterStatusReconciler, error) {
	logger := log.From(ctx).WithName("reconciler").WithName("ClusterStatus")
	client, err := k8s.GetKubeClientFromKubeconfig(kubeconfig)
	if err != nil {
		logger.Error(err, "unable to get kubeconfig , new cluster status reconciler")
		return nil, err
	}

	return &ClusterStatusReconciler{
		context:  context.Background(),
		client:   client,
		interval: 10 * time.Second,
		log:      logger,
		stopped:  make(chan struct{}),
		quit:     make(chan bool),
		osUtil:   osutility.New(),
	}, nil
}

func (csr *ClusterStatusReconciler) Name() string {
	return ClusterStatusReconcilerName
}

func (csr *ClusterStatusReconciler) Reconcile(ctx context.Context) {
	csr.log.Info("Starting cluster status reconciler")
	go heartbeat.HeartBeat(ctx, csr.getClusterHeartbeat, csr.interval, csr.stopped, csr.quit)
}

func (csr *ClusterStatusReconciler) RunUntil(timeout time.Duration) {
	ctx, _ := context.WithTimeout(csr.context, timeout) // nolint
	go heartbeat.HeartBeat(ctx, csr.getClusterHeartbeat, csr.interval, csr.stopped, csr.quit)
}

// Stop stops the go routine and gets confirmation back via stop channel
func (csr *ClusterStatusReconciler) Stop() <-chan struct{} {
	go func() {
		csr.quit <- true
	}()
	return csr.stopped
}

func (csr *ClusterStatusReconciler) getClusterHeartbeat() error {
	status := &cluster.LiveStatus{}
	clusterStatus := status.GetStatus(csr.context)
	clusterSpec := status.GetSpec(csr.context)
	// if cluster is performing operation or is not present or deleted. Don't reconcile!
	if csr.inProgressState(clusterStatus) || csr.notPresentOrDeleted(clusterStatus) {
		return nil
	}

	defer status.SetStatus(csr.context, clusterStatus)

	// check if the controlplane has been restarted successfully,for the cert rotation case
	// if kubeconfig is not present
	if !csr.checkIfKubeConfigPresent() {
		conditions.DeleteAll(clusterStatus)
		conditions.MarkFalse(clusterStatus, v1alpha1.ConditionType_ClusterReady, constants.ClusterReadyStatusMessageFailed, constants.ConditionSeverityError, "Kubeconfig not present for cluster status")
		clusterStatus.Phase = constants.ClusterPhaseDelete
		csr.log.Info("Stopping the reconciler as cluster kubeconfig is not present")
		// soft stop! we don't wait for the reconciler to confirm its stopped status.
		// need to be very careful here, as it might cause deadlock.
		csr.Stop()
	}
	err := csr.isStaticPodReady()
	if err != nil {
		return err
	}

	nodeready, node, err := csr.getNodeStatus()
	if err != nil || node == nil {
		csr.log.Error(err, "Node status returned err")
		conditions.MarkFalse(clusterStatus, v1alpha1.ConditionType_NodeReady,
			constants.NodeReadyStatusMessageFailed,
			constants.ConditionSeverityError, "Node is not in ready status")
		return err
	}
	clusterStatus.Unschedulable = node.Spec.Unschedulable
	clusterStatus.KubernetesVersion = node.Status.NodeInfo.KubeletVersion
	if !nodeready {
		conditions.MarkFalse(clusterStatus, v1alpha1.ConditionType_NodeReady, constants.NodeReadyStatusMessageFailed, constants.ConditionSeverityError, "Node is not in ready status")
	} else {
		conditions.MarkTrue(clusterStatus, v1alpha1.ConditionType_NodeReady)
	}
	var cpReady bool
	// Triggering auto upgrade if kubeadm,kubelet have been upgraded offline without using kubecluster agent
	if clusterSpec.ClusterType == "kubeadm" {
		cpReady, err = csr.genericControlPlaneHeartBeatInfo()
		if err != nil {
			csr.log.Error(err, "Get ControlPlane status returned err")
			return err
		}
	}
	if !cpReady {
		conditions.MarkFalse(clusterStatus, v1alpha1.ConditionType_ControlPlaneReady, constants.ControlPlaneStatusMessageFailed, constants.ConditionSeverityError, "Control Plane is not in ready status")
	} else {
		// csr.log.Info("ControlPlane status is ready")
		conditions.MarkTrue(clusterStatus, v1alpha1.ConditionType_ControlPlaneReady)
	}
	if nodeready && cpReady {
		conditions.MarkTrue(clusterStatus, v1alpha1.ConditionType_ClusterReady)
	} else {
		conditions.MarkFalse(clusterStatus, v1alpha1.ConditionType_ClusterReady, constants.ClusterReadyStatusMessageFailed, constants.ConditionSeverityError, "Control Plane is not in ready status")
	}
	return nil
}

// isStaticPodReady this function will check for static pods(api-server,etcd,scheduler and controller-manager) readiness
func (csr *ClusterStatusReconciler) isStaticPodReady() error {
	exists, err := csr.osUtil.Filesystem().Exists(csr.context, constants.StaticPodManifestsBkp)
	if err != nil {
		return err
	}
	if exists {
		code, _, err := csr.osUtil.Exec().Command(csr.context, "mv", nil, []string{constants.StaticPodManifestsBkp, constants.StaticPodManifests}...)
		if err != nil || code != 0 {
			currErr := fmt.Errorf("error occuered while configuring the static pods code:%d,error:%v", code, err)
			err = multierr.Append(err, currErr)
			return err
		}
		csr.log.Info("wait for controlplane to come up")
		time.Sleep(40 * time.Second)
		file, err := os.ReadFile(constants.KubeadmKubeconfigPath)
		if err != nil {
			return err
		}
		csr.client, err = k8s.GetKubeClientFromKubeconfig(string(file))
		if err != nil {
			return err
		}
	}
	return nil
}

func (csr *ClusterStatusReconciler) getNodeStatus() (bool, *v1.Node, error) {
	retryCount := 0
nodeStatus:
	node, err := k8s.GetNode(csr.context, csr.client)
	if err != nil {
		err := csr.reinitializingKubeClient()
		if err == nil {
			retryCount++
			if retryCount < 5 {
				time.Sleep(10 * time.Second)
				csr.log.Info("retrying node status")
				goto nodeStatus
			}
		}
		csr.log.Error(err, "Get node status failed with error")
		return false, nil, err
	}
	if node == nil {
		err := errors.New("heart beat services expected a single node, so stopping with failure")
		csr.log.V(1).Error(err, "Heart beat services expected a single node, so stopping with failure")
		return false, nil, err
	}

	nodeReady := false
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			nodeReady = true
		}
	}
	return nodeReady, node, nil
}

func (csr *ClusterStatusReconciler) genericControlPlaneHeartBeatInfo() (bool, error) {
	controlPlanePods := []string{"kube-apiserver", "kube-proxy", "kube-scheduler", "etcd"}
	return k8s.GetKubeSystemPodStatus(csr.context, controlPlanePods, csr.client)
}

func (csr *ClusterStatusReconciler) inProgressState(status *v1alpha1.ClusterStatus) bool {
	switch status.GetPhase() {
	case constants.ClusterPhaseDeleting:
		return true
	case constants.ClusterPhaseProvisioning:
		return true
	case constants.ClusterPhaseKubeConfigResetting:
		return true
	case constants.ClusterPhaseUpgrading:
		return true

	default:
		return false
	}
}

func (csr *ClusterStatusReconciler) checkIfKubeConfigPresent() bool {
	if _, err := os.Stat(constants.KubeadmKubeconfigPath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (csr *ClusterStatusReconciler) notPresentOrDeleted(status *v1alpha1.ClusterStatus) bool {
	switch status.GetPhase() {
	case constants.ClusterPhaseNotInitialised:
		return true
	case constants.ClusterPhaseDelete:
		return true
	default:
		return false
	}
}

func (csr *ClusterStatusReconciler) reinitializingKubeClient() error {
	var kubeconfig string
	var err error
	kubeconfigBytes, err := os.ReadFile(constants.KubeadmKubeconfigPath)
	if err != nil {
		return err
	}
	kubeconfig = string(kubeconfigBytes)
	csr.client, err = k8s.GetKubeClientFromKubeconfig(kubeconfig)
	if err != nil {
		return err
	}
	return nil
}

func (csr *ClusterStatusReconciler) triggerAutoUpgrade() {
	//if clusterSpec.Version != node.Status.NodeInfo.KubeletVersion {
	//	csr.log.Info("cluster status reconciler detected cluster upgrade",
	//		"current-version", clusterSpec.Version,
	//		"desiered-version", node.Status.NodeInfo.KubeletVersion)
	//	var kubeToolFactory kubernetestoolsfactory.KubeToolsFactory = &kubernetestoolsfactory.KubeManager{}
	//	err = kubeToolFactory.KubernetesClusterUpgradeManager(csr.context, node.Status.NodeInfo.KubeletVersion)
	//	if err != nil {
	//		csr.log.Error(err, "upgrade failed with error,cluster will be automatically rolled back,"+
	//			"reconciler will try to upgrade in the next attempt")
	//	}
	//} else {
	//	clusterStatus.KubernetesVersion = node.Status.NodeInfo.KubeletVersion
	//}
}
