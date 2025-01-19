package kubeadm

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"net/http"
	"os/exec"
	"strings"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
)

type Binaries struct {
}

var _ task.Task = &Binaries{}

func NewInstallBinaries() *Binaries {
	t := &Binaries{}

	return t
}

func (t *Binaries) Name() string {
	return "install-binaries"
}

func (t *Binaries) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx)
	logger.Info("Install Kubernetes binaries")
	if clusterSpec.Version == "" {
		return errors.New("no Kubernetes version supplied")
	}
	installed, err := verifyIfKubeadmAlreadyInstalled(ctx, clusterSpec, ou)
	if err != nil {
		logger.Error(err, "error checking if Kubeadm is already installed, proceeding anyway")
	}
	if installed {
		logger.Info("Kubeadm is already installed, skipping installation")
		return nil
	}
	if err := ou.PackageManager().Update(ctx); err != nil {
		return fmt.Errorf("update packages: %w", err)
	}
	stableVersion, err := extractStableVersion(clusterSpec.Version)
	if err != nil {
		return err
	}
	kubernetesRepoKey := fmt.Sprintf("https://pkgs.k8s.io/core:/stable:/v%s/deb/Release.key", stableVersion)
	resp, err := http.Get(kubernetesRepoKey)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	cmd := exec.Command("gpg", "--batch", "--yes", "--dearmor", "-o", "/etc/apt/keyrings/kubernetes-apt-keyring.gpg")
	cmd.Stdin = resp.Body
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// Run the command
	err = cmd.Run()
	if err != nil {
		return err
	}
	if err := ou.PackageManager().Install(ctx, []string{"apt-transport-https", "ca-certificates", "curl"}...); err != nil {
		logger.Error(err, "error installing packages")
	}
	repo := fmt.Sprintf("deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v%s/deb/ /", stableVersion)
	if err := exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo \"%s\" > /etc/apt/sources.list.d/kubernetes.list", repo)).Run(); err != nil {
		logger.Error(err, "error adding Kubernetes repository")
		return err
	}
	if err := exec.Command("sudo", "apt", "update").Run(); err != nil {
		logger.Error(err, "Error updating package list")
		return err
	}
	versions := make(map[string]string)
	packages := []string{"kubeadm", "kubelet", "kubectl"}
	for _, pkg := range packages {
		version, err := getLatestVersion(ctx, clusterSpec.Version, pkg, ou)
		if err != nil {
			logger.Error(err, "error determining version for package", "pkg", pkg)
			return err
		}
		versions[pkg] = version
	}
	packagesWithVersion := make([]string, 0)
	for pkg, version := range versions {
		packagesWithVersion = append(packagesWithVersion, fmt.Sprintf("%s=%s", pkg, version))
	}
	logger.Info("installing packages", "packages", packagesWithVersion)
	if err := ou.PackageManager().Install(ctx, packagesWithVersion...); err != nil {
		logger.Error(err, "error installing packages", "packages", packagesWithVersion)
		return err
	}
	return nil
}

func (t *Binaries) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {

	return nil
}

func (t *Binaries) generatePackageNames(kubernetesVersion string) []string {
	names := []string{"kubelet", "kubeadm", "kubectl"}
	var out = make([]string, 0)
	for _, name := range names {
		out = append(out, fmt.Sprintf("%s=%s-00", name, kubernetesVersion))
	}

	return out
}

func extractStableVersion(fullVersion string) (string, error) {
	// Split the version string into components
	fullVersion = convertKubernetesVersionContainsV(fullVersion)
	parts := strings.Split(fullVersion, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid version format: %s", fullVersion)
	}
	// Combine the major and minor parts
	stableVersion := fmt.Sprintf("%s.%s", parts[0], parts[1])
	return stableVersion, nil
}

// Get the latest subversion for a specific Kubernetes version
func getLatestVersion(ctx context.Context, version string, packageName string, ou osutility.OSUtil) (string, error) {
	version = convertKubernetesVersionContainsV(version)
	code, output, err := ou.Exec().Command(ctx, "apt-cache", nil, []string{"madison", packageName}...)
	if err != nil {
		return "", fmt.Errorf("error querying versions for %s: %v, code:%d", packageName, err, code)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, version) {
			// Extract the full version (e.g., "1.31.4-1.1")
			fields := strings.Fields(line)
			if len(fields) > 2 {
				return fields[2], nil
			}
		}
	}
	return "", fmt.Errorf("no matching version found for %s=%s", packageName, version)
}

func convertKubernetesVersionContainsV(kubernetesVersion string) string {
	if strings.Contains(kubernetesVersion, "v") {
		split := strings.Split(kubernetesVersion, "v")
		kubernetesVersion = split[1]
	}
	return kubernetesVersion
}

func verifyIfKubeadmAlreadyInstalled(ctx context.Context, clusterSpec *v1alpha1.ClusterSpec, ou osutility.OSUtil) (bool, error) {
	version, err := ou.Kubeadm().Version(ctx)
	if err != nil {
		return false, err
	}
	return version == clusterSpec.Version, nil
}
