package common

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/util/osutility/linux"
	"reflect"
	"testing"
)

func TestDrainNode_Name(t1 *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "testutil-drain-node", want: "drain-node"},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &DrainNode{}
			if got := t.Name(); got != tt.want {
				t1.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDrainNode_Rollback(t1 *testing.T) {
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
		// TODO: Add testutil cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &DrainNode{}
			if err := t.Rollback(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t1.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDrainNode_Run(t1 *testing.T) {
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
		{name: "testutil-drain-node-run", args: struct {
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
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &DrainNode{}
			if err := t.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t1.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewDrainNode(t *testing.T) {
	tests := []struct {
		name string
		want *DrainNode
	}{
		{name: "testutil-new-drain-node", want: NewDrainNode()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDrainNode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDrainNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
