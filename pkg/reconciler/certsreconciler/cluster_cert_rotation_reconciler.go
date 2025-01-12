package certsreconciler

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/tools/kubernetestoolsfactory/kubernetesproviders/kubeadm"
	"kubeclusteragent/pkg/util/heartbeat"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"time"
)

const (
	ClusterCertsReconcilerName     = "cluster-certs-reconciler"
	DefaultClusterCertRotationDays = constants.ClusterCertsRotationDays
)

type ClusterCertsReconciler struct {
	stopped  chan struct{}
	quit     chan bool
	context  context.Context
	interval time.Duration
	log      logr.Logger
	osutil   osutility.OSUtil
}

func NewCertificateReconciler(ctx context.Context) (*ClusterCertsReconciler, error) {
	logger := log.From(ctx).WithName("reconciler").WithName("CertRotation")

	return &ClusterCertsReconciler{
		context:  ctx,
		interval: 10 * time.Hour,
		log:      logger,
		stopped:  make(chan struct{}),
		quit:     make(chan bool),
		osutil:   osutility.New(),
	}, nil
}

func (ccr *ClusterCertsReconciler) Name() string {
	return ClusterCertsReconcilerName
}

func (ccr *ClusterCertsReconciler) Reconcile(ctx context.Context) {
	ccr.log.Info("Starting cluster certificate reconciler")
	go heartbeat.HeartBeatWithCtx(ctx, ccr.reconcileCertificateExpiry, ccr.interval, ccr.stopped, ccr.quit)
}

// Stop stops the go routine and gets confirmation back via stop channel
func (ccr *ClusterCertsReconciler) Stop() <-chan struct{} {
	go func() {
		ccr.quit <- true
	}()
	return ccr.stopped
}

func (ccr *ClusterCertsReconciler) validateAllCertsRotated(ctx context.Context) bool {
	expiry, allCertsExpiry, err := ccr.osutil.Kubeadm().GetCertsExpiry(ctx)
	if err != nil {
		ccr.log.Error(err, "unable to validate the expiry of the certs")
		return false
	}
	if allCertsExpiry == nil {
		return false
	}
	for k, v := range allCertsExpiry {
		if expiry != int(v) {
			ccr.log.Error(fmt.Errorf("partial certificate rotation error"), "certificate yet to be rotated", "Cert", k, "Value", v)
		}
	}
	return true
}

func (ccr *ClusterCertsReconciler) reconcileCertificateExpiry(ctx context.Context) error {
	status := &cluster.LiveStatus{}
	expiry, _, err := ccr.osutil.Kubeadm().GetCertsExpiry(ctx)
	if err != nil {
		return err
	}
	kubeadmTool := kubeadm.NewKubeadmInstallTool(status, false)
	if expiry <= DefaultClusterCertRotationDays {
		err = kubeadmTool.ResetConfig(ctx)
		if err != nil {
			return err
		}
		if ccr.validateAllCertsRotated(ctx) {
			ccr.log.Info("all certificates rotated successfully")
		}
		expiry, _, err = ccr.osutil.Kubeadm().GetCertsExpiry(ctx)
		if err != nil {
			return err
		}
	}
	deadline := expiry - DefaultClusterCertRotationDays
	ccr.log.Info("No of days to certificate rotation", "Deadline", deadline)
	return nil
}
