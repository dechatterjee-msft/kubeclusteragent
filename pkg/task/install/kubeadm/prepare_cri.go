package kubeadm

import (
	"context"
	"fmt"
	"github.com/pelletier/go-toml"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"os"
	"path/filepath"
	"time"
)

type Containerd struct{}

var containerdConfigFile = constants.ConfigFileLocation

func (t *Containerd) Run(ctx context.Context, status cluster.Status, clusterSpec *v1alpha1.ClusterSpec, ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("preparing containerd for kubernetes installation")
	config, err := toml.LoadFile(containerdConfigFile)
	if err != nil {
		logger.Error(err, "unable to find containerd config file at location,generating the config file",
			"Location", constants.ConfigFileLocation)
		dir := filepath.Dir(containerdConfigFile)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Error(err, "error creating directories for containerd config file")
			return err
		}
		if _, err := os.Stat(containerdConfigFile); os.IsNotExist(err) {
			// File does not exist, so create it
			file, err := os.Create(containerdConfigFile)
			if err != nil {
				logger.Error(err, "unbale to create containerd config file")
				return err
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
				}
			}(file) // Ensure the file is closed after creation
			logger.Info("containerd config file generated successfully", "location", containerdConfigFile)
		}
		_, _, err = ou.Exec().Command(ctx, "containerd", nil, "config", "default", ">", containerdConfigFile)
		config, err = toml.LoadFile(containerdConfigFile)
		if err != nil {
			logger.Error(err, "failed to generate containerd config", err)
			return err
		}
	}
	config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", "runc", "options", "SystemdCgroup"}, true)
	//cpImageTags, err := criutil.GetK8sControlPlaneImagesFromPropertiesFile()
	//if err != nil {
	//	logger.Error(err, "unable to find pause image", err)
	//	return err
	//}
	//if cpImageTags[constants.PauseImage] == "" {
	//	err = fmt.Errorf("unable to fetch version for pause image %s", constants.PauseImage)
	//	logger.Error(err, "error occurred while fetching pause image")
	//	return err
	//}
	//pauseImage := fmt.Sprintf("%s:%s", constants.PauseImage, cpImageTags[constants.PauseImage])
	//config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "disable_apparmor"}, true)
	//config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "sandbox_image"}, pauseImage)
	//config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", "runc", "options", "SystemdCgroup"}, true)
	//// Private registry configuration
	//logger.Info("starting private registry configuration")
	//if clusterSpec.ClusterRuntime != nil && clusterSpec.ClusterRuntime.CustomiseClusterRuntime {
	//	logger.Info("Cluster Runtime Info", "ClusterRuntime", clusterSpec.ClusterRuntime)
	//	config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "registry", "config_path"}, "/etc/containerd/certs.d")
	//	var containerdRegistryConfiguration criutil.ContainerRegistry = &criutil.PrivateRegistry{
	//		RegistryFQDN:         clusterSpec.ClusterRuntime.ClusterCri.PrivateRegistryFQDN,
	//		RegistryEndpoint:     fmt.Sprintf("https://%s", clusterSpec.ClusterRuntime.ClusterCri.PrivateRegistryFQDN),
	//		SkipVerify:           clusterSpec.ClusterRuntime.ClusterCri.SkipTls,
	//		RegistryCapabilities: []string{"pull", "resolve"},
	//		Ou:                   ou,
	//		CertFilesLocation:    clusterSpec.ClusterRuntime.ClusterCri.CertFiles,
	//	}
	//	err = containerdRegistryConfiguration.Add(ctx)
	//	if err != nil {
	//		logger.Error(err, "error occurred while adding private registry information to containerd")
	//		return err
	//	}
	//	if clusterSpec.ClusterRuntime.ClusterCri.RegistryAuth.IsAuthRequired {
	//		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "registry", "configs",
	//			clusterSpec.ClusterRuntime.ClusterCri.PrivateRegistryFQDN, "auth", "username"},
	//			clusterSpec.ClusterRuntime.ClusterCri.RegistryAuth.Username)
	//		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "registry", "configs",
	//			clusterSpec.ClusterRuntime.ClusterCri.PrivateRegistryFQDN, "auth", "password"},
	//			clusterSpec.ClusterRuntime.ClusterCri.RegistryAuth.Password)
	//	}
	//}
	marshal, err := toml.Marshal(config)
	if err != nil {
		logger.Error(err, "unable to marshall containerd configuration")
		return err
	}
	_, _, err = ou.Exec().Command(ctx, "mv", nil, containerdConfigFile, fmt.Sprintf("%s.original", containerdConfigFile))
	if err != nil {
		logger.Error(err, "unable to make a copy of containerd configuration file from path", "Location",
			containerdConfigFile, "BackupLocation", fmt.Sprintf("%s.original", containerdConfigFile))
		return err
	}
	err = ou.Filesystem().WriteFile(ctx, containerdConfigFile, marshal, 0644)
	if err != nil {
		logger.Error(err, "unable to write file to destination", "Location", containerdConfigFile)
		return err
	}
	containerdRetryCount := 0
restart:
	err = ou.Systemd().Restart(ctx, "containerd")
	if err != nil {
		logger.Error(err, "error occurred while restarting containerd")
		return err
	}
	time.Sleep(10 * time.Second)
	ok, err := ou.Systemd().IsRunning(ctx, "containerd")
	if err != nil {
		logger.Error(err, "error occurred while checking the status of containerd")
		return err
	}
	if !ok {
		if containerdRetryCount > 3 {
			logger.Error(fmt.Errorf("unbale to start containerd"), "waited for 30 seconds unable to start containerd")
		}
		containerdRetryCount++
		goto restart
	}
	logger.Info("containerd configuration has been updated successfully", "Location", containerdConfigFile)
	return nil
}

func (t *Containerd) Rollback(ctx context.Context, status cluster.Status, clusterSpec *v1alpha1.ClusterSpec, ou osutility.OSUtil) error {
	return nil
}

var _ task.Task = &Containerd{}

func NewPrepareContainerd() *Containerd {
	t := &Containerd{}

	return t
}

func (t *Containerd) Name() string {
	return "prepare-containerd"
}
