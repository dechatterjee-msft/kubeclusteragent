package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/util/osutility/linux"
	"reflect"
	"testing"
)

func TestCluster_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "upgrade-cluster", want: "upgrade-cluster"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := Cluster{}
			if got := u.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_Run(t *testing.T) {
	type args struct {
		ctx         context.Context
		status      cluster.Status
		clusterSpec *v1alpha1.ClusterSpec
		ou          linux.OSUtil
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "kubeadm-upgrade-cluster", args: struct {
			ctx         context.Context
			status      cluster.Status
			clusterSpec *v1alpha1.ClusterSpec
			ou          linux.OSUtil
		}{ctx: context.Background(), status: nil, clusterSpec: &v1alpha1.ClusterSpec{
			ClusterType: "kubeadm",
			ClusterName: "testutil-cluster",
			Networking: &v1alpha1.ClusterNetworking{
				PodSubnet: "100.100.0.0/16",
				SvcSubnet: "100.101.0.0/16",
				ClusterCni: &v1alpha1.ContainerNetworkInterface{
					Name:    "calico",
					Version: "v3.25.1",
				},
			},
			Storage: &v1alpha1.ClusterStorage{
				ClusterCsi: &v1alpha1.ContainerStorageInterface{
					Name:    "local-path-storage",
					Version: "v0.0.24",
				},
			},
			ApiServer:        &v1alpha1.ClusterAPIServer{},
			Version:          "v1.26.5+",
			DisableWorkloads: new(bool),
			ExtraArgs:        nil,
		}, ou: linux.NewDryRun(),
		}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := Cluster{}
			if err := u.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewUpgradeCluster(t *testing.T) {
	tests := []struct {
		name string
		want *Cluster
	}{
		{name: "NewUpgradeCluster", want: NewUpgradeCluster()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUpgradeCluster(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUpgradeCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}
