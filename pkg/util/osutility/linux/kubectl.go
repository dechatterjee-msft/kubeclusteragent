package linux

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"strings"
)

type Kubectl interface {
	Run(ctx context.Context, cmdArgs ...string) error
	RunWithResponse(ctx context.Context, cmdArgs ...string) (string, error)
}

type LiveKubectl struct {
	cmd Exec
}

type K3sLiveKubectl struct {
	cmd Exec
}

type FakeKubectl struct{}
type FakeKubectlError struct{}

func (f *FakeKubectl) Run(ctx context.Context, cmdArgs ...string) error {
	return nil
}

func (f *FakeKubectl) RunWithResponse(ctx context.Context, cmdArgs ...string) (string, error) {
	args := strings.Join(cmdArgs, ",")
	if strings.Contains(args, "taint") {
		return "untainted", nil
	}
	if strings.Contains(args, "nodes") && strings.Contains(args, "get") {
		return "SchedulingDisabled", nil
	}
	return "", nil
}

func (f *FakeKubectlError) Run(ctx context.Context, cmdArgs ...string) error {
	return fmt.Errorf("testutil error")
}

func (f *FakeKubectlError) RunWithResponse(ctx context.Context, cmdArgs ...string) (string, error) {
	return "", fmt.Errorf("testutil error")
}

func NewFakeKubectl() *FakeKubectl {
	return &FakeKubectl{}
}

func NewLiveKubectl(cmd Exec) *LiveKubectl {
	return &LiveKubectl{
		cmd: cmd,
	}
}

func (l *LiveKubectl) Run(ctx context.Context, cmdArgs ...string) error {
	logger := log.From(ctx).WithName("kubectl")
	cmdArgs = append(cmdArgs, "--kubeconfig", "/etc/kubernetes/admin.conf")
	_, data, err := l.cmd.Command(ctx, "kubectl", nil, cmdArgs...)
	if err != nil || strings.Contains(string(data), "error") {
		return fmt.Errorf("run kubectl: %s", string(data))
	}
	logger.Info(string(data))
	return nil
}

func (l *LiveKubectl) RunWithResponse(ctx context.Context, cmdArgs ...string) (string, error) {
	cmdArgs = append(cmdArgs, "--kubeconfig", "/etc/kubernetes/admin.conf")
	_, data, err := l.cmd.Command(ctx, "kubectl", nil, cmdArgs...)
	if err != nil || strings.Contains(string(data), "error") {
		return string(data), fmt.Errorf("run kubectl: %s", string(data))
	}
	return string(data), nil
}

func (l *K3sLiveKubectl) Run(ctx context.Context, cmdArgs ...string) error {
	logger := log.From(ctx).WithName("k3s-kubectl")
	kubectlArgs := []string{"kubectl"}
	kubectlArgs = append(kubectlArgs, cmdArgs...)
	_, data, err := l.cmd.Command(ctx, "k3s", nil, kubectlArgs...)
	if err != nil || strings.Contains(string(data), "error") {
		return fmt.Errorf("run k3s kubectl: %s", string(data))
	}
	logger.Info(string(data))
	return nil
}

func (l *K3sLiveKubectl) RunWithResponse(ctx context.Context, cmdArgs ...string) (string, error) {
	logger := log.From(ctx).WithName("k3s-kubectl")
	kubectlArgs := []string{"kubectl"}
	kubectlArgs = append(kubectlArgs, cmdArgs...)
	_, data, err := l.cmd.Command(ctx, "k3s", nil, kubectlArgs...)
	if err != nil || strings.Contains(string(data), "error") {
		return string(data), fmt.Errorf("run k3s kubectl: %s", string(data))
	}
	logger.Info(string(data))
	return string(data), nil
}
