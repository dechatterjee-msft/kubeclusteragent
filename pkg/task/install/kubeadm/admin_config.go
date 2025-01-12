package kubeadm

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/task"
)

type AdminConfig struct{}

var _ task.Task = &AdminConfig{}

func NewAdminConfig() *AdminConfig {
	t := &AdminConfig{}
	return t
}

func (t *AdminConfig) Name() string {
	return "set-admin-kubeconfig"
}

func (t *AdminConfig) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())

	// Check if kube-config file exists
	srcPath := constants.KubeadmKubeconfigPath
	exists, err := ou.Filesystem().Exists(ctx, srcPath)
	if err != nil || !exists {
		logger.Error(err, "kube-config file doesn't exits in the given path or the path is not correct", "path", srcPath)
		return err
	}

	// Create .kube directory
	dirPath := constants.AdminKubeconfigDirPath
	err = ou.Filesystem().MkdirAll(ctx, dirPath, constants.DirPerm)
	if err != nil {
		logger.Error(err, "error creating directory")
		return fmt.Errorf("error creating directory: %v", err)
	}
	logger.Info("Directory created", "dir", dirPath)

	// Lookup administrative user by name
	adminUsername := constants.AdminUserName
	adminUser, err := user.Lookup(adminUsername)
	if err != nil {
		logger.Error(err, "error retrieving information for user %w", adminUsername)
	}
	if adminUser == nil {

	}
	// Retrieve UID and GID
	uid := adminUser.Uid
	gid := adminUser.Gid
	UID, err := strconv.ParseInt(uid, constants.Base, constants.BitSize)
	if err != nil {
		logger.Error(err, "error parsing (string)uid to (int)uid")
		return fmt.Errorf("error parsing (string)uid to (int)uid: %v", err)
	}
	GID, err := strconv.ParseInt(gid, constants.Base, constants.BitSize)
	if err != nil {
		logger.Error(err, "error parsing (string)gid to (int)gid")
		return fmt.Errorf("error parsing (string)gid to (int)gid: %v", err)
	}

	// Change ownership of the directory
	err = ou.Filesystem().Chown(ctx, dirPath, int(UID), int(GID))
	if err != nil {
		logger.Error(err, "error changing ownership")
		return fmt.Errorf("error changing ownership: %v", err)
	}
	logger.Info("Ownership changed", "user", adminUsername, "uid", UID, "gid", GID)

	// Open the source file
	src, err := ou.Filesystem().Open(ctx, srcPath)
	defer func(src *os.File) {
		err := src.Close()
		if err != nil {
			logger.Error(err, "error closing source file")
		}
	}(src)
	if err != nil {
		logger.Error(err, "error opening source file")
		return fmt.Errorf("error opening source file: %v", err)
	}
	logger.Info("Opened source file for reading", "Source FilePath", srcPath)

	// Set destination path
	dstPath := filepath.Join(dirPath, constants.KubeconfigFileName)

	// Create the destination file
	dst, err := ou.Filesystem().OpenFileWithPermission(ctx, dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, constants.FilePerm)
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			logger.Error(err, "error closing destination file")
		}
	}(dst)
	if err != nil {
		logger.Error(err, "error creating destination file")
		return fmt.Errorf("error creating destination file: %v", err)
	}
	logger.Info("Created destination file for writing", "Destination FilePath", dstPath)

	// Copy the contents from source to destination
	logger.Info("Copying file", "Source Path", srcPath, "Destination Path", dstPath)
	_, err = ou.Filesystem().Copy(ctx, dst, src)
	if err != nil {
		logger.Error(err, "error copying file")
		return fmt.Errorf("error copying file: %v", err)
	}

	return nil
}

func (t *AdminConfig) Rollback(ctx context.Context, // nolint
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}
