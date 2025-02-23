package linux

import (
	"context"
	"errors"
	"kubeclusteragent/pkg/util/log/log"
	"os/exec"
	"syscall"
)

type Exec interface {
	Command(ctx context.Context, name string, env []string, args ...string) (int, []byte, error)
	CommandWithNoLogging(ctx context.Context, name string, env []string, args ...string) (int, []byte, error)
}

type FakeExec struct{}

var _ Exec = &FakeExec{}

func NewFakeExec() *FakeExec {
	f := &FakeExec{}

	return f
}

func (f *FakeExec) Command(ctx context.Context, name string, env []string, args ...string) (int, []byte, error) {
	logger := log.From(ctx)
	logger.Info("Running command", "name", name, "arg", args)
	logger.Info("Your Kubernetes control-plane has initialized successfully!")
	output := `Your Kubernetes control-plane has initialized successfully!`
	return 0, []byte(output), nil
}

func (f *FakeExec) CommandWithNoLogging(ctx context.Context, name string, env []string, args ...string) (int, []byte, error) {
	logger := log.From(ctx)
	logger.Info("Running command", "name", name, "arg", args)

	cmd := exec.Command(name, args...)
	cmd.Env = env

	data, err := cmd.CombinedOutput()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), data, nil
			}
		}

		return -1, data, err
	}

	return 0, data, nil
}

type LiveExec struct{}

var _ Exec = &LiveExec{}

func NewLiveExec() *LiveExec {
	l := &LiveExec{}

	return l
}

func (l LiveExec) Command(ctx context.Context, name string, env []string, args ...string) (int, []byte, error) {
	logger := log.From(ctx)
	logger.Info("Running command", "name", name, "arg", args)

	cmd := exec.Command(name, args...)
	cmd.Env = env

	data, err := cmd.CombinedOutput()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), data, nil
			}
		}

		return -1, data, err
	}

	return 0, data, nil
}

func (l LiveExec) CommandWithNoLogging(ctx context.Context, name string, env []string, args ...string) (int, []byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = env

	data, err := cmd.CombinedOutput()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), data, nil
			}
		}

		return -1, data, err
	}

	return 0, data, nil
}
