package clustertesttoolfacotry

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/agent"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/reconciler/certsreconciler"
	"kubeclusteragent/pkg/reconciler/statusreconciler"
	"kubeclusteragent/pkg/tools/patchtool"
	"kubeclusteragent/pkg/util/auth"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/reconcile"
	"os"
	"testing"
	"time"
)

type ClusterLifeCycle struct{}

type ClusterLifeCycleTestManager interface {
	InstallCluster(ctx context.Context, version, clusterType string, t *testing.T) ClusterLifeCycle
	UpgradeCluster(ctx context.Context, version string, t *testing.T) ClusterLifeCycle
	DeleteCluster(ctx context.Context, t *testing.T) ClusterLifeCycle
	PatchCluster(ctx context.Context, t *testing.T) ClusterLifeCycle
	RotateCertificate(ctx context.Context, t *testing.T) ClusterLifeCycle
}

func (c ClusterLifeCycle) InstallCluster(ctx context.Context, version, clusterType string, t *testing.T) ClusterLifeCycle {
	logger := log.From(context.Background()).WithValues("stf", "install cluster")
	svc, clusterStatus := getClusterLifeCycleService(t)
	_, err := svc.CreateCluster(ctx, &v1alpha1.CreateClusterRequest{
		ApiVersion: "",
		Kind:       "",
		Spec: &v1alpha1.ClusterSpec{
			ClusterType: clusterType,
			Networking: &v1alpha1.ClusterNetworking{
				PodSubnet: "",
				SvcSubnet: "",
			},
			ApiServer: &v1alpha1.ClusterAPIServer{
				CertSANs: []string{"34.66.72.135"},
			},
			Version: version,
		},
	})
	if err != nil {
		logger.Error(err, "cluster creation failed with an error")
		t.Fail()
		os.Exit(1)
	}
	time.Sleep(3 * time.Minute)

	clusterPhase := clusterStatus.GetStatus(ctx).Phase
	switch clusterPhase {
	case constants.ClusterPhaseProvisioning:
		logger.Error(err, "unable to create cluster after 5 mins , stf will exit")
		t.Fail()
		os.Exit(1)
	case constants.ClusterPhaseProvisioned:
		logger.Info("cluster successfully created")
	default:
		logger.Error(err, "unable to create cluster  , stf will exit")
		t.Fail()
		os.Exit(1)
	}
	return c
}

func (c ClusterLifeCycle) UpgradeCluster(ctx context.Context, version string, t *testing.T) ClusterLifeCycle {
	logger := log.From(context.Background()).WithValues("stf", "upgrade cluster")
	svc, clusterStatus := getClusterLifeCycleService(t)
	_, err := svc.UpgradeCluster(ctx, &v1alpha1.UpgradeClusterRequest{
		ApiVersion: "",
		Kind:       "",
		Spec: &v1alpha1.ClusterSpec{
			ClusterType: "",
			Version:     version,
		},
	})
	if err != nil {
		logger.Error(err, "error occurred during upgrade")
		t.Fail()
		os.Exit(1)
	}
	time.Sleep(3 * time.Minute)
	clusterPhase := clusterStatus.GetStatus(ctx).Phase

	switch clusterPhase {
	case constants.ClusterPhaseUpgrading:
		logger.Error(err, "unable to upgrade cluster after 5 mins , stf will exit")
		t.Fail()
		os.Exit(1)
	case constants.ClusterPhaseProvisioned:
		logger.Info("cluster successfully upgraded")
	default:
		logger.Error(err, "unable to upgrade cluster  , stf will exit")
		t.Fail()
		os.Exit(1)
	}
	return c
}

