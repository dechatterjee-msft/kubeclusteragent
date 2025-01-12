package cluster

import (
	"context"
	"kubeclusteragent/pkg/util/log/log"

	"google.golang.org/protobuf/types/known/timestamppb"
	"kubeclusteragent/gen/go/agent/v1alpha1"
)

type AuditCluster interface {
	GetAuditlogs(ctx context.Context) ([]*v1alpha1.Operations, error)
	SetAuditLog(ctx context.Context, Operation string, request *v1alpha1.CreateClusterRequest, status string, message string, reason string)
}

var auditStore Status = &LiveStatus{}

func SetAuditLog(ctx context.Context, operation string, clusterType string, version string, status string, message string, reason string) {
	logger := log.From(ctx).WithName("cluster-audit").WithName("generate-audit-history")
	auditCondition := &v1alpha1.Operations{
		Operation:      operation,
		Status:         status,
		Reason:         reason,
		Message:        message,
		LastExecuted:   timestamppb.Now(),
		ClusterType:    clusterType,
		CurrentVersion: version,
	}
	err := auditStore.SetAuditHistory(ctx, auditCondition)
	if err != nil {
		logger.Error(err, "unable to set condition to audit history")
	}
}

func GetAuditLogs(ctx context.Context) ([]*v1alpha1.Operations, error) {
	return auditStore.GetAuditHistory(ctx)
}
