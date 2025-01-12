package kubernetesproviders

import (
	"context"
	"errors"
	"fmt"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/operations"
	"kubeclusteragent/pkg/tools/metricstool"
	"kubeclusteragent/pkg/util/conditions"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/metrcis"
	"kubeclusteragent/pkg/util/osutility"
	"os"
	"time"
)

type DefaultKubernetesProvider struct {
	ClusterStatus       cluster.Status
	dryRun              bool
	metricsTool         metricstool.PrometheusMetricsTool
	Tasks               operations.TaskDetails
	SpecValidationError error
}

type KubernetesProviderFactory interface {
	IsInitialized(ctx context.Context) bool
	Install(ctx context.Context, request *v1alpha1.CreateClusterRequest) error
	Reset(ctx context.Context) error
	Cluster(ctx context.Context) (*v1alpha1.Cluster, error)
	Config(context.Context) ([]byte, error)
	ResetConfig(context.Context) error
	Upgrade(ctx context.Context, request *v1alpha1.UpgradeClusterRequest) error
	GetCerts(ctx context.Context) (*v1alpha1.ClusterCertificatesResponse, error)
}

func (t *DefaultKubernetesProvider) ExecutionInProgress(ctx context.Context) bool {
	clusterStatus := t.ClusterStatus.GetStatus(ctx)
	switch clusterStatus.Phase {
	case constants.ClusterPhaseProvisioning:
		return true
	case constants.ClusterPhaseKubeConfigResetting:
		return true
	case constants.ClusterPhaseDeleting:
		return true
	}
	return false
}

func NewDefaultKubernetesInstallTool(clusterStatus cluster.Status, dryRun bool) *DefaultKubernetesProvider {
	t := &DefaultKubernetesProvider{
		ClusterStatus: clusterStatus,
		dryRun:        dryRun,
	}
	return t
}

func (t *DefaultKubernetesProvider) IsInitializedForUpgrade(ctx context.Context) bool {
	clusterStatus := t.ClusterStatus.GetStatus(ctx)
	if clusterStatus == nil {
		return false
	}
	return clusterStatus.Phase == constants.ClusterPhaseProvisioned || clusterStatus.Phase == constants.ClusterPhaseFailed
}

func (t *DefaultKubernetesProvider) IsInitialized(ctx context.Context) bool {
	clusterStatus := t.ClusterStatus.GetStatus(ctx)
	if clusterStatus == nil {
		return false
	}
	// return clusterStatus.Phase != constants.ClusterPhaseNotInitialised && clusterStatus.Phase != constants.ClusterPhaseDelete
	return clusterStatus.Phase == constants.ClusterPhaseProvisioned
}

func (t *DefaultKubernetesProvider) Install(ctx context.Context, request *v1alpha1.CreateClusterRequest) error {
	var metricsResponseCode, auditMessage, auditReason, status string
	var startTime time.Time
	clusterStatus := t.ClusterStatus.GetStatus(ctx)
	if request.Spec.ClusterType == "" {
		request.Spec.ClusterType = constants.DefaultKubernetesTool
	}
	defer func() {
		t.metricsTool.MetricsLabels = []string{request.Spec.ClusterType, request.Spec.Version, metricsResponseCode, "POST", "api/v1alpha1/cluster"}
		t.metricsTool.PopulateToPrometheusMetrics(startTime)
		cluster.SetAuditLog(ctx, "Install", request.Spec.ClusterType, request.Spec.Version, status, auditMessage, auditReason)
		t.ClusterStatus.SetStatus(ctx, clusterStatus)
	}()
	if t.IsInitialized(ctx) {
		auditMessage = "Cluster is already initialized"
		return fmt.Errorf("cluster already initialized")
	}
	auditMessage = "Cluster installation is in progress"
	clusterStatus.Phase = constants.ClusterPhaseProvisioning

	if err := t.SpecValidationError; err != nil {
		auditMessage = "Cluster validation failed"

		metricsResponseCode = metrcis.ClusterAlreadyInitialized
		return fmt.Errorf("validate cluster spec: %w", err)
	}
	t.ClusterStatus.SetSpec(ctx, request.Spec)
	go func() {
		defer func() {
			t.metricsTool.MetricsLabels = []string{request.Spec.ClusterType, request.Spec.Version, metricsResponseCode, "POST", "api/v1alpha1/cluster"}
			t.metricsTool.PopulateToPrometheusMetrics(startTime)
			t.ClusterStatus.SetStatus(ctx, clusterStatus)
			cluster.SetAuditLog(ctx, "Install", request.Spec.ClusterType, request.Spec.Version, clusterStatus.GetPhase(), auditMessage, auditReason)
		}()
		logger := log.From(ctx)
		taskDetails := t.Tasks
		installer := operations.NewOperation("install cluster", t.ClusterStatus, request.Spec, taskDetails)
		if err := installer.Run(ctx); err != nil {
			auditMessage = "Cluster installation failed"
			auditReason = err.Error()
			logger.Error(err, "cluster failed while installation")
			clusterStatus.Phase = constants.ClusterPhaseFailed
			metricsResponseCode = metrcis.ClusterFailed
			conditions.MarkFalse(clusterStatus,
				v1alpha1.ConditionType_InstallSuccess,
				constants.InstallReadyStatusMessageFailed,
				constants.ConditionSeverityError, auditMessage)
			logger.Error(err, "Unable to complete install")
		} else {
			auditMessage = "Cluster has been successfully installed"
			clusterStatus.Phase = constants.ClusterPhaseProvisioned
			metricsResponseCode = metrcis.ClusterCreated
		}

	}()

	return nil
}