func (c ClusterLifeCycle) DeleteCluster(ctx context.Context, t *testing.T) ClusterLifeCycle {
	logger := log.From(context.Background()).WithValues("stf", "delete cluster")
	svc, clusterStatus := getClusterLifeCycleService(t)
	_, err := svc.DeleteCluster(ctx)
	if err != nil {
		logger.Error(err, "error occurred during cluster delete")
		t.Fail()
		os.Exit(1)
	}
	time.Sleep(3 * time.Minute)
	clusterPhase := clusterStatus.GetStatus(ctx).Phase

	switch clusterPhase {
	case constants.ClusterPhaseDeleting:
		logger.Error(err, "unable to delete cluster after 5 mins , stf will exit")
		t.Fail()
		os.Exit(1)
	case constants.ClusterPhaseDelete:
		logger.Info("cluster successfully deleted")
	default:
		logger.Error(err, "unable to delete cluster  , stf will exit")
		t.Fail()
		os.Exit(1)
	}

	return c
}

func (c ClusterLifeCycle) PatchCluster(ctx context.Context, t *testing.T) ClusterLifeCycle {
	logger := log.From(context.Background()).WithValues("stf", "patch cluster")
	logger.Info("patching cluster")
	return c
}

func (c ClusterLifeCycle) RotateCertificate(ctx context.Context, t *testing.T) ClusterLifeCycle {
	logger := log.From(context.Background()).WithValues("stf", "rotate certs")
	logger.Info("certs rotation")
	return c
}

type ClusterTestToolFactoryInterface interface {
	ClusterLifeCycleTest(ctx context.Context, clusterType, version, upgradeVersion string, t *testing.T)
	ClusterReconcilationTest(ctx context.Context)
	ClusterSecureAccessTest(ctx context.Context)
}

type ClusterTest struct{}

func (c ClusterTest) ClusterLifeCycleTest(ctx context.Context, clusterType, version, upgradeVersion string, t *testing.T) {
	var clusterLCM ClusterLifeCycleTestManager = &ClusterLifeCycle{}
	clusterLCM.InstallCluster(context.Background(), version, clusterType, t).
		UpgradeCluster(context.Background(), upgradeVersion, t).PatchCluster(context.Background(), t).
		RotateCertificate(context.Background(), t).DeleteCluster(context.Background(), t)

}

func (c ClusterTest) ClusterReconcilationTest(ctx context.Context) {

}

func (c ClusterTest) ClusterSecureAccessTest(ctx context.Context) {

}

func getClusterLifeCycleService(t *testing.T) (*agent.LiveService, *cluster.LiveStatus) {
	var ctx context.Context
	logger := log.From(context.Background()).WithValues("stf", "cluster lcm service registration")
	clusterStatus, err := cluster.NewLiveStatus(ctx, false) // nolint
	if err != nil {
		logger.Error(err, "error occurred while initializing the status service , stf will exit")
		t.Fail()
		os.Exit(1)
	}
	reconcilerRegistery, err := reconcile.NewReconcileManager()
	if err != nil {
		logger.Error(err, "error occurred while launching the reconcile manager , stf will exit")
		t.Fail()
		os.Exit(1)
	}
	jwtManager := *auth.CreateJwtManager("")
	svc := agent.NewLiveService(ctx, patchtool.NewClusterConfigInstallTool(clusterStatus, false), jwtManager, reconcilerRegistery)
	if kc, err := svc.InstallTool.Config(ctx); err == nil {
		if svc.ReconcileRegistry.GetReconciler(statusreconciler.ClusterStatusReconcilerName) == nil && kc != nil {
			clusterStatusReconciler, err := statusreconciler.NewClusterStatusReconciler(ctx, string(kc))
			if err != nil {
				logger.Error(err, "error occurred while launching cluster status reconciliation , stf will exit")
				t.Fail()
				os.Exit(1)
			}
			clusterCertsReconciler, err := certsreconciler.NewCertificateReconciler(ctx)
			if err != nil {
				logger.Error(err, "error occurred while launching cluster certs reconciliation , stf will exit")
				t.Fail()
				os.Exit(1)
			}
			svc.ReconcileRegistry.Register(clusterStatusReconciler)
			svc.ReconcileRegistry.Register(clusterCertsReconciler)
		}
	}
	return svc, clusterStatus
}
