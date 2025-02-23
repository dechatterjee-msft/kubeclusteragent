package cri

import (
	"context"
	"fmt"
	"kubeclusteragent/pkg/util/osutility/linux"
	"os"
	"strings"

	"github.com/go-logr/logr"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/magiconair/properties"
	"kubeclusteragent/pkg/constants"
)

var propFileLocation = constants.K8sControlPlaneImageVersionPropertiesFile
var imagePath = constants.K8sControlplaneImageLocation

type LiveContainerdClient struct {
	Address    string
	Namespace  string
	Connection *containerd.Client
}

func NewConnection(address, namespace string) (*LiveContainerdClient, error) {
	client, err := containerd.New(address)
	if err != nil {
		return nil, err
	}
	return &LiveContainerdClient{
		Address:    address,
		Namespace:  namespace,
		Connection: client,
	}, nil
}

type Client interface {
	ListImages(ctx context.Context, imageTag string) ([]string, error)
	DeleteImage(ctx context.Context, image string) error
	Close(ctx context.Context) error
	ListK8sControlplaneImages(ctx context.Context, imageTag string) ([]string, error)
}

func (c LiveContainerdClient) ListImages(ctx context.Context, imageTag string) ([]string, error) {
	containerdContext := namespaces.WithNamespace(ctx, c.Namespace)
	listOfImages := make([]string, 0)
	list, err := c.Connection.ImageService().List(containerdContext)
	if err != nil {
		return []string{}, err
	}
	for _, i := range list {
		if strings.Contains(i.Name, imageTag) {
			listOfImages = append(listOfImages, i.Name)
		}
	}
	return listOfImages, nil
}

// ListK8sControlplaneImages this is only applicable for 1.26.5 -> 1.27.2
func (c LiveContainerdClient) ListK8sControlplaneImages(ctx context.Context, imageTag string) ([]string, error) {
	images, err := c.ListImages(ctx, imageTag)
	if err != nil {
		return nil, err
	}
	results := make([]string, 0)
	for _, i := range images {
		if strings.Contains(i, "kube-scheduler") ||
			strings.Contains(i, "kube-controller-manager") ||
			strings.Contains(i, "kube-apiserver") ||
			strings.Contains(i, "kube-proxy") {
			results = append(results, i)
		}
	}
	return results, nil
}

func (c LiveContainerdClient) DeleteImage(ctx context.Context, image string) error {
	containerdContext := namespaces.WithNamespace(ctx, c.Namespace)
	return c.Connection.ImageService().Delete(containerdContext, image)
}

func (c LiveContainerdClient) Close(ctx context.Context) error {
	return c.Connection.Close()
}

func GetK8sControlPlaneImagesFromPropertiesFile() (map[string]string, error) {
	file, err := properties.LoadFile(propFileLocation, properties.UTF8)
	if err != nil {
		return nil, err
	}
	results := make(map[string]string)
	coreDNS, ok := file.Get(constants.CoreDNSPropFile)
	if !ok {
		return nil, fmt.Errorf("unable to find  coredns  version i.e. %s in the properties files %s", constants.CoreDNSPropFile,
			constants.K8sControlPlaneImageVersionPropertiesFile)
	}
	results[constants.CoreDNSImage] = "v" + strings.ReplaceAll(coreDNS, "+", "_")
	etcdVersion, ok := file.Get(constants.EtcdPropFile)
	if !ok {
		return nil, fmt.Errorf("unable to find  etcd  version i.e. %s in the properties files %s", constants.EtcdPropFile,
			constants.K8sControlPlaneImageVersionPropertiesFile)
	}
	results[constants.EtcdImage] = "v" + strings.ReplaceAll(etcdVersion, "+", "_")
	pause, ok := file.Get(constants.PausePropFile)
	if !ok {
		return nil, fmt.Errorf("unable to find  pause  version i.e. %s in the properties files %s", constants.PausePropFile,
			constants.K8sControlPlaneImageVersionPropertiesFile)
	}
	results[constants.PauseImage] = strings.ReplaceAll(pause, "+", "_")
	return results, nil
}

func GetImageVersionForCleanup() string {
	// TODO get all the image's that need to be purged
	// need to get all the controlplane images this will be implemented in the next release where we have to
	// cleanup CNI,CSI and Core-dns as well
	return "v1.26.5"
}

func LoadContainerdImages(ctx context.Context, ou linux.OSUtil, logger logr.Logger) error {
	err := os.Chdir(imagePath)
	if err != nil {
		return err
	}
	logger.Info("deleting current .tar files")
	_, _, err = ou.Exec().Command(ctx, "sh", nil, []string{"-c", "rm -rf *.tar"}...)
	if err != nil {
		return err
	}
	logger.Info("decompressing image tar.gz files")
	_, status, err := ou.Exec().Command(ctx, "sh", nil, []string{"-c", "gzip --decompress *.tar.gz"}...)
	if err != nil {
		return err
	}
	logger.Info("command to decompress tar.gz file ran successfully", "status", status)
	files, err := os.ReadDir(imagePath)
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar") {
			logger.Info("importing ", file.Name(), "to containerd")
			_, data, err := ou.Exec().Command(ctx, "ctr", nil, []string{"-n=k8s.io", "images", "import", file.Name()}...)
			if err != nil {
				return err
			}
			logger.Info("successfully imported", file.Name(), "to containerd", "status", data)
		}
	}
	if err != nil {
		return err
	}
	return nil
}
