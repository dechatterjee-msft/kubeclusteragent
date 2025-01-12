package osutility_test

import (
	"context"
	"errors"
	"io/fs"
	"kubeclusteragent/pkg/util/osutility"
	"kubeclusteragent/pkg/util/test"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"kubeclusteragent/mocks"
)

type any = interface{}

type packageHarness struct {
	exec *mocks.MockExec
	fs   *mocks.MockFilesystem
	pkg  *osutility.FakePackageManager
}

func newPackageHarness(ctrl *gomock.Controller) *packageHarness {
	exec := mocks.NewMockExec(ctrl)
	fsUtil := mocks.NewMockFilesystem(ctrl)
	h := &packageHarness{
		exec: exec,
		fs:   fsUtil,
		pkg:  osutility.NewFakePackageManager(exec, fsUtil),
	}

	return h
}

func (h *packageHarness) ExpectCommand(name string, env, args []string, ret []any) {
	h.exec.EXPECT().Command(gomock.Any(), name, env, args).Return(ret...).AnyTimes()
}

func (h *packageHarness) WriteFile(destination string, data []byte, perm fs.FileMode, err error) {
	h.fs.EXPECT().WriteFile(gomock.Any(), destination, data, perm).Return(err).AnyTimes()
}

func TestLivePackage_CheckInstalled(t *testing.T) {
	type args struct {
		packageName string
	}

	tests := []struct {
		name    string
		args    args
		harness func(ctrl *gomock.Controller) *packageHarness
		want    bool
	}{
		{
			name: "package exists",
			args: args{
				packageName: "package",
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("dpkg", nil, []string{"-l", "package"}, []any{0, nil, nil})
				return h
			},
			want: true,
		},
		{
			name: "empty package name",
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				return h
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)

			l := h.pkg
			got := l.CheckInstalled(context.Background(), test.args.packageName)
			require.Equal(t, test.want, got)
		})
	}
}

func TestLivePackage_Install(t *testing.T) {
	type args struct {
		packageNames []string
	}
	tests := []struct {
		name      string
		args      args
		harness   func(ctrl *gomock.Controller) *packageHarness
		wantError bool
	}{
		{
			name: "install package successfully",
			args: args{
				packageNames: []string{"package1"},
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("apt-get", nil, []string{"install", "-y", "package1"}, []any{0, nil, nil})
				return h
			},
			wantError: false,
		},
		{
			name: "install multiple packages successfully",
			args: args{
				packageNames: []string{"package1", "package2"},
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("apt-get", nil, []string{"install", "-y", "package1", "package2"}, []any{0, nil, nil})
				return h
			},
			wantError: false,
		},
		{
			name: "install package with error code != 0",
			args: args{
				packageNames: []string{"package1"},
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("apt-get", nil, []string{"install", "-y", "package1"}, []any{1, nil, nil})
				return h
			},
			wantError: false,
		},
		{
			name: "command returns error",
			args: args{
				packageNames: []string{"package1"},
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("apt-get",
					nil,
					[]string{"install", "-y", "package1"},
					[]any{1, nil, errors.New("fail")})
				return h
			},
			wantError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)
			l := h.pkg
			err := l.Install(context.Background(), test.args.packageNames...)
			test.CheckError(t, test.wantError, err)
		})
	}
}

func TestLivePackage_Update(t *testing.T) {

	tests := []struct {
		name      string
		harness   func(ctrl *gomock.Controller) *packageHarness
		wantError bool
	}{
		{
			name: "zero exit code",
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("apt-get", nil, []string{"update"}, []any{0, nil, nil})
				return h
			},
			wantError: false,
		},
		{
			name: "non zero exit code",
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("apt-get", nil, []string{"update"}, []any{1, nil, nil})
				return h
			},
			wantError: false,
		},
		{
			name: "error",
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("apt-get", nil, []string{"update"}, []any{0, nil, errors.New("fail")})

				return h
			},
			wantError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)
			l := h.pkg
			err := l.Update(context.Background())
			test.CheckError(t, test.wantError, err)
		})
	}
}

func TestLivePackage_AddKey(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name      string
		args      args
		harness   func(ctrl *gomock.Controller) *packageHarness
		wantError bool
	}{
		{
			name: "zero exit code",
			args: args{
				url: "https://example.com",
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("bash",
					nil,
					[]string{"-c", "wget -qO - https://example.com | sudo apt-key add -"},
					[]any{0, nil, nil})

				return h
			},
			wantError: false,
		},
		{
			name: "non zero exit code",
			args: args{
				url: "https://example.com",
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("bash",
					nil,
					[]string{"-c", "wget -qO - https://example.com | sudo apt-key add -"},
					[]any{1, nil, nil})
				return h
			},
			wantError: false,
		},
		{
			name: "error",
			args: args{
				url: "https://example.com",
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.ExpectCommand("bash",
					nil,
					[]string{"-c", "wget -qO - https://example.com | sudo apt-key add -"},
					[]any{0, nil, errors.New("fail")})

				return h
			},
			wantError: false,
		},
		{
			name: "blank url",
			args: args{
				url: "",
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				return h
			},
			wantError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)
			l := h.pkg
			err := l.AddKey(context.Background(), test.args.url)
			test.CheckError(t, test.wantError, err)
		})
	}
}

func TestLivePackage_AddRepository(t *testing.T) {
	type args struct {
		repository string
		filename   string
	}

	tests := []struct {
		name    string
		args    args
		harness func(ctrl *gomock.Controller) *packageHarness
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				repository: "repo",
				filename:   "test",
			},
			harness: func(ctrl *gomock.Controller) *packageHarness {
				h := newPackageHarness(ctrl)
				h.WriteFile("/etc/apt/sources.list.d/test.list", []byte("repo"), 0644, nil)
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
			err := l.AddRepository(context.Background(), test.args.repository, test.args.filename)
			test.CheckError(t, test.wantErr, err)
		})
	}
}
