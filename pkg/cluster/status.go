package cluster

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/log/log"

	"go.uber.org/multierr"
	"kubeclusteragent/gen/go/agent/v1alpha1"
)

type Context struct {
	context.Context
	ClusterSpec   *v1alpha1.ClusterSpec
	ClusterStatus *v1alpha1.ClusterStatus
	State         string
}

type Status interface {
	ClusterSpec(ctx context.Context) *v1alpha1.ClusterSpec
	SetSpec(ctx context.Context, spec *v1alpha1.ClusterSpec)
	GetSpec(ctx context.Context) *v1alpha1.ClusterSpec
	SetStatus(ctx context.Context, status *v1alpha1.ClusterStatus)
	GetStatus(ctx context.Context) *v1alpha1.ClusterStatus
	SetAuditHistory(ctx context.Context, currentOperation *v1alpha1.Operations) error
	GetAuditHistory(ctx context.Context) ([]*v1alpha1.Operations, error)
	PurgeAllClusterData(ctx context.Context) error
	StoreConfigMap(ctx context.Context, configMap *v1.ConfigMap, name string) error
	GetConfigMap(ctx context.Context, name string) (*v1.ConfigMap, error)
}

type LiveStatus struct {
	dryRun      bool
	clusterSpec *v1alpha1.ClusterSpec
}

var _ Status = &LiveStatus{}
var clusterInfo clusterStore = &liveStore{}

func NewLiveStatus(ctx context.Context, dryRun bool) (*LiveStatus, error) {
	s := &LiveStatus{
		dryRun:      dryRun,
		clusterSpec: &v1alpha1.ClusterSpec{},
	}
	spec := s.GetSpec(ctx)
	s.clusterSpec = spec
	return s, nil
}

func (s *LiveStatus) GetStatus(ctx context.Context) *v1alpha1.ClusterStatus {
	logger := log.From(ctx).WithName("cluster-store").WithName("get-status")
	clusterStatus, err := clusterInfo.ReadClusterStatus(ctx)
	if err != nil {
		logger.Error(err, "error occurred while getting status", "ClusterGetStatus", "failed")
	}
	if clusterStatus == nil {
		return &v1alpha1.ClusterStatus{
			Phase: constants.ClusterPhaseNotInitialised,
		}
	}
	return clusterStatus
}

func (s *LiveStatus) SetStatus(ctx context.Context, status *v1alpha1.ClusterStatus) {
	logger := log.From(ctx).WithName("cluster-store").WithName("set-status")
	err := clusterInfo.WriteClusterStatus(ctx, status)
	if err != nil {
		logger.Error(err, "error occurred while saving the status", "ClusterWriteStatus", "failed")
	}
}

func (s *LiveStatus) GetSpec(ctx context.Context) *v1alpha1.ClusterSpec {
	logger := log.From(ctx).WithName("cluster-store").WithName("get-spec")
	clusterSpec, err := clusterInfo.ReadClusterSpec(ctx)
	if err != nil {
		logger.Error(err, "error occurred while getting the cluster spec", "ClusterGetSpec", "failed")
	}
	if clusterSpec == nil {
		return &v1alpha1.ClusterSpec{}
	}
	return clusterSpec
}

func (s *LiveStatus) ClusterSpec(ctx context.Context) *v1alpha1.ClusterSpec {
	return s.GetSpec(ctx)
}

func (s *LiveStatus) SetSpec(ctx context.Context, spec *v1alpha1.ClusterSpec) {
	logger := log.From(ctx).WithName("cluster-store").WithName("set-spec")
	if spec == nil {
		return
	}
	err := clusterInfo.WriteClusterSpec(ctx, spec)
	if err != nil {
		logger.Error(err, "error occurred while saving the cluster spec", "ClusterSetSpec", "failed")
	}
}

func (s *LiveStatus) SetAuditHistory(ctx context.Context, currentCondition *v1alpha1.Operations) error {
	logger := log.From(ctx).WithName("cluster-store").WithName("set-audit-history")
	audits, err := s.GetAuditHistory(ctx)
	if err != nil {
		logger.Error(err, "error occurred while setting the audit history of the cluster", "GetHistory", "failed")
		return err
	}
	err = clusterInfo.WriteAuditHistory(ctx, append(audits, currentCondition))
	if err != nil {
		logger.Error(err, "error occurred while setting the audit history of the cluster", "SetAuditHistory", "failed")
		return err
	}
	return nil
}

func (s *LiveStatus) StoreConfigMap(ctx context.Context, configMap *v1.ConfigMap, name string) error {
	return clusterInfo.WriteConfigMap(ctx, configMap, name)
}

func (s *LiveStatus) GetConfigMap(ctx context.Context, name string) (*v1.ConfigMap, error) {
	return clusterInfo.ReadConfigMap(ctx, name)
}

func (s *LiveStatus) GetAuditHistory(ctx context.Context) ([]*v1alpha1.Operations, error) {
	logger := log.From(ctx).WithName("cluster-store").WithName("get-audit-history")
	auditHistory, err := clusterInfo.ReadAuditHistory(ctx)
	if err != nil {
		logger.Error(err, "error occurred while getting the audit history of the cluster", "GetAuditHistory", "failed")
		return nil, err
	}
	return auditHistory, nil
}

func (s *LiveStatus) PurgeAllClusterData(ctx context.Context) error {
	return clusterInfo.PurgeAll(ctx)
}

type StateData struct {
	APIVersion  string                `json:"apiVersion"`
	Kind        string                `json:"kind"`
	ClusterSpec *v1alpha1.ClusterSpec `json:"clusterSpec"`
}

func (d *StateData) Validate() error {
	var err error
	if d.APIVersion != "v1alpha1" {
		err = multierr.Append(err, fmt.Errorf("unknown API version: %q", d.APIVersion))
	}
	if d.Kind != "StateData" {
		err = multierr.Append(err, fmt.Errorf("unknown kind: %q", d.Kind))
	}
	return err
}