func (t *DefaultKubernetesProvider) Reset(ctx context.Context) error {
	var metricsResponseCode, auditMessage, auditReason, status string
	var startTime time.Time
	var clusterStatus = t.ClusterStatus.GetStatus(ctx)
	defer func() {
		t.metricsTool.MetricsLabels = []string{t.ClusterStatus.GetSpec(ctx).ClusterType, t.ClusterStatus.GetSpec(ctx).Version, metricsResponseCode, "DELETE", "api/v1alpha1/cluster"}
		t.metricsTool.PopulateToPrometheusMetrics(startTime)
		t.ClusterStatus.SetStatus(ctx, clusterStatus)
		cluster.SetAuditLog(ctx, "Reset", t.ClusterStatus.GetSpec(ctx).ClusterType, t.ClusterStatus.GetSpec(ctx).Version, status, auditMessage, auditReason)
	}()
	if clusterStatus != nil && clusterStatus.Phase == constants.ClusterPhaseNotInitialised {
		auditMessage = "cluster is not initialized,cannot perform delete operation"
		return fmt.Errorf("cluster is not initialized,cannot perform delete operation")
	}
	auditMessage = "Cluster is getting sundown"
	clusterStatus.Phase = constants.ClusterPhaseDeleting

	go func() {
		defer func() {
			t.metricsTool.MetricsLabels = []string{t.ClusterStatus.GetSpec(ctx).ClusterType, t.ClusterStatus.GetSpec(ctx).Version, metricsResponseCode, "DELETE", "api/v1alpha1/cluster"}
			t.metricsTool.PopulateToPrometheusMetrics(startTime)
			cluster.SetAuditLog(ctx, "Reset", t.ClusterStatus.GetSpec(ctx).ClusterType, t.ClusterStatus.GetSpec(ctx).Version, clusterStatus.GetPhase(), auditMessage, auditReason)
		}()
		logger := log.From(ctx)
		clusterSpec := t.ClusterStatus.GetSpec(ctx)
		taskDetails := t.Tasks
		resetter := operations.NewOperation("reset cluster", t.ClusterStatus, clusterSpec, taskDetails)
		if err := resetter.Run(ctx); err != nil {
			metricsResponseCode = metrcis.DeleteFailed
			auditMessage = "Cluster reset failed"
			auditReason = err.Error()
			clusterStatus.Phase = constants.ClusterPhaseFailed
			conditions.MarkFalse(clusterStatus, v1alpha1.ConditionType_DeleteSuccess, constants.ClusterDeleteMessageFailed, constants.ConditionSeverityWarning, auditMessage)
			logger.Error(err, "Unable to complete reset")
		} else {
			auditMessage = "Cluster has been successfully sundown"
			clusterStatus.Phase = constants.ClusterPhaseDelete
			metricsResponseCode = metrcis.DeleteCompleted
			err = t.ClusterStatus.PurgeAllClusterData(ctx)
			if err != nil {
				logger.Error(err, "unable to remove the cluster data , agent need to be restarted")
			}
		}
	}()

	return nil
}

