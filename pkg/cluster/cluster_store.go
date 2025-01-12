package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/db"
	"sort"

	"go.uber.org/multierr"
	"kubeclusteragent/gen/go/agent/v1alpha1"
)

type clusterStore interface {
	WriteClusterSpec(ctx context.Context, clusterSpec *v1alpha1.ClusterSpec) error
	ReadClusterSpec(ctx context.Context) (*v1alpha1.ClusterSpec, error)
	WriteClusterStatus(ctx context.Context, clusterStatus *v1alpha1.ClusterStatus) error
	ReadClusterStatus(ctx context.Context) (*v1alpha1.ClusterStatus, error)
	WriteAuditHistory(ctx context.Context, auditHistory []*v1alpha1.Operations) error
	ReadAuditHistory(ctx context.Context) ([]*v1alpha1.Operations, error)
	PurgeAll(ctx context.Context) error
	WriteConfigMap(ctx context.Context, configMap *v1.ConfigMap, name string) error
	ReadConfigMap(ctx context.Context, name string) (*v1.ConfigMap, error)
}

type liveStore struct {
	clusterStore db.Store
}

const (
	clusterSpecKey         = "clusterSpec"
	clusterStatusKey       = "clusterStatus"
	clusterAuditHistoryKey = "clusterAudits"
	NilStingInBoltDB       = "<nil>"
)

func (s *liveStore) PurgeAll(ctx context.Context) error {
	clusterStoreConnect := s.clusterStore.Connect(db.DBClusterTableName)
	err := clusterStoreConnect.Set(clusterSpecKey, "")
	if err != nil {
		return err
	}
	clusterStoreConnect = s.clusterStore.Connect(db.DBClusterStatusTableName)
	err = clusterStoreConnect.Set(clusterStatusKey, "")
	if err != nil {
		return err
	}
	err = clusterStoreConnect.Set(clusterSpecKey, "")
	if err != nil {
		return err
	}
	return err
}

func (s *liveStore) ReadClusterSpec(ctx context.Context) (*v1alpha1.ClusterSpec, error) {
	clusterStoreConnect := s.clusterStore.Connect(db.DBClusterTableName)
	clusterSpec := &v1alpha1.ClusterSpec{}
	if clusterStoreConnect != nil {
		data := clusterStoreConnect.Get(clusterSpecKey)
		clusterSpecStr := fmt.Sprintf("%v", data)
		if clusterSpecStr != NilStingInBoltDB {
			if err := json.Unmarshal([]byte(clusterSpecStr), clusterSpec); err != nil {
				return nil, multierr.Append(fmt.Errorf("no cluster spec found"), err)
			}
		}
	}
	return clusterSpec, nil
}

func (s *liveStore) WriteClusterStatus(ctx context.Context, clusterStatus *v1alpha1.ClusterStatus) error {

	stateStore := s.clusterStore.Connect(db.DBClusterStatusTableName)
	if stateStore != nil {
		data, err := json.MarshalIndent(clusterStatus, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal state data to JSON: %w", err)
		}
		err = stateStore.Set(clusterStatusKey, string(data))
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("error occoured making connection with the data store")
}

func (s *liveStore) WriteAuditHistory(ctx context.Context, auditHistory []*v1alpha1.Operations) error {
	auditHistory = sortAuditHistoryByTimestamp(auditHistory)
	stateStore := s.clusterStore.Connect(db.DBClusterAuditHistoryTableName)
	if stateStore != nil {
		data, err := json.MarshalIndent(auditHistory, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal state data to JSON: %w", err)
		}
		err = stateStore.Set(clusterAuditHistoryKey, string(data))
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("error occoured making connection with the data store")

}

func (s *liveStore) ReadAuditHistory(ctx context.Context) ([]*v1alpha1.Operations, error) {
	stateStore := s.clusterStore.Connect(db.DBClusterAuditHistoryTableName)
	auditHistory := make([]*v1alpha1.Operations, 0)
	if stateStore != nil {
		data := stateStore.Get(clusterAuditHistoryKey)
		stateDataStr := fmt.Sprintf("%v", data)
		if stateDataStr != NilStingInBoltDB {
			if err := json.Unmarshal([]byte(stateDataStr), &auditHistory); err != nil {
				return nil, multierr.Append(fmt.Errorf("no audit history found"), err)
			}
		}
	}
	return auditHistory, nil
}

func (s *liveStore) ReadClusterStatus(ctx context.Context) (*v1alpha1.ClusterStatus, error) {
	stateStore := s.clusterStore.Connect(db.DBClusterStatusTableName)
	clusterStatus := &v1alpha1.ClusterStatus{
		Phase: constants.ClusterPhaseNotInitialised,
		Conditions: []*v1alpha1.Condition{
			{Type: v1alpha1.ConditionType_ClusterReady, Status: "False"},
			{Type: v1alpha1.ConditionType_NodeReady, Status: "False"},
			{Type: v1alpha1.ConditionType_ControlPlaneReady, Status: "False"},
		},
	}
	if stateStore != nil {
		data := stateStore.Get(clusterStatusKey)
		stateDataStr := fmt.Sprintf("%v", data)
		if stateDataStr != NilStingInBoltDB {
			if err := json.Unmarshal([]byte(stateDataStr), clusterStatus); err != nil {
				return nil, multierr.Append(fmt.Errorf("no cluster status found"), err)
			}
		}
	}
	return clusterStatus, nil
}

func (s *liveStore) WriteClusterSpec(ctx context.Context, clusterSpec *v1alpha1.ClusterSpec) error {
	stateStore := s.clusterStore.Connect(db.DBClusterTableName)
	if stateStore != nil {
		data, err := json.MarshalIndent(clusterSpec, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal state data to JSON: %w", err)
		}
		err = stateStore.Set(clusterSpecKey, string(data))
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("error occoured making connection with the data store")
}

func (s *liveStore) WriteConfigMap(ctx context.Context, configMap *v1.ConfigMap, name string) error {
	stateStore := s.clusterStore.Connect(db.DBClusterTableName)
	if stateStore != nil {
		data, err := json.MarshalIndent(configMap, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal state data to JSON: %w", err)
		}
		err = stateStore.Set(name, string(data))
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("error occoured making connection with the data store")
}

func (s *liveStore) ReadConfigMap(ctx context.Context, name string) (*v1.ConfigMap, error) {
	stateStore := s.clusterStore.Connect(db.DBClusterTableName)
	if stateStore != nil {
		k8sConfigMap := new(v1.ConfigMap)
		configMap := stateStore.Get(name)
		configMapStr := fmt.Sprintf("%v", configMap)
		if configMapStr != NilStingInBoltDB {
			if err := json.Unmarshal([]byte(configMapStr), k8sConfigMap); err != nil {
				return nil, multierr.Append(fmt.Errorf("no cluster status found"), err)
			}
			return k8sConfigMap, nil
		}
	}
	return nil, fmt.Errorf("error occoured making connection with the data store")
}

func sortAuditHistoryByTimestamp(audits []*v1alpha1.Operations) []*v1alpha1.Operations {
	sort.Slice(audits, func(i, j int) bool {
		return audits[i].LastExecuted.AsTime().Before(audits[i].LastExecuted.AsTime())
	})
	return audits
}
