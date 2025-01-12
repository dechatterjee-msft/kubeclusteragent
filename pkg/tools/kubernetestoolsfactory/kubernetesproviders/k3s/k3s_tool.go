package k3s

import (
	"context"
	"errors"
	"go.uber.org/multierr"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/operations"
	kubernetestool "kubeclusteragent/pkg/tools/kubernetestoolsfactory/kubernetesproviders"
)

var (
	defaultPodNetwork     = "10.42.0.0/16"
	defaultServiceNetwork = "10.43.0.0/16"
)

type K3sTool struct {
	//kubeTool      kubernetesproviders.KubernetesProviderFactory
	clusterStatus cluster.Status
	dryRun        bool
	//metricsTool   metricstool.PrometheusMetricsTool
}

var _ kubernetestool.KubernetesProviderFactory = &K3sTool{}
var defaultKubernetesTool kubernetestool.KubernetesProviderFactory = &kubernetestool.DefaultKubernetesProvider{}

func NewK3sInstallTool(clusterStatus cluster.Status, dryRun bool) *K3sTool {
	t := &K3sTool{
		clusterStatus: clusterStatus,
		dryRun:        dryRun,
	}
	return t
}

func (t *K3sTool) IsInitialized(ctx context.Context) bool {
	clusterStatus := t.clusterStatus.GetStatus(ctx)
	if clusterStatus == nil {
		return false
	}
	return clusterStatus.Phase != constants.ClusterPhaseNotInitialised &&
		clusterStatus.Phase != constants.ClusterPhaseDelete &&
		clusterStatus.Phase != constants.ClusterPhaseFailed
}

func (t *K3sTool) Install(ctx context.Context, request *v1alpha1.CreateClusterRequest) error {
	var options []operations.Option
	if t.dryRun {
		options = append(options, operations.DryRun())
	}
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus:       t.clusterStatus,
		Tasks:               buildInstallOptions(options...),
		SpecValidationError: t.validateSpec(request.Spec),
	}
	err := defaultKubernetesTool.Install(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (t *K3sTool) Reset(ctx context.Context) error {
	var options []operations.Option
	if t.dryRun {
		options = append(options, operations.DryRun())
	}
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
	}
	err := defaultKubernetesTool.Reset(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (t *K3sTool) Cluster(ctx context.Context) (*v1alpha1.Cluster, error) {
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
	}
	return defaultKubernetesTool.Cluster(ctx)
}

func (t *K3sTool) Config(ctx context.Context) ([]byte, error) {
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
	}
	return defaultKubernetesTool.Config(ctx)
}

func (t *K3sTool) ResetConfig(ctx context.Context) error {
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
	}
	return defaultKubernetesTool.ResetConfig(ctx)
}

func (t *K3sTool) Upgrade(ctx context.Context, request *v1alpha1.UpgradeClusterRequest) error {
	var options []operations.Option
	if t.dryRun {
		options = append(options, operations.DryRun())
	}
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus:       t.clusterStatus,
		Tasks:               buildUpgradeOptions(options...),
		SpecValidationError: t.validateSpec(request.Spec),
	}
	err := defaultKubernetesTool.Upgrade(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (t *K3sTool) GetCerts(ctx context.Context) (*v1alpha1.ClusterCertificatesResponse, error) {
	return nil, nil
}

func (t *K3sTool) validateSpec(spec *v1alpha1.ClusterSpec) error {
	var err error
	if spec.Version == "" {
		spec.Version = "latest"
	}
	if spec.Networking == nil {
		spec.Networking = new(v1alpha1.ClusterNetworking)
	}
	if spec.Networking.PodSubnet == "" {
		spec.Networking.PodSubnet = defaultPodNetwork
	}
	if spec.Networking.SvcSubnet == "" {
		spec.Networking.SvcSubnet = defaultServiceNetwork
	}
	if spec.DisableWorkloads == nil {
		disableWorkload := false
		spec.DisableWorkloads = &disableWorkload
	}

	if spec.Version == "" {
		err = multierr.Append(err, errors.New("no K3s version found"))
	}
	return err
}
