package k3s

import (
	"context"
	"encoding/base64"
	"fmt"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"

	"kubeclusteragent/pkg/task"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
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
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithValues(
		"Cluster Type", clusterSpec.ClusterType,
		"Version", clusterSpec.Version,
		"Operation", "Install", "Task", t.Name())

	cniManifestURL := clusterSpec.Networking.GetCniManifestURL()
	if cniManifestURL == "" {
		logger.Info("skipping CNI installation as no CNI manifest found in the spec")
		return nil
	}
	logger.Info("Starting CNI installation")
	err := installManifestOrURL(ctx, "", constants.CNIManifestFilePath, cniManifestURL, ou)
	if err != nil {
		logger.Error(err, "error occurred while installing CNI")
		return fmt.Errorf("error installing CNI: %w", err)
	}
	logger.Info("CNI installation successful")

	//metaCniManifestURL := clusterSpec.Networking.GetMetaCNIManifestURL()
	//metaCniManifest := clusterSpec.Networking.GetMetaCNIManifest()
	//if metaCniManifest != "" || metaCniManifestURL != "" {
	//	logger.Info("Starting metaCNI installation")
	//	err := installManifestOrURL(ctx, cniManifest, constants.CNIManifestFilePath, cniManifestURL, ou)
	//	if err != nil {
	//		logger.Error(err, "error occurred while installing metaCNI")
	//		return fmt.Errorf("error installing metaCNI: %w", err)
	//	}
	//	logger.Info("metaCNI installation successful")
	//}

	return nil
}

func (t *Cni) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}

func installManifestOrURL(ctx context.Context, manifestData, manifestFilePath, url string, ou linux.OSUtil) error {
	logger := log.From(ctx)
	var kubectlClient linux.Kubectl = &linux.K3sLiveKubectl{}
	manifestToApply := ""
	if url != "" {
		manifestToApply = url
	} else {
		decodedManifest, err := base64.StdEncoding.DecodeString(manifestData)
		if err != nil {
			logger.Error(err, "error decoding manifest")
			return fmt.Errorf("error decoding manifest: %w", err)
		}
		manifestContents := string(decodedManifest)
		err = ou.Filesystem().WriteFile(ctx, manifestFilePath, []byte(manifestContents), 0644)
		if err != nil {
			logger.Error(err, "error writing manifest to file")
			return fmt.Errorf("error writing manifest to file: %w", err)
		}
		manifestToApply = manifestFilePath
	}
	output, err := kubectlClient.RunWithResponse(ctx, []string{"apply", "-f", manifestToApply}...)
	if err != nil {
		logger.Error(err, "error installing manifest", "output", output)
		return fmt.Errorf("error installing manifest:%s %w", manifestToApply, err)
	}
	return nil
}
