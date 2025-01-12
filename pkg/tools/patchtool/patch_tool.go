package patchtool

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/tools/metricstool"
	"kubeclusteragent/pkg/util/conditions"
	"kubeclusteragent/pkg/util/log/log"
	"time"

	"kubeclusteragent/pkg/constants"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/operations"
)

type ClusterConfigurationChange interface {
	Patch(ctx context.Context, request *v1alpha1.PatchClusterRequest) error
}

type ClusterConfigTool struct {
	clusterStatus cluster.Status
	dryRun        bool
	metricsTool   metricstool.PrometheusMetricsTool
}

var _ ClusterConfigurationChange = &ClusterConfigTool{}

func NewClusterConfigInstallTool(clusterStatus cluster.Status, dryRun bool) *ClusterConfigTool {
	t := &ClusterConfigTool{
		clusterStatus: clusterStatus,
		dryRun:        dryRun,
	}
	return t
}

func (t *ClusterConfigTool) IsInitializedForPatch(ctx context.Context) bool {
	clusterStatus := t.clusterStatus.GetStatus(ctx)
	switch clusterStatus.Phase { //nolint
	case constants.ClusterPhaseProvisioned:
		return true
	}

	return false
}

func (t *ClusterConfigTool) ExecutionInProgress(ctx context.Context) bool {
	switch t.clusterStatus.GetStatus(ctx).Phase {
	case constants.ClusterPhaseProvisioning:
		return true
	case constants.ClusterPhaseKubeConfigResetting:
		return true
	case constants.ClusterPhaseDeleting:
		return true
	}

	return false
}

func (t ClusterConfigTool) Patch(ctx context.Context, request *v1alpha1.PatchClusterRequest) error {
	logger := log.From(ctx).WithName("Patch Configuration")
	var metricsResponseCode, auditMessage, auditReason string
	var startTime time.Time

	var clusterStatus = t.clusterStatus.GetStatus(ctx)

	defer func() {
		t.metricsTool.MetricsLabels = []string{request.Spec.ClusterType, request.Spec.Version, metricsResponseCode, "POST", "api/v1alpha1/cluster"}
		t.metricsTool.PopulateToPrometheusMetrics(startTime)
		t.clusterStatus.SetStatus(ctx, clusterStatus)
		cluster.SetAuditLog(ctx, "Patch", request.Spec.ClusterType, request.Spec.Version, clusterStatus.GetPhase(), auditMessage, auditReason)
	}()
	if !t.IsInitializedForPatch(ctx) {
		auditMessage = "Cluster must be installed properly for Patch to take place"

		return fmt.Errorf("cluster is not initialized for patch")
	}
	t.clusterStatus.SetSpec(ctx, request.Spec)
	currentExecution := t.ExecutionInProgress(ctx)
	if currentExecution {
		auditMessage = fmt.Sprintf("Currently %v is running, please try after it reaches a terminal state complete or failed", currentExecution)

		return fmt.Errorf("currently %v is running", currentExecution)
	}

	auditMessage = "Cluster patch is in progress"

	go func() {
		defer func() {
			t.clusterStatus.SetStatus(ctx, clusterStatus)
			cluster.SetAuditLog(ctx, "Patch", request.Spec.ClusterType, request.Spec.Version, clusterStatus.GetPhase(), auditMessage, auditReason)
		}()
		var options []operations.Option
		if t.dryRun {
			options = append(options, operations.DryRun())
		}
		taskDetails := buildPatchOptions(options...)
		patcher := operations.NewOperation("patch cluster", t.clusterStatus, request.Spec, taskDetails)
		if err := patcher.Run(ctx); err != nil {
			logger.Error(err, "Unable to patch cluster")
			auditMessage = "failed to patch cluster"
			auditReason = err.Error()
			conditions.MarkFalse(clusterStatus, v1alpha1.ConditionType_PackageReady, constants.PackageReadStatusMessageFailed, clusterStatus.GetPhase(), auditMessage, auditReason)
		} else {
			// TODO check the diff between two structs and apply that to current Cluster Spec
			auditMessage = "Cluster is successfully patched"
			conditions.MarkTrue(clusterStatus, v1alpha1.ConditionType_PackageReady)
		}
	}()
	return nil
}
