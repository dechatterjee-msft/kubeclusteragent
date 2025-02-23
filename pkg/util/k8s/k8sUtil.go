package k8s

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
)

type K8sUtil struct{}

var retryCount int
var sleep = 100 * time.Second

var kubectlClient linux.Kubectl = linux.NewLiveKubectl(linux.NewLiveExec())
var hostUtil linux.Host = &linux.LiveHost{}

func (k8s *K8sUtil) NodeWorkloadScheduler(ctx context.Context, operationName string) error {
	logger := log.From(ctx)
	logger.Info("Running kubectl command ", operationName)
	nodeName, err := hostUtil.GetHostname()
	if err != nil {
		return fmt.Errorf("hostname :  %w", err)
	}
	_, err = kubectlClient.RunWithResponse(ctx, []string{operationName, nodeName}...)
	if err != nil {
		return fmt.Errorf("kubectl run  : %w", err)
	}
	// Verifying if node is in Ready state
	if operationName == "uncordon" {
		data, err := kubectlClient.RunWithResponse(ctx, []string{"get", "nodes"}...)
		if err != nil {
			return fmt.Errorf("kubectl run  : %w", err)
		}
		if strings.Contains(data, "SchedulingDisabled") {
			// Retrying to enable scheduling in the node
			logger.Info("retrying to enable scheduling ", operationName)
			time.Sleep(sleep)
			retryCount++
			if retryCount < 3 {
				err = k8s.NodeWorkloadScheduler(ctx, operationName)
				if err != nil {
					return fmt.Errorf("kubectl run  : %w", err)
				}
			}
		}
	}
	return nil
}

func GetKubeClientFromKubeconfig(kubeConfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func GetNode(ctx context.Context, clientset *kubernetes.Clientset) (*v1.Node, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(nodes.Items) == 1 {
		node := nodes.Items[0]
		return &node, nil
	}
	return nil, nil
}

func GetKubeSystemPodStatus(ctx context.Context, kubeSystem []string, clientset *kubernetes.Clientset) (bool, error) {
	pods, err := clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, pod := range pods.Items {
		if isKubeSystemPod(kubeSystem, pod.Name) && pod.Status.Phase != "Running" {
			return false, nil
		}
	}
	return true, nil
}

func isKubeSystemPod(kubeSystem []string, podName string) bool {
	for _, s := range kubeSystem {
		if strings.HasPrefix(podName, s) {
			return true
		}
	}
	return false
}

func CopyConfigMap(clientset *kubernetes.Clientset, ctx context.Context, namespace, name string) (*v1.ConfigMap, error, bool) {
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil, nil, false
	}
	if err != nil {
		return nil, err, false
	}
	cmCopy := cm.DeepCopy()
	cmCopy.ObjectMeta.CreationTimestamp = metav1.Time{}
	cmCopy.ObjectMeta.ResourceVersion = ""
	cmCopy.ObjectMeta.UID = ""
	return cmCopy, nil, false
}

func DeleteConfigMap(clientset *kubernetes.Clientset, ctx context.Context, namespace, name string) error {
	err := clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}

func CreateConfigMap(clientset *kubernetes.Clientset, ctx context.Context, namespace string, configMap *v1.ConfigMap) error {
	_, err := clientset.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
	return err
}

func UpdateK8sSecret(ctx context.Context, clientset *kubernetes.Clientset, name, namespace, data, key string) (bool, error) {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		return true, err
	}
	secretData := map[string][]byte{
		key: []byte(data),
	}
	secret.Data = secretData
	_, err = clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

func UpdateConfigMap(clientset *kubernetes.Clientset, ctx context.Context, namespace string, configMap *v1.ConfigMap) error {
	_, err := clientset.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	return err
}

func GetConfigMap(clientset *kubernetes.Clientset, ctx context.Context, key, name, namespace string) (*v1.ConfigMap, error) {
	return clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}
