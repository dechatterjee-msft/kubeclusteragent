package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/util/osutility/linux"
	"reflect"
	"testing"
)

func TestCertsRotation_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "rotate-certs", want: "kubeadm-rotate-certs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := CertsRotation{}
			if got := u.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCertsRotation_Run(t *testing.T) {
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
		{name: "rotate-certs-run", args: struct {
			ctx         context.Context
			status      cluster.Status
			clusterSpec *v1alpha1.ClusterSpec
			ou          linux.OSUtil
		}{ctx: context.Background(), status: nil, clusterSpec: nil, ou: linux.NewDryRun()}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := CertsRotation{}
			if err := u.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRotateCerts(t *testing.T) {
	tests := []struct {
		name string
		want *CertsRotation
	}{
		{name: "NewCertsRotate", want: NewRotateCerts()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRotateCerts(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRotateCerts() = %v, want %v", got, tt.want)
			}
		})
	}
}
