package osutility_test

import (
	"context"
	"kubeclusteragent/pkg/util/osutility"
	"kubeclusteragent/pkg/util/test"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"kubeclusteragent/mocks"
)

type systemdHarness struct {
	exec    *mocks.MockExec
	systemd *osutility.LiveSystemd
}

func newSystemHarness(ctrl *gomock.Controller) *systemdHarness {
	exec := mocks.NewMockExec(ctrl)
	h := &systemdHarness{
		exec:    exec,
		systemd: osutility.NewLiveSystemd(exec),
	}

	return h
}

func (h *systemdHarness) ExpectCommand(name string, env, args []string, ret []any) {
	h.exec.EXPECT().Command(gomock.Any(), name, env, args).Return(ret...)
}

func TestLiveSystemd_IsRunning(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		harness func(ctrl *gomock.Controller) *systemdHarness
		want    bool
		wantErr bool
	}{
		{
			name: "active",
			args: args{
				name: "name",
			},
			harness: func(ctrl *gomock.Controller) *systemdHarness {
				h := newSystemHarness(ctrl)
				h.ExpectCommand("systemctl",
					nil,
					[]string{"show", "-p", "ActiveState", "--value", "name"},
					[]any{0, []byte("active"), nil})
				return h
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "active",
			args: args{
				name: "name",
			},
			harness: func(ctrl *gomock.Controller) *systemdHarness {
				h := newSystemHarness(ctrl)
				h.ExpectCommand("systemctl",
					nil,
					[]string{"show", "-p", "ActiveState", "--value", "name"},
					[]any{0, []byte("inactive"), nil})
				return h
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)

			l := h.systemd
			got, err := l.IsRunning(context.Background(), test.args.name)
			test.CheckError(t, test.wantErr, err, func() {
				require.Equal(t, test.want, got)
			})
		})
	}
}

func TestLiveSystemd_Start(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		harness func(ctrl *gomock.Controller) *systemdHarness
		want    bool
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				name: "name",
			},
			harness: func(ctrl *gomock.Controller) *systemdHarness {
				h := newSystemHarness(ctrl)
				h.ExpectCommand("systemctl",
					nil,
					[]string{"start", "name"},
					[]any{0, nil, nil})
				return h
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)

			l := h.systemd
			err := l.Start(context.Background(), test.args.name)
			test.CheckError(t, test.wantErr, err)
		})
	}
}

func TestLiveSystemd_Stop(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		harness func(ctrl *gomock.Controller) *systemdHarness
		want    bool
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				name: "name",
			},
			harness: func(ctrl *gomock.Controller) *systemdHarness {
				h := newSystemHarness(ctrl)
				h.ExpectCommand("systemctl",
					nil,
					[]string{"stop", "name"},
					[]any{0, nil, nil})
				return h
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)

			l := h.systemd
			err := l.Stop(context.Background(), test.args.name)
			test.CheckError(t, test.wantErr, err)
		})
	}
}

func TestLiveSystemd_Restart(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		harness func(ctrl *gomock.Controller) *systemdHarness
		want    bool
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				name: "name",
			},
			harness: func(ctrl *gomock.Controller) *systemdHarness {
				h := newSystemHarness(ctrl)
				h.ExpectCommand("systemctl",
					nil,
					[]string{"restart", "name"},
					[]any{0, nil, nil})
				return h
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)

			l := h.systemd
			err := l.Restart(context.Background(), test.args.name)
			test.CheckError(t, test.wantErr, err)
		})
	}
}

func TestLiveSystemd_Reload(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		harness func(ctrl *gomock.Controller) *systemdHarness
		want    bool
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				name: "name",
			},
			harness: func(ctrl *gomock.Controller) *systemdHarness {
				h := newSystemHarness(ctrl)
				h.ExpectCommand("systemctl",
					nil,
					[]string{"reload", "name"},
					[]any{0, nil, nil})
				return h
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := test.harness(ctrl)

			l := h.systemd
			err := l.Reload(context.Background(), test.args.name)
			test.CheckError(t, test.wantErr, err)
		})
	}
}
