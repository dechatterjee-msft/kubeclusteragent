package kubeadm

import (
	"bytes"
	"context"
	"fmt"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility"
	"runtime"
	"strings"
	"text/template"

	"kubeclusteragent/pkg/task"

	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
)

type Cluster struct{}

var _ task.Task = &Cluster{}

func NewInstallCluster() *Cluster {
	t := &Cluster{}

	return t
}

func (t *Cluster) Name() string {
	return "install-kubeadm-cluster"
}

var apiServerAddress = ""

func (t *Cluster) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	logger := log.From(ctx).WithName("task").WithName(t.Name())
	logger.Info("Install Kubernetes cluster", "version", clusterSpec.Version)
	ok, err := ou.Filesystem().Exists(ctx, "/etc/kubernetes/pki/ca.key")
	if err != nil {
		return fmt.Errorf("file exist error: %w", err)
	}
	if ok {
		return nil
	}
	configFilename := "/tmp/kubeadm-config.yaml"
	// apiServerAddress = getSecondaryNetworkForK8s(ctx, ou)
	apiServerAddress = constants.PrivateIPv4Address
	contents, err := t.generateTemplate(ctx, clusterSpec)
	if err != nil {
		logger.Error(err, "Failed to generate kubeadm template")
		return err
	}
	if err := ou.Filesystem().WriteFile(ctx, configFilename, contents, 0600); err != nil {
		return fmt.Errorf("write kubeadm config file: %w", err)
	}
	var cmdArgs []string
	if runtime.NumCPU() < 2 {
		cmdArgs = append(cmdArgs, "init", "--config", configFilename, "--ignore-preflight-errors=NumCPU")
		logger.Info("this machine has 1 CPU , still we are progressing ")
	} else {
		cmdArgs = append(cmdArgs, "init", "--config", configFilename)
	}
	_, output, err := ou.Exec().Command(ctx, "kubeadm", nil, cmdArgs...)
	if err != nil {
		return fmt.Errorf("run kubeadm: %w", err)
	}
	if strings.Contains(string(output), constants.KubeadmClusterSuccessfulInstallationMessage) {
		logger.Info("Kubeadm init output", "KubeadmOutput", constants.KubeadmClusterSuccessfulInstallationMessage)
	} else {
		err = fmt.Errorf("cluster installation failed %v", string(output))
		logger.Error(err, "cluster installation failed")
		return err
	}
	return nil
}

func (t *Cluster) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou osutility.OSUtil) error {
	return nil
}

type kubeadmTemplateData struct {
	KubernetesVersion string
	ClusterName       string
	PodSubnet         string
	ServiceSubnet     string
	CertSANs          []string
	EtcdDataDir       string
	EtcdImageRepo     string
	EtcdImageTag      string
	DNSImageRepo      string
	DNSImageTag       string
	PauseImageTag     string
	APIServerAddress  string
	BindPort          string
	NodeIP            string
}

func (t *Cluster) generateTemplate(ctx context.Context, clusterSpec *v1alpha1.ClusterSpec) ([]byte, error) {
	logger := log.From(ctx).WithName("generate-kubeadm-config")
	tmpl, err := template.New("install").Parse(kubeadmInstallTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse kubeadm configuration template: %w", err)
	}
	//cpImageTags, err := cri.GetK8sControlPlaneImagesFromPropertiesFile()
	//if err != nil {
	//	return nil, fmt.Errorf("fetching controlplane images tags error using crictl command: %w", err)
	//}
	logger.Info("kubernetes will be installed kubernetes version", "Version", clusterSpec.Version)
	// logger.Info("kubernetes will be installed with coredns and etcd", "Version", cpImageTags)
	data := kubeadmTemplateData{
		KubernetesVersion: clusterSpec.Version,
		PodSubnet:         clusterSpec.Networking.PodSubnet,
		ServiceSubnet:     clusterSpec.Networking.SvcSubnet,
		ClusterName:       clusterSpec.ClusterName,
		//DNSImageTag:       cpImageTags[constants.CoreDNSImage],
		//EtcdImageTag:      cpImageTags[constants.EtcdImage],
		//PauseImageTag:     cpImageTags[constants.PauseImage],
		BindPort: constants.DefaultKubernetesBindPort,
		NodeIP:   fmt.Sprintf("%s,%s", constants.PrivateIPv4Address, constants.PrivateIPv6Address),
	}
	if len(clusterSpec.ApiServer.CertSANs) > 0 {
		data.CertSANs = clusterSpec.ApiServer.CertSANs
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

const kubeadmInstallTemplate = `
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
nodeRegistration:
  criSocket: "/run/containerd/containerd.sock"
---
kind: ClusterConfiguration
apiVersion: kubeadm.k8s.io/v1beta3
kubernetesVersion: {{ .KubernetesVersion }}
kubeletConfiguration:
  cgroupDriver: systemd
clusterName: {{ .ClusterName }}
{{ if .APIServerAddress }}
controlPlaneEndpoint: "{{ .APIServerAddress }}:{{ .BindPort }}"
{{ end }}
controllerManager:
  extraArgs:
    leader-elect: "false"
    profiling: "false"
scheduler:
  extraArgs:
    leader-elect: "false"
    profiling: "false"
etcd:
  local:
    dataDir: /var/lib/etcd
    extraArgs:
      cipher-suites: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
networking:
  podSubnet: "{{ .PodSubnet }}"
  serviceSubnet: "{{ .ServiceSubnet }}"
apiServer:
  extraArgs:
    profiling: "false"
{{ if .CertSANs }}
  certSANs:
{{ range .CertSANs }}
    - {{ . }}
{{ end }}
{{ end }}

---
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
cgroupDriver: systemd
imageGCHighThresholdPercent: 100
imageGCLowThresholdPercent: 99
evictionHard:
  "imagefs.available": "0%"`

/*func getSecondaryNetworkForK8s(ctx context.Context, ou osutil.OSUtil) string {
	logger := log.From(ctx)
	secondaryNetwork := networkutil.NewNetworkInterface(constants.VirtualInterfaceName, constants.VirtualInterfaceType,
		constants.PrivateIPv4AddressMaskDigest, constants.PrivateIPv6AddressMaskDigest, ou)
	results, err := secondaryNetwork.Get(ctx)
	if err != nil {
		logger.Error(err, "error occurred while getting the network interfaces,kubernetes cluster will be created on primary network")
		return ""
	}
	logger.Info("network interfaces available during cluster deployment", "Interface", results)
	if results[constants.PrivateIPv4AddressMaskDigest] {
		logger.Info("cluster will be bootstrapped", "IP", constants.PrivateIPv4AddressMaskDigest, constants.PrivateIPv6AddressMaskDigest)
		return constants.PrivateIPv4Address
	}
	return ""
}
*/
