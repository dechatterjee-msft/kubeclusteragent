package osutility_test

import (
	"context"
	"io/fs"
	"kubeclusteragent/pkg/util/osutility"
	"kubeclusteragent/pkg/util/osutility/packagemanager"
	"testing"

	"github.com/golang/mock/gomock"
	"kubeclusteragent/mocks"
)

type sysctlHarness struct {
	exec *mocks.MockExec
	fs   *mocks.MockFilesystem
	pkg  *osutility.LiveSysctl
}

func newSysctlHarness(ctrl *gomock.Controller) *sysctlHarness {
	exec := mocks.NewMockExec(ctrl)
	fsUtil := mocks.NewMockFilesystem(ctrl)
	h := &sysctlHarness{
		exec: exec,
		fs:   fsUtil,
		pkg:  osutility.NewLiveSysctl(exec, fsUtil),
	}

	return h
}

func (h *sysctlHarness) ExpectCommand(name string, env, args []string, ret []packagemanager.any) {
	h.exec.EXPECT().Command(gomock.Any(), name, env, args).Return(ret...)
}

func (h *sysctlHarness) WriteFile(destination string, data []byte, perm fs.FileMode, err error) {
	h.fs.EXPECT().WriteFile(gomock.Any(), destination, data, perm).Return(err)
}

func TestLiveSysctl_Reload(t *testing.T) {
	tests := []struct {
		name    string
		harness func(ctrl *gomock.Controller) *sysctlHarness
		wantErr bool
	}{
		{
			name: "success",
			harness: func(ctrl *gomock.Controller) *sysctlHarness {
				h := newSysctlHarness(ctrl)
				h.ExpectCommand("sysctl", nil, []string{"--system"}, []packagemanager.any{0, nil, nil})
				return h
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)

			l := h.pkg
			err := l.Reload(context.Background())
			test.CheckError(t, test.wantErr, err)
		})
	}
}

func TestLiveSysctl_Set(t *testing.T) {
	type args struct {
		values map[string]string
	}

	tests := []struct {
		name    string
		args    args
		harness func(ctrl *gomock.Controller) *sysctlHarness
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				values: map[string]string{
					"foo": "1",
				},
			},
			harness: func(ctrl *gomock.Controller) *sysctlHarness {
				h := newSysctlHarness(ctrl)
				h.WriteFile("/etc/sysctl.d/99-kubernetes.conf", []byte("foo = 1\n"), 0644, nil)
				return h
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)

			l := h.pkg
			err := l.Set(context.Background(), test.args.values)
			test.CheckError(t, test.wantErr, err)
		})
	}
}
