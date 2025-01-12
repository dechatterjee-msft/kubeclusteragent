package osutility

import (
	"bytes"
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
)

type Sysctl interface {
	Reload(ctx context.Context) error
	Set(ctx context.Context, values map[string]string) error
}

type FakeSysctl struct{}

var _ Sysctl = &FakeSysctl{}

func NewFakeSysctl() *FakeSysctl {
	f := &FakeSysctl{}

	return f
}

func (f *FakeSysctl) Reload(ctx context.Context) error {
	logger := log.From(ctx)
	logger.Info("Reloading sysctl")

	return nil
}

func (f *FakeSysctl) Set(ctx context.Context, values map[string]string) error {
	logger := log.From(ctx)
	logger.Info("Setting sysctl values", "values", values)

	return nil
}

type LiveSysctl struct {
	exec Exec
	fs   Filesystem
}

var _ Sysctl = &LiveSysctl{}

func NewLiveSysctl(execUtil Exec, fsUtil Filesystem) *LiveSysctl {
	f := &LiveSysctl{
		exec: execUtil,
		fs:   fsUtil,
	}

	return f
}

func (f *LiveSysctl) Reload(ctx context.Context) error {
	logger := log.From(ctx)
	logger.Info("Reloading sysctl")

	code, _, err := f.exec.Command(ctx, "sysctl", nil, "--system")
	if err != nil {
		return fmt.Errorf("reload sysctl: %w", err)
	}

	if code != 0 {
		return fmt.Errorf("unexpected error code: %d", code)
	}

	return nil
}

func (f *LiveSysctl) Set(ctx context.Context, values map[string]string) error {
	logger := log.From(ctx)
	logger.Info("Setting sysctl values", "values", values)

	var buf bytes.Buffer

	for k, v := range values {
		buf.WriteString(fmt.Sprintf("%s = %s\n", k, v))
	}

	if err := f.fs.WriteFile(ctx, "/etc/sysctl.d/99-kubernetes.conf", buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write sysctl: %w", err)
	}

	return nil
}
