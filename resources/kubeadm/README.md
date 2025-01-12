## Install  Kubernetes 

### Resource Prerequisites

    Linux Host.
    RAM 2GB.
    CPU 2 .
    Unique hostname, MAC address, and product_uuid for the node.


### Download location
    

### Setting Unique Hostname 

```
### Disable swap 

```shell
  sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
```

### Provisioning IP tables

```shell

   # etcd
      iptables -A INPUT -p tcp -m tcp --dport 2379:2380 -j ACCEPT
   # api-server
      iptables -A INPUT -p tcp -m tcp --dport 6443 -j ACCEPT
   # node-port
      iptables -A INPUT -p tcp -m tcp --dport 10250:10252 -j ACCEPT
   # calico optional
      iptables -A INPUT -p tcp -m tcp --dport 179 -j ACCEPT
      iptables -A INPUT -p tcp -m tcp --dport 4789 -j ACCEPT
   # save
      iptables-save > /etc/systemd/scripts/ip4save 
      
```

### Containerd Installation

```shell
    tar -zxvf cri-containerd-v1.6.18+.linux-amd64.tar.gz
    mv usr/local/bin/containerd usr/local/bin/containerd-shim usr/local/bin/containerd-shim-runc-v1 usr/local/bin/containerd-shim-runc-v2 usr/local/bin/containerd-stress usr/local/bin/crictl usr/local/bin/critest usr/local/bin/ctr  /usr/bin/
    mv usr/local/sbin/runc /usr/bin/
    sed -i 's/ExecStart=.*/ExecStart=\/usr\/bin\/containerd/' etc/systemd/system/containerd.service
    mv etc/systemd/system/containerd.service  /usr/lib/systemd/system/containerd.service
    systemctl enable containerd
    systemctl start containerd
    cat > /etc/modules-load.d/containerd.conf <<EOF
    br_netfilter
    EOF
    modprobe br_netfilter
    cat > /etc/sysctl.d/99-kubernetes-cri.conf <<EOF
        net.bridge.bridge-nf-call-iptables = 1
        net.ipv4.ip_forward = 1
        net.bridge.bridge-nf-call-ip6tables = 1
    EOF
    sysctl --system
    containerd config default > /etc/containerd/config.toml
    systemctl restart containerd 
```

### Kubernetes RPM

     kubeadm-1.24.11-1.el7.x86_64.rpm
     kubectl-1.24.11-1.el7.x86_64.rpm
     kubelet-1.24.11-1.el7.x86_64.rpm
     kubernetes-cni-1.1.1-1.el7.x86_64.rpm
     cri-tools-1.24.2-1.el7.x86_64.rpm

#### Installation

Dependencies

    conntrack 
    socat
    ebtables
    ethtool

```shell
   rpm -i *.rpm 
```
### Controlplane Images

      coredns-v1.9.3.tar
      etcd-v3.5.6.tar
      kube-apiserver-v1.24.11.tar
      kube-controller-manager-v1.24.11.tar
      kube-proxy-v1.24.11.tar
      kube-scheduler-v1.24.11.tar
      pause-3.8.tar

#### Installation

Upon downloading images from the build-web these images will be in tar.gz format,they need to be decompressed to tar
as containerd will not accept tar.gz format

```shell
    gzip --decomrpess *.tar.gz
    ctr -n=k8s.io images import coredns-v1.9.3.tar
    ctr -n=k8s.io images import etcd-v3.5.6.tar
    ctr -n=k8s.io images import kube-apiserver-v1.24.11.tar
    ctr -n=k8s.io images import kube-controller-manager-v1.24.11.tar
    ctr -n=k8s.io images import kube-proxy-v1.24.11.tar
    ctr -n=k8s.io images import kube-scheduler-v1.24.11.tar
    ctr -n=k8s.io images import pause-3.8.tar
```


