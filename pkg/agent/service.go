package agent

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/reconciler/certsreconciler"
	"kubeclusteragent/pkg/tools/kubernetestoolsfactory"
	"kubeclusteragent/pkg/tools/kubernetestoolsfactory/kubernetesproviders"
	"kubeclusteragent/pkg/tools/metricstool"
	"kubeclusteragent/pkg/tools/patchtool"
	"kubeclusteragent/pkg/util/auth"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"kubeclusteragent/pkg/util/reconcile"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/reconciler/statusreconciler"
)

type Service interface {
	GetCluster(ctx context.Context) (*v1alpha1.Cluster, error)
	CreateCluster(ctx context.Context, request *v1alpha1.CreateClusterRequest) (*v1alpha1.Cluster, error)
	DeleteCluster(ctx context.Context) (*v1alpha1.Cluster, error)
	GetKubeConfig(ctx context.Context) (*v1alpha1.Kubeconfig, error)
	ResetCerts(ctx context.Context) (*v1alpha1.ResetKubeconfigRequest, error)
	GetCerts(ctx context.Context) (*v1alpha1.ClusterCertificatesResponse, error)
	PatchCluster(ctx context.Context, request *v1alpha1.PatchClusterRequest) (*v1alpha1.Cluster, error)
	UpgradeCluster(ctx context.Context, request *v1alpha1.UpgradeClusterRequest) (*v1alpha1.Cluster, error)
	Audit(ctx context.Context) (*v1alpha1.AuditHistoryResponse, error)
	ReconcilerStatus(ctx context.Context) (*v1alpha1.GetClusterStatusReconcilerResponse, error)
	ReconcilerStop(ctx context.Context)
	ReconcilerStart(ctx context.Context)
}

type LiveService struct {
	InstallTool       kubernetesproviders.KubernetesProviderFactory
	jwtManager        auth.JWTManager
	patchTool         patchtool.ClusterConfigurationChange
	metricsTool       metricstool.PrometheusMetricsTool
	ReconcileRegistry reconcile.ReconcilerRegistry
}

var _ Service = &LiveService{}
var kubeToolFactory kubernetestoolsfactory.KubeToolsFactory = &kubernetestoolsfactory.KubeManager{}

// NewLiveService Let the live service to decide which all k8s service need to be initialized
func NewLiveService(ctx context.Context, patchTool patchtool.ClusterConfigurationChange, jwtManager auth.JWTManager, registry reconcile.ReconcilerRegistry) *LiveService {
	// Get Current cluster spec and check cluster type , default will be kubeadm tool
	s := &LiveService{
		InstallTool:       kubeToolFactory.GetKubernetesProviderOnStartup(ctx),
		patchTool:         patchTool,
		jwtManager:        jwtManager,
		ReconcileRegistry: registry,
	}
	return s
}

