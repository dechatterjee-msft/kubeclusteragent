package common

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/util/osutility"
	"reflect"
	"testing"
)

func TestKubeletReload_Name(t1 *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "test-kubelet-reload-name", want: "kubelet-reload"},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &KubeletReload{}
			if got := t.Name(); got != tt.want {
				t1.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubeletReload_Rollback(t1 *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &KubeletReload{}
			if err := t.Rollback(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t1.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKubeletReload_Run(t1 *testing.T) {
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
		{name: "test-reload-kubelet-run", args: struct {
			ctx         context.Context
			status      cluster.Status
			clusterSpec *v1alpha1.ClusterSpec
			ou          osutility.OSUtil
		}{ctx: context.Background(), status: nil, clusterSpec: nil, ou: osutility.NewDryRun()}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &KubeletReload{}
			if err := t.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t1.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewKubeletReload(t *testing.T) {
	tests := []struct {
		name string
		want *KubeletReload
	}{
		{name: "test-new-reload-kubelet", want: NewKubeletReload()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKubeletReload(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKubeletReload() = %v, want %v", got, tt.want)
			}
		})
	}
}
