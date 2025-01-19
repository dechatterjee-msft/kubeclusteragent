package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/util/osutility"
	"reflect"
	"testing"
)

func TestDeleteAdminKubeConfig_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "testutil-delete-admin-kube-config", want: "delete-admin-kube-config"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DeleteAdminKubeConfig{}
			if got := r.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteAdminKubeConfig_Rollback(t *testing.T) {
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
		// TODO: Add testutil cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DeleteAdminKubeConfig{}
			if err := r.Rollback(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteAdminKubeConfig_Run(t *testing.T) {
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
		{name: "kubeadm-delete-cluster-kubeconfig-admin", args: struct {
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
		t.Run(tt.name, func(t *testing.T) {
			r := &DeleteAdminKubeConfig{}
			if err := r.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewDeleteAdminKubeConfig(t *testing.T) {
	tests := []struct {
		name string
		want *DeleteAdminKubeConfig
	}{
		{name: "testutil-new-admin-config", want: NewDeleteAdminKubeConfig()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDeleteAdminKubeConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDeleteAdminKubeConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