func (s *LiveService) GetCluster(ctx context.Context) (*v1alpha1.Cluster, error) {
	cluster, err := s.InstallTool.Cluster(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return cluster, nil
}

func (s *LiveService) CreateCluster(
	ctx context.Context,
	request *v1alpha1.CreateClusterRequest) (*v1alpha1.Cluster, error) {
	validationResponse := s.createClusterRequestPreValidation(request.Spec)
	if validationResponse != nil {
		return nil, validationResponse
	}

	s.installToolGenerator(ctx, request.Spec)
	if err := s.InstallTool.Install(ctx, request); err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	cluster, err := s.InstallTool.Cluster(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	// register reconciler if not already done
	go func() {
		logger := log.From(ctx).WithName("service").WithName("create-cluster").WithName("RegisterReconciliations")
		registerReconciler := func() {
			cl, err := s.InstallTool.Cluster(ctx)
			if err != nil {
				logger.Error(err, "Error while getting cluster information")
				return
			}
			for cl.Status.Phase == constants.ClusterPhaseProvisioning {
				time.Sleep(10 * time.Second)
				cl, err = s.InstallTool.Cluster(ctx)
				if err != nil {
					logger.Error(err, "Error while getting cluster information")
					return
				}
			}
			if cl.Status.Phase == constants.ClusterPhaseFailed {
				return
			}
			if kc, err := s.InstallTool.Config(ctx); err == nil {
				clusterStatusReconciler, err := statusreconciler.NewClusterStatusReconciler(ctx, string(kc))
				if err != nil {
					logger.Error(err, "Error while register status reconciler")
				}
				certificateRotationReconciler, err := certsreconciler.NewCertificateReconciler(ctx)
				if err != nil {
					logger.Error(err, "Error while register certificate reconciler")
				}
				s.ReconcileRegistry.Register(clusterStatusReconciler)
				s.ReconcileRegistry.Register(certificateRotationReconciler)
				logger.Info("Cluster Status Reconciler registered")
				logger.Info("Cluster Certs Reconciler registered")
				return
			}
		}
		registerReconciler()
	}()
	return cluster, nil
}

func (s *LiveService) GetCerts(ctx context.Context) (*v1alpha1.ClusterCertificatesResponse, error) {
	return s.InstallTool.GetCerts(ctx)
}

func (s *LiveService) DeleteCluster(ctx context.Context) (*v1alpha1.Cluster, error) {
	if err := s.InstallTool.Reset(ctx); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	cl, err := s.InstallTool.Cluster(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	// Unregister the status reconciler
	go func() {
		logger := log.From(ctx).WithName("service").WithName("delete-cluster").WithName("RegisterReconciliations")
		registerStatusReconciler := func() {
			cl, err := s.InstallTool.Cluster(ctx)
			if err != nil {
				logger.Error(err, "Error while getting cluster information")
				return
			}
			for cl.Status.Phase == constants.ClusterPhaseDeleting {
				time.Sleep(10 * time.Second)
				cl, err = s.InstallTool.Cluster(ctx)
				if err != nil {
					logger.Error(err, "Error while getting cluster information")
					return
				}
			}
			if cl.Status.Phase == constants.ClusterPhaseFailed {
				return
			}
			s.ReconcileRegistry.UnRegister(statusreconciler.ClusterStatusReconcilerName)
		}
		registerStatusReconciler()
	}()
	s.metricsTool.MetricsLabels = []string{codes.Unknown.String(), "DELETE", "api/v1alpha1/cluster"}
	return cl, nil
}

func (s *LiveService) UpgradeCluster(ctx context.Context, request *v1alpha1.UpgradeClusterRequest) (*v1alpha1.Cluster, error) {
	if err := s.InstallTool.Upgrade(ctx, request); err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	clusterInfo, err := s.InstallTool.Cluster(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return clusterInfo, nil
}

func (s *LiveService) PatchCluster(ctx context.Context, request *v1alpha1.PatchClusterRequest) (*v1alpha1.Cluster, error) {
	if err := s.patchTool.Patch(ctx, request); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	clusterInfo, err := s.InstallTool.Cluster(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return clusterInfo, nil
}

func (s *LiveService) GetKubeConfig(ctx context.Context) (*v1alpha1.Kubeconfig, error) {
	config, err := s.InstallTool.Config(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	kubeconfig := &v1alpha1.Kubeconfig{
		Contents: string(config),
	}

	return kubeconfig, nil
}

func (s *LiveService) ResetCerts(ctx context.Context) (*v1alpha1.ResetKubeconfigRequest, error) {
	if err := s.InstallTool.ResetConfig(ctx); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	_, err := s.GetKubeConfig(ctx)
	// When kubeconfig changes, refresh the status reconciler
	go func() {
		logger := log.From(ctx).WithName("service").WithName("reset-cluster").WithName("refresh-status-reconciliation")
		refreshStatusReconciler := func() {
			cl, err := s.InstallTool.Cluster(ctx)
			if err != nil {
				logger.Error(err, "Error while getting cluster information")
				return
			}
			if cl.GetStatus().GetPhase() != constants.ClusterPhaseProvisioned {
				return
			}
			s.ReconcileRegistry.UnRegister(statusreconciler.ClusterStatusReconcilerName)
			if kc, err := s.InstallTool.Config(ctx); err == nil {
				if s.ReconcileRegistry.GetReconciler(statusreconciler.ClusterStatusReconcilerName) == nil && kc != nil {
					clusterStatusReconciler, err := statusreconciler.NewClusterStatusReconciler(ctx, string(kc))
					if err != nil {
						logger.Error(err, "Error while register status reconciler")
					}
					s.ReconcileRegistry.Register(clusterStatusReconciler)
					logger.Info("Cluster Status Reconciler registered")
				}
				return
			}
		}
		refreshStatusReconciler()
	}()
	return &v1alpha1.ResetKubeconfigRequest{}, err
}

func (s *LiveService) Audit(ctx context.Context) (*v1alpha1.AuditHistoryResponse, error) {
	auditHistory, err := cluster.GetAuditLogs(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if auditHistory == nil {
		return nil, fmt.Errorf("audit history not found")
	}
	return &v1alpha1.AuditHistoryResponse{
		Operations: auditHistory,
	}, nil
}

func (s *LiveService) installToolGenerator(ctx context.Context, requestClusterSpec *v1alpha1.ClusterSpec) {
	clusterInfo := cluster.LiveStatus{}
	clusterStatus := clusterInfo.GetStatus(ctx)
	if clusterStatus != nil {
		if clusterStatus.Phase == constants.ClusterPhaseNotInitialised {
			s.InstallTool = kubeToolFactory.GetKubernetesToolByProvider(ctx, requestClusterSpec.ClusterType)
		}
	}
}

func (s *LiveService) createClusterRequestPreValidation(clusterSpec *v1alpha1.ClusterSpec) error {
	if clusterSpec == nil {
		return status.Error(codes.InvalidArgument, "request spec cannot be empty")
	}
	if clusterSpec.ClusterType == "" {
		return status.Error(codes.InvalidArgument, "cluster type cannot be empty")
	}
	if clusterSpec.ClusterName == "" {
		var hostUtil linux.Host = &linux.LiveHost{}
		hostname, err := hostUtil.GetHostname()
		if err != nil {
			clusterSpec.ClusterName = "kubernetes"
		}
		clusterSpec.ClusterName = hostname
	}
	return nil
}

func (s *LiveService) ReconcilerStatus(ctx context.Context) (*v1alpha1.GetClusterStatusReconcilerResponse, error) {
	return &v1alpha1.GetClusterStatusReconcilerResponse{
		Reconciler: &v1alpha1.Reconciler{
			Name:   s.ReconcileRegistry.GetReconciler("cluster-status-reconciler").Name(),
			Status: "OK",
		},
	}, nil
}

func (s *LiveService) ReconcilerStop(ctx context.Context) {}

func (s *LiveService) ReconcilerStart(ctx context.Context) {}