func (t *DefaultKubernetesProvider) Cluster(ctx context.Context) (*v1alpha1.Cluster, error) {
	var responseCode string
	startTime := time.Now()
	defer func() {
		t.metricsTool.MetricsLabels = []string{t.ClusterStatus.GetSpec(ctx).ClusterType,
			t.ClusterStatus.GetSpec(ctx).Version, responseCode, "GET", "api/v1alpha1/cluster"}
		t.metricsTool.PopulateToPrometheusMetrics(startTime)
	}()
	responseCode = metrcis.Success
	return &v1alpha1.Cluster{
		ApiVersion: "v1alpha1",
		Kind:       "Cluster",
		Spec:       t.ClusterStatus.GetSpec(ctx),
		Status:     t.ClusterStatus.GetStatus(ctx),
	}, nil
}

func (t *DefaultKubernetesProvider) Config(ctx context.Context) ([]byte, error) {
	var metricsResponseCode string
	var startTime time.Time
	var clusterSpec = t.ClusterStatus.GetSpec(ctx)
	defer func() {
		t.metricsTool.MetricsLabels = []string{clusterSpec.ClusterType, clusterSpec.Version, metricsResponseCode, "GET", "api/v1alpha1/config"}
		t.metricsTool.PopulateToPrometheusMetrics(startTime)
	}()
	if !t.IsInitialized(ctx) {
		metricsResponseCode = metrcis.Failed
		return nil, errors.New("cluster is not initialized")
	}
	logger := log.From(ctx).WithValues("ClusterType", clusterSpec.ClusterType, "Version", clusterSpec.Version)
	logger.Info("Retrieving kubeconfig")
	metricsResponseCode = metrcis.Success
	return os.ReadFile(constants.KubeadmKubeconfigPath)
}

func (t *DefaultKubernetesProvider) GetCerts(ctx context.Context) (*v1alpha1.ClusterCertificatesResponse, error) {
	var metricsResponseCode string
	var startTime time.Time
	var clusterSpec = t.ClusterStatus.GetSpec(ctx)
	defer func() {
		t.metricsTool.MetricsLabels = []string{clusterSpec.ClusterType,
			clusterSpec.Version,
			metricsResponseCode,
			"GET", "api/v1alpha1/certs"}
		t.metricsTool.PopulateToPrometheusMetrics(startTime)
	}()
	if !t.IsInitialized(ctx) {
		metricsResponseCode = metrcis.Failed
		return nil, errors.New("cluster is not initialized")
	}
	logger := log.From(ctx).WithValues("ClusterType", clusterSpec.ClusterType, "Version", clusterSpec.Version)
	logger.Info("Retrieving kubernetes control-plane certificates")
	metricsResponseCode = metrcis.Success
	osUtility := osutility.New()
	_, m, err := osUtility.Kubeadm().GetCertsExpiry(ctx)
	if err != nil {
		return nil, err
	}
	response := make([]*v1alpha1.CertsInfo, 0)
	for k, v := range m {
		rotationDays := v - constants.ClusterCertsRotationDays
		if rotationDays <= 0 {
			rotationDays = 0
		}
		response = append(response, &v1alpha1.CertsInfo{
			Name:                  k,
			RemainingDaysToExpire: v,
			ExpiryDate:            time.Now().AddDate(0, 0, int(v)).Format("Jan 02, 2006 15:04 MST"),
			RotationDate:          time.Now().AddDate(0, 0, int(rotationDays)).Format("Jan 02, 2006 15:04 MST"),
		})
	}
	return &v1alpha1.ClusterCertificatesResponse{
		CertsInfo: response,
	}, nil
}

func (t *DefaultKubernetesProvider) ResetConfig(ctx context.Context) error {
	logger := log.From(ctx)
	var metricsResponseCode, auditMessage, auditReason string
	var startTime time.Time
	var clusterStatus = t.ClusterStatus.GetStatus(ctx)
	defer func() {
		t.metricsTool.MetricsLabels = []string{t.ClusterStatus.GetSpec(ctx).ClusterType, t.ClusterStatus.GetSpec(ctx).Version, metricsResponseCode, "DELETE", "api/v1alpha1/certs"}
		t.metricsTool.PopulateToPrometheusMetrics(startTime)
		t.ClusterStatus.SetStatus(ctx, clusterStatus)
		cluster.SetAuditLog(ctx, "Reset Certs", t.ClusterStatus.GetSpec(ctx).ClusterType, t.ClusterStatus.GetSpec(ctx).Version, clusterStatus.GetPhase(), auditMessage, auditReason)
	}()
	if !t.IsInitialized(ctx) {
		auditMessage = "Cluster is not initialized"
		metricsResponseCode = metrcis.ResetFailed
		return fmt.Errorf("cluster is not initialized")
	}
	logger.Info("Resetting kubernetes certificates")
	clusterStatus.Phase = constants.ClusterPhaseKubeConfigResetting
	tasks := t.Tasks
	restConfig := operations.NewOperation("reset-certs", t.ClusterStatus, t.ClusterStatus.GetSpec(ctx), tasks)
	err := restConfig.Run(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("Error resetting certs: %s", err.Error())
		auditMessage = "Error resetting certs"
		auditReason = err.Error()
		metricsResponseCode = metrcis.ResetFailed
		return fmt.Errorf(errMsg)
	}

	auditMessage = "certs successfully reset"
	clusterStatus.Phase = constants.ClusterPhaseProvisioned

	metricsResponseCode = metrcis.ResetDone
	return nil
}

