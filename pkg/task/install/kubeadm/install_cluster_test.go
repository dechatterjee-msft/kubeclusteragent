package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/osutility"
	"reflect"
	"testing"
)

func TestCluster_Name(t1 *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Install-Cluster-Name",
			want: "install-kubeadm-cluster",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Cluster{}
			if got := t.Name(); got != tt.want {
				t1.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_Run(t1 *testing.T) {
	type args struct {
		ctx         context.Context
		status      cluster.Status
		clusterSpec *v1alpha1.ClusterSpec
		ou          osutility.OSUtil
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "kubeadm-install-cluster", args: struct {
			ctx         context.Context
			status      cluster.Status
			clusterSpec *v1alpha1.ClusterSpec
			ou          osutility.OSUtil
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
		}, ou: osutility.NewDryRun(),
		}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Cluster{}
			apiServerAddress = constants.PrivateIPv4Address
			if err := t.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t1.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCluster_generateTemplate(t1 *testing.T) {
	type args struct {
		ctx         context.Context
		clusterSpec *v1alpha1.ClusterSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "generate-template", args: struct {
			ctx         context.Context
			clusterSpec *v1alpha1.ClusterSpec
		}{
			ctx: context.Background(),
			clusterSpec: &v1alpha1.ClusterSpec{
				ClusterType: "kubeadm",
				ClusterName: "testutil-cluster",
				Networking: &v1alpha1.ClusterNetworking{
					PodSubnet: "100.100.0.0/16,2001:db8:1::/112",
					SvcSubnet: "100.101.0.0/16,2001:db8:2::/112",
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
			},
		},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Cluster{}
			apiServerAddress = constants.PrivateIPv4Address
			got, err := t.generateTemplate(tt.args.ctx, tt.args.clusterSpec)
			if (err != nil) != tt.wantErr {
				t1.Errorf("generateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t1.Fail()
			}
		})
	}
}

func TestNewInstallCluster(t *testing.T) {
	tests := []struct {
		name string
		want *Cluster
	}{
		{name: "install-new-cluster-obj", want: NewInstallCluster()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInstallCluster(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInstallCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}
