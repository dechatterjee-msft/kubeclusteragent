package cri

import (
	"context"
	"github.com/pelletier/go-toml"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"path/filepath"
	"sync"
)

type PrivateRegistry struct {
	RegistryFQDN         string
	RegistryEndpoint     string
	SkipVerify           bool
	RegistryCapabilities []string
	CertFilesLocation    []string
	Ou                   linux.OSUtil
}

var containerdRootDirectory = constants.ContainerdRootDirectory
var mutex = &sync.Mutex{}

type ContainerRegistry interface {
	Add(ctx context.Context) error
	Del(ctx context.Context) error
	Get(ctx context.Context) ([]byte, error)
	Update(ctx context.Context) error
}

// Add will create the hosts file and the directory name with FQDN
func (a PrivateRegistry) Add(ctx context.Context) error {
	mutex.Lock()
	defer mutex.Unlock()
	logger := log.From(ctx)
	registryRootDirectory := filepath.Join(containerdRootDirectory, "certs.d", a.RegistryFQDN)
	exists, err := a.Ou.Filesystem().Exists(ctx, registryRootDirectory)
	if err != nil {
		return err
	}
	if !exists {
		err = a.Ou.Filesystem().MkdirAll(ctx, registryRootDirectory, constants.DirPerm)
		if err != nil {
			return err
		}
	}
	if len(a.RegistryCapabilities) == 0 {
		a.RegistryCapabilities = append(a.RegistryCapabilities, "pull", "resolve")
	}
	tomlContent := make(map[string]interface{})
	registryInformation := make(map[string]interface{})
	tomlContent["server"] = a.RegistryEndpoint
	hostInformation := make(map[string]interface{})
	registryInformation["capabilities"] = a.RegistryCapabilities
	if a.SkipVerify {
		registryInformation["skip_verify"] = true
	} else {
		if len(a.CertFilesLocation) == 0 {
			// passing the default value
			a.CertFilesLocation = append(a.CertFilesLocation, constants.DefaultRegistryCaFilesLocation)
		}
		registryInformation["ca"] = a.CertFilesLocation
	}
	hostInformation[a.RegistryEndpoint] = registryInformation
	tomlContent["host"] = hostInformation
	tomlContentMap, err := toml.TreeFromMap(tomlContent)
	if err != nil {
		return err
	}
	tomlContentBytes, err := tomlContentMap.Marshal()
	if err != nil {
		return err
	}
	logger.Info("writing toml content", "Content", string(tomlContentBytes))
	err = a.Ou.Filesystem().WriteFile(ctx,
		filepath.Join(registryRootDirectory, "hosts.toml"),
		tomlContentBytes, constants.FilePerm)
	if err != nil {
		return err
	}
	file, err := a.Ou.Filesystem().ReadFile(ctx,
		filepath.Join(registryRootDirectory, "hosts.toml"))
	if err != nil {
		return err
	}
	logger.Info("hosts file content", "Content", string(file))
	return nil
}

// Del will Delete the hosts file and the directory name with FQDN
func (a PrivateRegistry) Del(ctx context.Context) error {
	return a.Ou.Filesystem().RemoveAll(ctx, filepath.Join(containerdRootDirectory, "certs.d", a.RegistryFQDN, "hosts.toml"))
}

// Get will return the content of the file
func (a PrivateRegistry) Get(ctx context.Context) ([]byte, error) {
	return a.Ou.Filesystem().ReadFile(ctx, filepath.Join(containerdRootDirectory, "certs.d", a.RegistryFQDN, "hosts.toml"))
}

// Update will update the content of the file
func (a PrivateRegistry) Update(ctx context.Context) error {
	err := a.Del(ctx)
	if err != nil {
		return err
	}
	return a.Add(ctx)
}
