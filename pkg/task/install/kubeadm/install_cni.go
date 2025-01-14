// nolint
package kubeadm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"net/http"
	"os"

	"path/filepath"
)

type Cni struct{}

var _ task.Task = &Cni{}

func NewInstallCNI() *Cni {
	t := &Cni{}
	return t
}

func (t *Cni) Name() string {
	return "install-cni"
}

func (t *Cni) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("installing cni")
	if clusterSpec.Networking.CniName == "Calico" {
		// TODO install calico
		logger.Info("cni is calico")
	}
	if clusterSpec.Networking.CniName == "Cilium" {
		return installCilium(ctx, logger, clusterSpec.Networking.CniVersion, ou)
	}
	response, err := ou.Kubectl().RunWithResponse(ctx, "apply", "-f", clusterSpec.Networking.CniManifestURL)
	if err != nil {
		logger.Error(err, "unable to apply CNI present in the give location", "path", clusterSpec.Networking.CniManifestURL)
		return err
	}
	logger.Info("cni installation response", "response", response)
	return nil
}

func (t *Cni) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}

func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func installBinary(filePath, installPath string) error {
	err := os.Rename(filePath, installPath)
	if err != nil {
		return fmt.Errorf("failed to move binary: %w", err)
	}
	err = os.Chmod(installPath, constants.BinaryPermissions)
	if err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}
	return nil
}

func installCilium(ctx context.Context, logger logr.Logger, version string, ou osutility.OSUtil) error {
	ciliumDownload := fmt.Sprintf("%s/%s/%s", constants.CiliumCLIURL, version, constants.CiliumTarFileName)
	err := downloadFile(ciliumDownload, fmt.Sprintf("/tmp/%s", constants.CiliumTarFileName))
	if err != nil {
		return err
	}
	err = ExtractTarGz(fmt.Sprintf("/tmp/%s", constants.CiliumTarFileName), "/tmp")
	if err != nil {
		return err
	}
	err = os.Chmod("/tmp/cilium", constants.BinaryPermissions)
	code, _, err := ou.Exec().Command(ctx, "mv", nil, []string{"/tmp/cilium", "/usr/local/bin/"}...)
	if err != nil {
		logger.Error(fmt.Errorf("installation of cilium failed"), "failed to move cilium", "path", "/tmp/cilium", "to", "/usr/local/bin/")
		return err
	}
	if code != 0 {
		logger.Error(fmt.Errorf("installation of cilium failed, code: %d", code), "failed to move cilium", "path", "/tmp/cilium", "to", "/usr/local/bin/")
		return fmt.Errorf("installation of cilium failed, code: %d", code)
	}
	code, out, err := ou.Exec().Command(ctx, "cilium", nil, []string{"install", "--kubeconfig", constants.KubeadmKubeconfigPath}...)
	if err != nil {
		return err
	}
	if code != 0 {
		logger.Error(fmt.Errorf("cilium installation failed %s", out), "failed to install cilium", "version", "1.16.5")
		return fmt.Errorf("installation of cilium failed, code: %d", code)
	}
	logger.Info("cilium successfully installed", "version", "1.16.5", "out", string(out))
	return nil
}

func ExtractTarGz(src string, dest string) error {
	// Open the tar.gz file
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Create a gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of archive
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %v", err)
		}

		// Determine the output path
		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create a directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			// Create a regular file
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to copy file content: %v", err)
			}
			outFile.Close()
		default:
			// Handle other file types if necessary
			return fmt.Errorf("unsupported file type: %v", header.Typeflag)
		}
	}
	return nil
}
