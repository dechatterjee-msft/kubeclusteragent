//go:build ignore

package generator

//go:generate mockgen -destination=mocks/mock_cluster_status.go -package=mocks -mock_names Status=MockClusterStatus kubeclusteragent/pkg/cluster Status
//go:generate mockgen -destination=mocks/mock_osutil_exec.go -package=mocks kubeclusteragent/pkg/util/osutility Exec
//go:generate mockgen -destination=mocks/mock_osutil_fs.go -package=mocks kubeclusteragent/pkg/util/osutility Filesystem
//go:generate mockgen -destination=mocks/agent_service/mock_service.go -package=pkg/mocks kubeclusteragent/pkg/agent Service
//go:generate mockgen -destination=mocks/agent_data_store/mock_cluster_store.go -package=pkg/mocks kubeclusteragent/pkg/cluster clusterStore
//go:generate mockgen -destination=mocks/agent_task/mock_agent_task.go -package=pkg/mocks kubeclusteragent/pkg/task Task
