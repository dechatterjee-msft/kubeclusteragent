package cilium

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/osutility"
	"os"
)

func Install(ctx context.Context, logger logr.Logger, version string, ou osutility.OSUtil) error {
	ciliumDownload := fmt.Sprintf("%s/%s/%s", constants.CiliumCLIURL, version, constants.CiliumTarFileName)
	_, err := ou.Filesystem().DownloadFileUsingHttp(ctx, ciliumDownload, fmt.Sprintf("/tmp/%s", constants.CiliumTarFileName), constants.BinaryPermissions)
	if err != nil {
		return err
	}
	err = ou.Filesystem().ExtractTarFile(ctx, fmt.Sprintf("/tmp/%s", constants.CiliumTarFileName), "/tmp")
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
