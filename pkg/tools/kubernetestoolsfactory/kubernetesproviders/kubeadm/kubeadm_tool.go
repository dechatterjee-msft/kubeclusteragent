package kubeadm

import (
	"context"
	"errors"
	"go.uber.org/multierr"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	kubernetestool "kubeclusteragent/pkg/tools/kubernetestoolsfactory/kubernetesproviders"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/operations"
)

type KubeadmTool struct {
	clusterStatus cluster.Status
	dryRun        bool
}

var _ kubernetestool.KubernetesProviderFactory = &KubeadmTool{}
var defaultKubernetesTool kubernetestool.KubernetesProviderFactory = &kubernetestool.DefaultKubernetesProvider{}

func NewKubeadmInstallTool(clusterStatus cluster.Status, dryRun bool) *KubeadmTool {
	t := &KubeadmTool{
		clusterStatus: clusterStatus,
		dryRun:        dryRun,
	}
	return t
}

func (t *KubeadmTool) IsInitialized(ctx context.Context) bool {
	clusterStatus := t.clusterStatus.GetStatus(ctx)
	if clusterStatus == nil {
		return false
	}
	return clusterStatus.Phase != constants.ClusterPhaseNotInitialised &&
		clusterStatus.Phase != constants.ClusterPhaseDelete &&
		clusterStatus.Phase != constants.ClusterPhaseFailed

}

func (t *KubeadmTool) Install(ctx context.Context, request *v1alpha1.CreateClusterRequest) error {
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

func (t *KubeadmTool) Upgrade(ctx context.Context, request *v1alpha1.UpgradeClusterRequest) error {
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

func (t *KubeadmTool) Reset(ctx context.Context) error {
	var options []operations.Option
	if t.dryRun {
		options = append(options, operations.DryRun())
	}
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
		Tasks:         buildResetOptions(options...),
	}
	err := defaultKubernetesTool.Reset(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (t *KubeadmTool) Cluster(ctx context.Context) (*v1alpha1.Cluster, error) {
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
	}
	return defaultKubernetesTool.Cluster(ctx)
}

func (t *KubeadmTool) Config(ctx context.Context) ([]byte, error) {
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
	}
	return defaultKubernetesTool.Config(ctx)
}

func (t *KubeadmTool) ResetConfig(ctx context.Context) error {
	var options []operations.Option
	if t.dryRun {
		options = append(options, operations.DryRun())
	}
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
		Tasks:         buildCertsRotationOptions(options...),
	}
	err := defaultKubernetesTool.ResetConfig(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (t *KubeadmTool) GetCerts(ctx context.Context) (*v1alpha1.ClusterCertificatesResponse, error) {
	defaultKubernetesTool = &kubernetestool.DefaultKubernetesProvider{
		ClusterStatus: t.clusterStatus,
	}
	return defaultKubernetesTool.GetCerts(ctx)
}

func (t *KubeadmTool) validateSpec(spec *v1alpha1.ClusterSpec) error {
	var err error
	if spec.Networking == nil {
		err = multierr.Append(err, errors.New("networking object is nil"))
	} else {
		if spec.Networking.PodSubnet == "" {
			spec.Networking.PodSubnet = constants.DefaultPodNetwork
		}
		if spec.Networking.SvcSubnet == "" {
			spec.Networking.SvcSubnet = constants.DefaultServiceNetwork
		}
		if spec.DisableWorkloads == nil {
			disableWorkload := false
			spec.DisableWorkloads = &disableWorkload
		}
		if spec.Networking.CniManifestURL == "" {
			err = multierr.Append(err, errors.New("cni details is missing cni name and version is mandatory"))
		}
	}
	if spec.Version == "" {
		// get k8s version form the distro available in the VM
		err = multierr.Append(err, errors.New("no Kubernetes version"))
	}
	if spec.ClusterRuntime != nil &&
		spec.ClusterRuntime.CustomiseClusterRuntime {
		if spec.ClusterRuntime.ClusterCri.PrivateRegistryFQDN == "" {
			err = multierr.Append(err, errors.New("cluster runtime is true, so private registry FQDN is mandatory"))
		}
		if spec.ClusterRuntime.ClusterCri.RegistryAuth.IsAuthRequired {
			if spec.ClusterRuntime.ClusterCri.RegistryAuth.Username == "" ||
				spec.ClusterRuntime.ClusterCri.RegistryAuth.Password == "" {
				err = multierr.Append(err, errors.New("cluster runtime private registry authentication is true , then username and password is mandatory"))
			}
		}
	}
	return err
}
