package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/util/osutility"
	"reflect"
	"testing"
)

func TestNewKubeadmReset(t *testing.T) {
	tests := []struct {
		name string
		want *Reset
	}{
		{name: "NewKubeadmRest", want: NewKubeadmReset()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKubeadmReset(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKubeadmReset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReset_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "kubeadm-reset", want: "kubeadm-reset"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reset{}
			if got := r.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReset_Run(t *testing.T) {
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
		{name: "delete-cluster", args: struct {
			ctx         context.Context
			status      cluster.Status
			clusterSpec *v1alpha1.ClusterSpec
			ou          osutility.OSUtil
		}{ctx: context.Background(), status: nil, clusterSpec: nil, ou: osutility.NewDryRun()}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reset{}
			if err := r.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
