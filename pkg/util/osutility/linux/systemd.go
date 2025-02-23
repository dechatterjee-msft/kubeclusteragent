package linux

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"strings"
)

type Systemd interface {
	IsRunning(ctx context.Context, name string) (bool, error)
	IsRunningNoLogging(ctx context.Context, name string) (bool, error)
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Restart(ctx context.Context, name string) error
	Reload(ctx context.Context, name string) error
	DaemonReload(ctx context.Context) error
}

type FakeSystemd struct{}

var _ Systemd = &FakeSystemd{}

func NewFakeSystemd() *FakeSystemd {
	f := &FakeSystemd{}

	return f
}

func (f *FakeSystemd) IsRunning(ctx context.Context, name string) (bool, error) {
	// systemctl show -p ActiveState --value x11-common

	logger := log.From(ctx)
	logger.Info("Check if systemd service is running", "name", name)

	return true, nil
}
func (f *FakeSystemd) IsRunningNoLogging(ctx context.Context, name string) (bool, error) {
	return true, nil
}

func (f *FakeSystemd) Start(ctx context.Context, name string) error {
	logger := log.From(ctx)
	logger.Info("Start systemd service", "name", name)

	return nil
}

func (f *FakeSystemd) Stop(ctx context.Context, name string) error {
	logger := log.From(ctx)
	logger.Info("Stop systemd service", "name", name)

	return nil
}

func (f *FakeSystemd) Restart(ctx context.Context, name string) error {
	logger := log.From(ctx)
	logger.Info("Restart systemd service", "name", name)

	return nil
}

func (f *FakeSystemd) Reload(ctx context.Context, name string) error {
	logger := log.From(ctx)
	logger.Info("Reload systemd service", "name", name)

	return nil
}

func (f *FakeSystemd) DaemonReload(ctx context.Context) error {
	logger := log.From(ctx)
	logger.Info("Reload systemd service", "daemon-reload")

	return nil

}

type LiveSystemd struct {
	exec Exec
}

var _ Systemd = &LiveSystemd{}

func NewLiveSystemd(execUtil Exec) *LiveSystemd {
	f := &LiveSystemd{
		exec: execUtil,
	}

	return f
}

func (f *LiveSystemd) IsRunning(ctx context.Context, name string) (bool, error) {
	// systemctl show -p ActiveState --value x11-common

	logger := log.From(ctx)
	logger.Info("Check if systemd service is running", "name", name)

	code, output, err := f.exec.Command(ctx, "systemctl", nil, "show", "-p", "ActiveState", "--value", name)
	if err != nil {
		return false, fmt.Errorf("run systemctl: %w", err)
	}

	if code != 0 {
		return false, fmt.Errorf("invalid return code %d", code)
	}
	// Trim space and new lines in the output
	statusOutput := string(output)
	statusOutput = strings.Trim(statusOutput, "\n")
	statusOutput = strings.TrimSpace(statusOutput)
	if statusOutput == "active" {
		return true, nil
	}

	return false, nil
}

func (f *LiveSystemd) IsRunningNoLogging(ctx context.Context, name string) (bool, error) {
	code, output, err := f.exec.CommandWithNoLogging(ctx, "systemctl", nil, "show", "-p", "ActiveState", "--value", name)
	if err != nil {
		return false, fmt.Errorf("run systemctl: %w", err)
	}

	if code != 0 {
		return false, fmt.Errorf("invalid return code %d", code)
	}
	// Trim space and new lines in the output
	statusOutput := string(output)
	statusOutput = strings.Trim(statusOutput, "\n")
	statusOutput = strings.TrimSpace(statusOutput)
	if statusOutput == "active" {
		return true, nil
	}

	return false, nil
}

func (f *LiveSystemd) Start(ctx context.Context, name string) error {
	logger := log.From(ctx)
	logger.Info("Start systemd service", "name", name)

	code, _, err := f.exec.Command(ctx, "systemctl", nil, "start", name)
	if err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	if code != 0 {
		return fmt.Errorf("invalid return code %d", code)
	}

	return nil
}

func (f *LiveSystemd) Stop(ctx context.Context, name string) error {
	logger := log.From(ctx)
	logger.Info("Stop systemd service", "name", name)

	code, _, err := f.exec.Command(ctx, "systemctl", nil, "stop", name)
	if err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	if code != 0 {
		return fmt.Errorf("invalid return code %d", code)
	}

	return nil
}

func (f *LiveSystemd) Restart(ctx context.Context, name string) error {
	logger := log.From(ctx)
	logger.Info("Restart systemd service", "name", name)

	code, _, err := f.exec.Command(ctx, "systemctl", nil, "restart", name)
	if err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	if code != 0 {
		return fmt.Errorf("invalid return code %d", code)
	}

	return nil
}

func (f *LiveSystemd) Reload(ctx context.Context, name string) error {
	logger := log.From(ctx)
	logger.Info("Reload systemd service", "name", name)

	code, _, err := f.exec.Command(ctx, "systemctl", nil, "reload", name)
	if err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	if code != 0 {
		return fmt.Errorf("invalid return code %d", code)
	}

	return nil
}

func (f *LiveSystemd) DaemonReload(ctx context.Context) error {
	logger := log.From(ctx)
	logger.Info("Reload systemd service", "name", "daemon reload")

	code, _, err := f.exec.Command(ctx, "systemctl", nil, "daemon-reload")
	if err != nil {
		return fmt.Errorf("start %s: %w", "daemon-reload", err)
	}

	if code != 0 {
		return fmt.Errorf("invalid return code %d", code)
	}

	return nil
}
