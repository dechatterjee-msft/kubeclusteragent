package osutility

import (
	"context"
	"errors"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"os"
	"path/filepath"
)

type FakeDnfPackageManager struct {
	exec Exec
	fs   Filesystem
}

var _ PackageManagerFactory = &FakeDnfPackageManager{}

func NewDnfFakePackage() *FakeDnfPackageManager {
	f := &FakeDnfPackageManager{}
	return f
}

func (f *FakeDnfPackageManager) Uninstall(ctx context.Context, packageNames ...string) error {
	return nil
}

func (f *FakeDnfPackageManager) CheckInstalled(ctx context.Context, packageName string) bool {
	logger := log.From(ctx)
	logger.Info("Check if package is installed", "packageName", packageName)
	return packageName != ""
}

func (f *FakeDnfPackageManager) Install(ctx context.Context, packageNames ...string) error {
	logger := log.From(ctx)

	for _, packageName := range packageNames {
		logger.Info("Installing package", "packageName", packageName)
	}

	return nil
}

func (f *FakeDnfPackageManager) Update(ctx context.Context) error {
	logger := log.From(ctx)
	logger.Info("Updating packages")

	return nil
}

func (f *FakeDnfPackageManager) AddKey(ctx context.Context, urlStr string) error {
	logger := log.From(ctx)
	logger.Info("Adding signing key", "url", urlStr)

	return nil
}

func (f *FakeDnfPackageManager) AddRepository(ctx context.Context, repository, filename string) error {
	logger := log.From(ctx)
	logger.Info("Adding package repository", "repository", repository, "filename", filename)

	return nil
}

func (f *FakeDnfPackageManager) RemoveRepository(ctx context.Context, repository, filename string) error {
	logger := log.From(ctx)
	logger.Info("Adding package repository", "repository", repository, "filename", filename)

	return nil
}

type LiveDnfPackageManager struct {
	exec Exec
	fs   Filesystem
}

var _ PackageManagerFactory = &LiveDnfPackageManager{}

func NewDnfLivePackageManager(execUtil Exec, fsUtil Filesystem) *LiveDnfPackageManager {
	f := &LiveDnfPackageManager{
		exec: execUtil,
		fs:   fsUtil,
	}

	return f
}

func NewDnfFakeManager(execUtil Exec, fsUtil Filesystem) *FakeDnfPackageManager {
	f := &FakeDnfPackageManager{
		exec: execUtil,
		fs:   fsUtil,
	}

	return f
}

func (f *LiveDnfPackageManager) CheckInstalled(ctx context.Context, packageName string) bool {
	logger := log.From(ctx)

	if packageName == "" {
		logger.Info("Check package installed called with blank package name")
		return false
	}
	logger.Info("Check if package is installed", "packageName", packageName)
	code, _, err := f.exec.Command(ctx, "dpkg", nil, "-l", packageName)
	if err != nil {
		return false
	}

	return code == 0
}

func (f *LiveDnfPackageManager) Install(ctx context.Context, packageNames ...string) error {
	logger := log.From(ctx)

	logger.Info("Installing packages", "packageNames", packageNames)

	_, _, err := f.exec.Command(ctx, "apt-mark", nil, append([]string{"unhold"}, packageNames...)...)
	if err != nil {
		return fmt.Errorf("unholding packages: %w", err)
	}

	_, _, err = f.exec.Command(ctx, "apt-get", append(os.Environ(), "DEBIAN_FRONTEND=noninteractive"), append([]string{"install", "-y"}, packageNames...)...)
	if err != nil {
		return fmt.Errorf("install packages: %w", err)
	}

	_, _, err = f.exec.Command(ctx, "apt-mark", nil, append([]string{"hold"}, packageNames...)...)
	if err != nil {
		return fmt.Errorf("holding packages: %w", err)
	}

	return nil
}

func (f *LiveDnfPackageManager) Uninstall(ctx context.Context, packageNames ...string) error {
	logger := log.From(ctx)

	logger.Info("Uninstalling packages", "packageNames", packageNames)

	_, _, err := f.exec.Command(ctx, "apt-mark", nil, append([]string{"unhold"}, packageNames...)...)
	if err != nil {
		return fmt.Errorf("unholding packages: %w", err)
	}

	_, _, err = f.exec.Command(ctx, "apt-get", append(os.Environ(), "DEBIAN_FRONTEND=noninteractive"), append([]string{"uninstall", "-y"}, packageNames...)...)
	if err != nil {
		return fmt.Errorf("uninstall packages: %w", err)
	}

	return nil
}

func (f *LiveDnfPackageManager) Update(ctx context.Context) error {
	logger := log.From(ctx)
	logger.Info("Updating packages")

	code, output, err := f.exec.Command(ctx, "apt-get", nil, "update")
	if err != nil {
		logger.Info("Update packages output", "output", string(output))
		return fmt.Errorf("update packages: %w", err)
	}

	if code != 0 {
		return fmt.Errorf("unexpected error code: %d", code)
	}

	return nil
}

func (f *LiveDnfPackageManager) AddKey(ctx context.Context, urlStr string) error {
	logger := log.From(ctx)

	if urlStr == "" {
		return errors.New("no key url")
	}

	logger.Info("Adding signing key", "url", urlStr)

	cmd := fmt.Sprintf("wget -qO - %s | sudo apt-key add -", urlStr)

	code, _, err := f.exec.Command(ctx, "bash", nil, "-c", cmd)
	if err != nil {
		return fmt.Errorf("update packages: %w", err)
	}

	if code != 0 {
		return fmt.Errorf("unexpected error code: %d", code)
	}

	return nil
}

func (f *LiveDnfPackageManager) AddRepository(ctx context.Context, repository, filename string) error {
	logger := log.From(ctx)
	logger.Info("Adding package repository", "repository", repository, "filename", filename)
	dest := filepath.Join("/etc/apt/sources.list.d", filename+".list")
	if err := f.fs.WriteFile(ctx, dest, []byte(repository), 0644); err != nil {
		return fmt.Errorf("write repository source: %w", err)
	}
	return nil
}

func (f *LiveDnfPackageManager) RemoveRepository(ctx context.Context, repository, filename string) error {
	logger := log.From(ctx)
	logger.Info("Removing package repository", "repository", repository, "filename", filename)
	dest := filepath.Join("/etc/apt/sources.list.d", filename+".list")
	if err := f.fs.RemoveAll(ctx, dest); err != nil {
		return fmt.Errorf("delete repository source: %w", err)
	}
	return nil
}
