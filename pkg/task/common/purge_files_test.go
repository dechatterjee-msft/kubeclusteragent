package common

import (
	"context"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/util/osutility/linux"
	"reflect"
	"testing"
)

func TestNewPurgeFiles(t *testing.T) {
	tests := []struct {
		name string
		want *PurgeFiles
	}{
		{name: "testutil-purge-files-obj", want: NewPurgeFiles()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPurgeFiles(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPurgeFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPurgeFiles_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "testutil-purge-file-name", want: "purge-files"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &PurgeFiles{}
			if got := k.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPurgeFiles_Rollback(t *testing.T) {
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
			k := &PurgeFiles{}
			if err := k.Rollback(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPurgeFiles_Run(t *testing.T) {
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
		{name: "testutil-purge-files-run", args: struct {
			ctx         context.Context
			status      cluster.Status
			clusterSpec *v1alpha1.ClusterSpec
			ou          linux.OSUtil
		}{ctx: context.Background(), status: nil, clusterSpec: nil, ou: linux.NewDryRun()}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &PurgeFiles{}
			if err := k.Run(tt.args.ctx, tt.args.status, tt.args.clusterSpec, tt.args.ou); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
