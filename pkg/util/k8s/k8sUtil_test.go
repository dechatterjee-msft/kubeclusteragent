package k8s

import (
	"context"
	"kubeclusteragent/pkg/util/osutility/linux"
	"testing"
	"time"
)

func TestK8sUtil_NodeWorkloadScheduler(t *testing.T) {
	kubectlClient = &linux.FakeKubectl{}
	hostUtil = &linux.FakeHost{}
	retryCount = 0
	sleep = 1 * time.Millisecond
	type args struct {
		ctx           context.Context
		operationName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "happy", args: args{
			ctx:           context.Background(),
			operationName: "uncordon",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8s := &K8sUtil{}
			if err := k8s.NodeWorkloadScheduler(tt.args.ctx, tt.args.operationName); (err != nil) != tt.wantErr {
				t.Errorf("NodeWorkloadScheduler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestK8sUtil_NodeWorkloadScheduler_HostError(t *testing.T) {
	kubectlClient = &linux.FakeKubectl{}
	hostUtil = &linux.FakeHostWithErr{}
	type args struct {
		ctx           context.Context
		operationName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "HostNameError", args: args{
			ctx:           context.Background(),
			operationName: "testutil",
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8s := &K8sUtil{}
			if err := k8s.NodeWorkloadScheduler(tt.args.ctx, tt.args.operationName); (err != nil) != tt.wantErr {
				t.Errorf("NodeWorkloadScheduler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestK8sUtil_NodeWorkloadScheduler_KubeClientError(t *testing.T) {
	kubectlClient = &linux.FakeKubectlError{}
	hostUtil = &linux.FakeHost{}
	type args struct {
		ctx           context.Context
		operationName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "KubeclientError", args: args{
			ctx:           context.Background(),
			operationName: "testutil",
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8s := &K8sUtil{}
			if err := k8s.NodeWorkloadScheduler(tt.args.ctx, tt.args.operationName); (err != nil) != tt.wantErr {
				t.Errorf("NodeWorkloadScheduler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
