package constants

import "time"

// Conditions Reason and Message for cluster provisioning and node
const (
	ControlPlaneStatusMessageSuccess = "Control Plane in ready state"
	ControlPlaneStatusMessageFailed  = "Control plane is not ready"

	ClusterReadyStatusMessageSuccess = "Cluster is in ready state"
	ClusterReadyStatusMessageFailed  = "Cluster is not in ready state"

	InstallReadyStatusMessageSuccess = "Install has succeeded"
	InstallReadyStatusMessageFailed  = "Install has not succeeded"

	ClusterDeleteMessageFailed = "Delete has failed"

	ControlUpgradeMessageFailed = "Upgrade has failed"

	NodeReadyStatusMessageSuccess = "Node is in ready state"
	NodeReadyStatusMessageFailed  = "Node is not in ready state"

	CniAddonStatusMessageSuccess = "Cni is in ready state"
	CniAddonStatusMessageFailed  = "Cni is not in ready state"
)

// Conditions and messages for Node customisation
const (
	PackageReadyStatusMessageSuccess = "Package installation success"
	PackageReadStatusMessageFailed   = "Package installation failed"
)

// Default
const (
	DefaultKubernetesTool = "kubeadm"
)

// Cluster Phase
const (
	ClusterPhaseProvisioning        = "Provisioning"
	ClusterPhaseDeleting            = "Deleting"
	ClusterPhaseKubeConfigResetting = "KubeConfigResetting"
	ClusterPhaseUpgrading           = "Upgrading"

	ClusterPhaseUnknown        = "Unknown"
	ClusterPhaseProvisioned    = "Provisioned"
	ClusterPhaseNotInitialised = "NotInitialised"
	ClusterPhaseFailed         = "Failed"
	ClusterPhaseDelete         = "Deleted"
)

// Package Phase
const (
	PackagePhaseInstall    = "Customized"
	PackagePhaseInstalling = "Customizing"
	PackagePhaseUnknown    = "Unknown"
)

// kernel

const (
	KubernetesKernelModuleFile = "/etc/modules-load.d/k8s.conf"
	KubernetesSysctlModuleFile = "/etc/sysctl.d/k8s.conf"
)

const (
	// ConditionSeverityError specifies that a condition with `Status=False` is an error.
	ConditionSeverityError = "Error"

	// ConditionSeverityWarning specifies that a condition with `Status=False` is a warning.
	ConditionSeverityWarning = "Warning"

	// ConditionSeverityInfo specifies that a condition with `Status=False` is informative.
	ConditionSeverityInfo = "Info"

	// ConditionSeverityNone should apply only to conditions with `Status=True`.
	ConditionSeverityNone = ""
)

const (
	KubeadmKubeconfigPath                       = "/etc/kubernetes/admin.conf"
	K3sKubeconfigPath                           = "/etc/rancher/k3s/k3s.yaml"
	KubeadmClusterSuccessfulInstallationMessage = "Your Kubernetes control-plane has initialized successfully!"
	KubeadmClusterSuccessfulUpgradeBanner       = "SUCCESS! Your cluster was upgraded to"
	AdminKubeconfigDirPath                      = "/home/admin/.kube"
	AdminKubeconfigPath                         = "/home/admin/.kube/config"
	KubeconfigFileName                          = "config"
	ClusterUpgradeWaitDuration                  = 10
	ClusterCertsRotationDays                    = 60
)

// Users
const (
	AdminUserName = "admin"
	RootUserName  = "root"
)

// Permissions
const (
	DirPerm               = 0755
	FilePerm              = 0644
	FileReadWriteAccess   = 0600
	OwnerReadWriteExecute = 0700
)

// Bits
const (
	Base    = 10
	BitSize = 64
)

// Resources
const (
	ResourceDirectory = "/opt/agent/kubeclusteragent/store"
	CertsDirectory    = "/opt/agent/kubeclusteragent/pki"
)

// Server Config
const (
	MetricsServerPort         = "31800"
	DefaultPortForGrpcServer  = "50050"
	DefaultPortForGrpcGateway = "8080"
)

// Creation Options
const (
	NodeReadinessMaxRetryCount = 30
	NodeReadinessRetryInterval = 10 * time.Second
)

// CRI
const (
	ConfigFileLocation             = "/etc/containerd/config.toml"
	ContainerdRootDirectory        = "/common/containerd/data"
	ContainerdStateDirectory       = "/common/containerd/run"
	DefaultRegistryCaFilesLocation = "/opt/config/airgap/ca.crt"
	ContainerdAddress              = "/run/containerd/containerd.sock"
	ContainerdKubernetesNamespace  = "k8s.io"
	K8sControlplaneImageLocation   = "/opt/images/cri_images"
)

// Networking
const (
	RootCNIPath           = "/opt/agent/cni"
	DefaultPodNetwork     = "100.100.0.0/16"
	DefaultServiceNetwork = "100.101.0.0/16"
	DefaultNetworkFolder  = "/etc/systemd/network"
	SecondaryIPfileName   = "sec_ip.conf"
)

// CSI

const (
	RootCSIPath = "/opt/agent/csi"
)

// k8s CP images
const (
	PauseImage                                = "/pause"
	CoreDNSImage                              = "/coredns"
	EtcdImage                                 = "/etcd"
	K8sControlPlaneImageVersionPropertiesFile = "/opt/config/k8s_versions.properties"
	CoreDNSPropFile                           = "K8S_RELEASE_COREDNS_VERSION"
	EtcdPropFile                              = "K8S_RELEASE_ETCD_VERSION"
	PausePropFile                             = "K8S_RELEASE_PAUSE_VERSION"
	K8sPropFile                               = "K8S_RELEASE_KUBERNETES_VERSION"
)

// Systemd Services

const (
	SystemdNetworkProcess = "systemd-networkd"
)

// IP

const (
	PrivateAddressPrefixLength   = "32"
	PrivateIPv4AddressMaskDigest = "100.102.1.1/32"
	PrivateIPv4Address           = "100.102.1.1"
	PrivateIPv6Address           = "2001:db8::1"
	PrivateIPv6AddressMaskDigest = "2001:db8::1/128"
	DefaultKubernetesBindPort    = "6443"
	IPCommand                    = "ip"
	IPConfigCommand              = "ifconfig"
	VirtualInterfaceType         = "dummy"
	VirtualInterfaceName         = "agent"
	PodIPv4Reservation           = "100.100.0.0/24"
	PodIPv6Reservation           = "2001:db8:1::/120"
	CNIManifestFilePath          = "/opt/cni/calico.yaml"
)

var (
	DefaultPrimaryNetworkInterface = "lo"
)

// CAP
const (
	CapSocketAddress      = "/tmp/capengine.sock"
	ApplianceProppFile    = "/common/configs/appliance.properties"
	StaticPodManifests    = "/etc/kubernetes/manifests"
	StaticPodManifestsBkp = "/etc/kubernetes/manifests-bkp"
)

// Authorized Keys

const (
	AuthorizedKeyCreated = "success"
	AuthorizedKeyFile    = "/home/admin/.ssh/authorized_keys"
	AuthorizedKeyDirPath = "/home/admin/.ssh"
)
