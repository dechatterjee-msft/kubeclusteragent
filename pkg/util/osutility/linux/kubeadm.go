package linux

import (
	"context"
	"fmt"
	"go.uber.org/multierr"
	"kubeclusteragent/pkg/constants"
	"math"
	"runtime"
	"strconv"
	"strings"
)

type Kubeadm interface {
	Run(ctx context.Context, cmdArgs ...string) error
	RunWithResponse(ctx context.Context, cmdArgs ...string) (string, error)
	CertsRotateAll(ctx context.Context) (string, error)
	CertsRotate(ctx context.Context, certName string) (string, error)
	GetCertsExpiry(ctx context.Context) (int, map[string]int64, error)
	Install(ctx context.Context, configFilePath string) (string, error)
	Upgrade(ctx context.Context, version string, ignorePreflight string) (string, error)
	Delete(ctx context.Context) (string, error)
	Version(ctx context.Context) (string, error)
}

type LiveKubeadm struct {
	cmd Exec
}

type FakeKubeadm struct{}

func NewFakeKubeadm() *FakeKubeadm {
	return &FakeKubeadm{}
}

func (f FakeKubeadm) Run(ctx context.Context, cmdArgs ...string) error {
	return nil
}

func (f FakeKubeadm) RunWithResponse(ctx context.Context, cmdArgs ...string) (string, error) {
	return "success", nil
}

func (f FakeKubeadm) CertsRotateAll(ctx context.Context) (string, error) {
	return "", nil
}

func (f FakeKubeadm) CertsRotate(ctx context.Context, certName string) (string, error) {
	return "", nil
}

func (f FakeKubeadm) GetCertsExpiry(ctx context.Context) (int, map[string]int64, error) {

	testMap := make(map[string]int64)
	testMap["admin.conf"] = 363
	testMap["apiserver"] = 363
	testMap["apiserver-etcd-client"] = 363
	testMap["apiserver-kubelet-client"] = 363
	testMap["controller-manager.conf "] = 363
	testMap["etcd-healthcheck-client"] = 363
	testMap["etcd-peer"] = 363
	testMap["etcd-server"] = 363
	testMap["front-proxy-client"] = 363
	testMap["scheduler.conf"] = 363

	return 363, testMap, nil
}

func (f FakeKubeadm) Install(ctx context.Context, configFilePath string) (string, error) {
	return "", nil
}

func (f FakeKubeadm) Upgrade(ctx context.Context, version string, ignorePreflight string) (string, error) {
	banner := fmt.Sprintf("%s \"%s\"", constants.KubeadmClusterSuccessfulUpgradeBanner, version)
	return banner, nil
}

func (f FakeKubeadm) Delete(ctx context.Context) (string, error) {
	return "", nil
}

func (f FakeKubeadm) Version(ctx context.Context) (string, error) {
	return "", nil
}

func NewLiveKubeadm(cmd Exec) *LiveKubeadm {
	return &LiveKubeadm{
		cmd: cmd,
	}
}

func (l LiveKubeadm) Run(ctx context.Context, cmdArgs ...string) error {
	_, _, err := l.cmd.Command(ctx, "kubeadm", nil, cmdArgs...)
	if err != nil {
		return err
	}
	return nil
}

func (l LiveKubeadm) RunWithResponse(ctx context.Context, cmdArgs ...string) (string, error) {
	_, i, err := l.cmd.Command(ctx, "kubeadm", nil, cmdArgs...)
	if err != nil {
		return "", err
	}
	return string(i), nil
}

func (l LiveKubeadm) CertsRotateAll(ctx context.Context) (string, error) {
	_, i, err := l.cmd.Command(ctx, "kubeadm", nil, []string{"certs", "renew", "all"}...)
	if err != nil {
		return "", err
	}
	return string(i), nil
}

func (l LiveKubeadm) CertsRotate(ctx context.Context, certName string) (string, error) {
	_, i, err := l.cmd.Command(ctx, "kubeadm", nil, []string{"certs", "renew", certName}...)
	if err != nil {
		return "", err
	}
	return string(i), nil
}

func (l LiveKubeadm) GetCertsExpiry(ctx context.Context) (int, map[string]int64, error) {
	_, i, err := l.cmd.Command(ctx, "kubeadm", nil, []string{"certs", "check-expiration"}...)
	if err != nil {
		return 0, nil, err
	}
	return evaluateOverallCertsExpiration(string(i))
}

func (l LiveKubeadm) Install(ctx context.Context, configFilename string) (string, error) {
	var cmdArgs []string
	if runtime.NumCPU() < 2 {
		cmdArgs = append(cmdArgs, "init", "--config", configFilename, "--ignore-preflight-errors=NumCPU")
	} else {
		cmdArgs = append(cmdArgs, "init", "--config", configFilename)
	}
	_, output, err := l.cmd.Command(ctx, "kubeadm", nil, cmdArgs...)
	if err != nil {
		return "", fmt.Errorf("run kubeadm: %w,output :%s", err, string(output))
	}
	return string(output), nil
}

func (l LiveKubeadm) Upgrade(ctx context.Context, version string, ignorePreflight string) (string, error) {
	var out []byte
	var err error
	var code int
	upgradeArgs := make([]string, 0)
	upgradeArgs = append(upgradeArgs, "upgrade", "apply", version, "-y")
	if ignorePreflight != "" {
		ignorePreflightArgs := fmt.Sprintf("--ignore-preflight-errors=%s", ignorePreflight)
		upgradeArgs = append(upgradeArgs, ignorePreflightArgs)
	}
	code, out, err = l.cmd.Command(ctx, "kubeadm", nil, upgradeArgs...)
	if err != nil || code != 0 {
		return "", multierr.Append(fmt.Errorf("%s", string(out)), err)
	}
	return string(out), nil
}

func (l LiveKubeadm) Delete(ctx context.Context) (string, error) {
	return "", nil
}

func evaluateOverallCertsExpiration(expiryInfo string) (int, map[string]int64, error) {
	p := strings.Split(expiryInfo, "\n")
	allCertsExpiryInfo := make(map[string]int64)
	var err error
	overallTimeExpiration := math.MaxInt
	for i := 0; i < len(p); i++ {
		t := strings.Fields(p[i])
		if len(t) > 6 &&
			t[0] != "CERTIFICATE" &&
			t[0] != "[check-expiration]" &&
			strings.HasSuffix(t[6], "d") {
			var residualTimeInteger int
			residualTime := strings.Trim(t[6], "d")
			residualTimeInteger, err = strconv.Atoi(residualTime)
			if err != nil {
				err = multierr.Append(err, fmt.Errorf("certs:%s,expiry:%s not able to evaluate", t[0], residualTime))
				continue
			}
			overallTimeExpiration = min(overallTimeExpiration, residualTimeInteger)
			allCertsExpiryInfo[t[0]] = int64(residualTimeInteger)
		}
	}
	return overallTimeExpiration, allCertsExpiryInfo, err
}

func (l LiveKubeadm) Version(ctx context.Context) (string, error) {
	code, output, err := l.cmd.Command(ctx, "kubeadm", nil, []string{"version", "-o", "short"}...)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("failed with error code %d", code)
	}
	return strings.TrimSpace(string(output)), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
