package agent

import (
	"context"

	"kubeclusteragent/gen/go/agent/v1alpha1"
)

type Server struct {
	v1alpha1.UnimplementedAgentAPIServer

	service Service
}

var _ v1alpha1.AgentAPIServer = &Server{}

func NewServer(service Service) *Server {
	server := &Server{
		service: service,
	}

	return server
}

func (s Server) GetCluster(ctx context.Context, _ *v1alpha1.GetClusterRequest) (*v1alpha1.Cluster, error) {
	return s.service.GetCluster(ctx)
}
func (s Server) CreateCluster(ctx context.Context, request *v1alpha1.CreateClusterRequest) (*v1alpha1.Cluster, error) {
	return s.service.CreateCluster(ctx, request)
}
func (s Server) UpgradeCluster(ctx context.Context, request *v1alpha1.UpgradeClusterRequest) (*v1alpha1.Cluster, error) {
	return s.service.UpgradeCluster(ctx, request)
}

func (s Server) PatchCluster(ctx context.Context, request *v1alpha1.PatchClusterRequest) (*v1alpha1.Cluster, error) {
	return s.service.PatchCluster(ctx, request)
}

func (s Server) DeleteCluster(ctx context.Context, _ *v1alpha1.DeleteClusterRequest) (*v1alpha1.Cluster, error) {
	return s.service.DeleteCluster(ctx)
}

func (s Server) GetKubeconfig(ctx context.Context, _ *v1alpha1.GetKubeconfigRequest) (*v1alpha1.Kubeconfig, error) {
	return s.service.GetKubeConfig(ctx)
}

func (s Server) ResetCerts(ctx context.Context, _ *v1alpha1.ResetKubeconfigRequest) (*v1alpha1.ResetKubeconfigRequest, error) {
	return s.service.ResetCerts(ctx)
}

func (s Server) GetCerts(ctx context.Context, _ *v1alpha1.ClusterCertificateRequest) (*v1alpha1.ClusterCertificatesResponse, error) {
	return s.service.GetCerts(ctx)
}

func (s Server) AuditHistory(ctx context.Context, request *v1alpha1.AuditHistoryRequest) (*v1alpha1.AuditHistoryResponse, error) {
	return s.service.Audit(ctx)
}

func (s Server) GetReconcilerRequest(ctx context.Context, _ *v1alpha1.GetClusterStatusReconcilerRequest) (*v1alpha1.GetClusterStatusReconcilerResponse, error) {
	return s.service.ReconcilerStatus(ctx)
}
