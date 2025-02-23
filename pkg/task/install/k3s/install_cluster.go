package k3s

import (
	"bytes"
	"context"
	"fmt"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/task"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"os"
	"text/template"
)

type Cluster struct{}

var _ task.Task = &Cluster{}

func NewInstallCluster() *Cluster {
	t := &Cluster{}

	return t
}

func (t *Cluster) Name() string {
	return "install-k3s-cluster"
}

func (t *Cluster) Run(
	ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	logger := log.From(ctx).WithValues(
		"Cluster Type", clusterSpec.ClusterType,
		"Version", clusterSpec.Version,
		"Operation", "Install", "Task", t.Name())
	logger.Info("Installing k3s cluster")
	configFilename := "/tmp/k3s-config.yaml"
	var output []byte
	contents, err := t.generateTemplate(clusterSpec)
	if err != nil {
		logger.Error(err, "Failed to generate kubeadm template")
		return err
	}
	if err := ou.Filesystem().WriteFile(ctx, configFilename, contents, 0600); err != nil {
		return fmt.Errorf("write k3s config file: %w", err)
	}
	cmdArgs := []string{"https://get.k3s.io", "-O", "/home/k3s.sh"}
	_, output, err = ou.Exec().Command(ctx, "wget", nil, cmdArgs...) // nolint
	if err != nil {
		return fmt.Errorf("download k3s installer failed: %w", err)
	}
	_, output, err = ou.Exec().Command(ctx, "chmod", nil, []string{"+x", "/home/k3s.sh"}...) // nolint
	if err != nil {
		return fmt.Errorf("enbaling permission to k3s installer failed: %w", err)
	}
	if clusterSpec.Networking.CniManifestURL != "" {
		_, output, err = ou.Exec().Command(ctx, "/bin/sh", append(os.Environ(),
			"INSTALL_K3S_VERSION="+clusterSpec.Version,
			"K3S_CONFIG_FILE="+configFilename),
			[]string{"/home/k3s.sh", "-s", "-", "--flannel-backend", "none", "-disable-agent"}...)
		if err != nil {
			return fmt.Errorf("k3s installation failed: %w", err)
		}
	} else {
		_, output, err = ou.Exec().Command(ctx, "/bin/sh", append(os.Environ(),
			"INSTALL_K3S_VERSION="+clusterSpec.Version,
			"K3S_CONFIG_FILE="+configFilename), []string{"/home/k3s.sh"}...)
		if err != nil {
			return fmt.Errorf("k3s installation failed: %w", err)
		}
	}
	err = ou.Systemd().Start(ctx, "k3s")
	if err != nil {
		return fmt.Errorf("k3s server start failed : %w", err)
	}
	logger.Info("k3s installation output ", "output", string(output))
	return nil
}

func (t *Cluster) Rollback(ctx context.Context,
	status cluster.Status,
	clusterSpec *v1alpha1.ClusterSpec,
	ou linux.OSUtil) error {
	return nil
}

type k3sTemplateData struct {
	ClusterCIDR string
	ServiceCIDR string
	CertsSANs   []string
}

func (t *Cluster) generateTemplate(clusterSpec *v1alpha1.ClusterSpec) ([]byte, error) {
	tmpl, err := template.New("install").Parse(k3sInstallTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse k3s configuration template: %w", err)
	}
	data := k3sTemplateData{
		ClusterCIDR: clusterSpec.Networking.PodSubnet,
		ServiceCIDR: clusterSpec.Networking.SvcSubnet,
		CertsSANs:   clusterSpec.ApiServer.CertSANs,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

var k3sInstallTemplate = `
cluster-cidr: "{{ .ClusterCIDR }}"
service-cidr: "{{ .ServiceCIDR }}"
tls-san:
{{ if .CertsSANs }}
{{ range .CertsSANs }}
    - "{{ . }}"
{{ end }}
{{ end }}
`
