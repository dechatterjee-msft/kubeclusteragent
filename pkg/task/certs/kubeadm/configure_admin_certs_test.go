package kubeadm

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/util/osutility/linux"
	"reflect"
	"testing"
)

func TestAdminCertsRotation_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "testutil-configure-admin-certs", want: "configure-admin-certs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := AdminCertsRotation{}
			if got := u.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdminCertsRotation_Rollback(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			u := AdminCertsRotation{}
			if err := u.Rollback(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdminCertsRotation_Run(t *testing.T) {
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
		{name: "testutil-rotate-admin-certs-run", args: struct {
			ctx         context.Context
			status      cluster.Status
			clusterSpec *v1alpha1.ClusterSpec
			ou          linux.OSUtil
		}{ctx: context.Background(), status: nil, clusterSpec: nil, ou: linux.NewDryRun()}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := AdminCertsRotation{}
			if err := u.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRotateAdminCerts(t *testing.T) {
	tests := []struct {
		name string
		want *AdminCertsRotation
	}{
		{name: "testutil-new-rotate-admin-certs", want: NewRotateAdminCerts()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRotateAdminCerts(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRotateAdminCerts() = %v, want %v", got, tt.want)
			}
		})
	}
}