func (t *DefaultKubernetesProvider) Upgrade(ctx context.Context, request *v1alpha1.UpgradeClusterRequest) error {
	var metricsResponseCode, auditMessage, auditReason string
	var startTime time.Time
	var currentClusterSpec *v1alpha1.ClusterSpec
	var clusterStatus = t.ClusterStatus.GetStatus(ctx)
	defer func() {
		t.metricsTool.MetricsLabels = []string{t.ClusterStatus.GetSpec(ctx).ClusterType, t.ClusterStatus.GetSpec(ctx).Version, metricsResponseCode, "PUT", "api/v1alpha1/cluster"}
		t.metricsTool.PopulateToPrometheusMetrics(startTime)
		cluster.SetAuditLog(ctx, "Upgrade", request.Spec.ClusterType, clusterStatus.KubernetesVersion, clusterStatus.GetPhase(), auditMessage, auditReason)
		t.ClusterStatus.SetStatus(ctx, clusterStatus)
	}()
	if !t.IsInitializedForUpgrade(ctx) {
		auditMessage = "Cluster must be installed properly for upgrade"
		metricsResponseCode = metrcis.UpgradeFailed
		clusterStatus.Phase = constants.ClusterPhaseFailed
		return fmt.Errorf("cluster is not initialized for upgrade")
	}
	upgradeVersion := request.Spec.Version
	currentClusterVersion := clusterStatus.KubernetesVersion
	currentExecution := t.ExecutionInProgress(ctx)
	if currentExecution {
		auditMessage = fmt.Sprintf("Currently %v is running, please try after it reaches a terminal state complete or failed", currentExecution)
		return fmt.Errorf("currently %v is running", currentExecution)
	}
	clusterStatus.Phase = constants.ClusterPhaseUpgrading
	auditMessage = fmt.Sprintf("Cluster upgrade to version %s in progress", upgradeVersion)
	currentClusterSpec = t.ClusterStatus.GetSpec(ctx)
	currentClusterSpec.Version = request.Spec.Version
	t.ClusterStatus.SetSpec(ctx, currentClusterSpec)
	go func() {
		defer func() {
			t.metricsTool.MetricsLabels = []string{t.ClusterStatus.GetSpec(ctx).ClusterType, request.Spec.Version, metricsResponseCode, "PUT", "api/v1alpha1/cluster"}
			t.metricsTool.PopulateToPrometheusMetrics(startTime)
			cluster.SetAuditLog(ctx, "Upgrade", request.Spec.ClusterType, clusterStatus.KubernetesVersion, clusterStatus.GetPhase(), auditMessage, auditReason)
			t.ClusterStatus.SetStatus(ctx, clusterStatus)
		}()
		logger := log.From(ctx)
		taskDetails := t.Tasks
		upgrader := operations.NewOperation("upgrade cluster", t.ClusterStatus, request.Spec, taskDetails)
		if err := upgrader.Run(ctx); err != nil {
			logger.Error(err, "Unable to upgrade cluster")
			auditMessage = fmt.Sprintf("failed to upgrade cluster to  %s", upgradeVersion)
			auditReason = err.Error()
			clusterStatus.Phase = constants.ClusterPhaseFailed
			metricsResponseCode = metrcis.UpgradeFailed
			clusterStatus.KubernetesVersion = currentClusterVersion
			// this is only a warning condition. Else Phase will just Fail and status won't have clear picture why its in Failed state
			conditions.MarkFalse(clusterStatus, v1alpha1.ConditionType_UpgradeSuccess, constants.ControlUpgradeMessageFailed, constants.ConditionSeverityWarning, auditMessage)
		} else {
			auditMessage = fmt.Sprintf("Cluster is successfully upgraded to %s", upgradeVersion)
			clusterStatus.Phase = constants.ClusterPhaseProvisioned
			clusterStatus.KubernetesVersion = upgradeVersion
			metricsResponseCode = metrcis.UpgradeDone
		}
	}()
	return nil
}
