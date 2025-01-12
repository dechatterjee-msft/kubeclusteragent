#Steps
#!/usr/bin/env bash

prepare_vm(){

    #Download kubeadm kubectl and kubelet "rpm" files
    echo "downloading rpm files"
    curl -O kubeadm-1.27.2-1.el7.x86_64.rpm
    curl -O kubectl-1.27.2-1.el7.x86_64.rpm
    curl -O kubelet-1.27.2-1.el7.x86_64.rpm

    #Run the rpm files using rpm -Uvh
    rpm -Uvh kubeadm-1.27.2-1.el7.x86_64.rpm
    rpm -Uvh kubectl-1.27.2-1.el7.x86_64.rpm
    systemctl stop kubelet
    rpm -Uvh kubelet-1.27.2-1.el7.x86_64.rpm
    systemctl start kubelet
    systemctl daemon-reload
    kubelet --version

    #create a folder with the latest images 
    echo "creating folder for containerd images..."
    mkdir cri_images
    cd cri_images/

    curl -O kube-apiserver-v1.27.2.tar.gz
    curl -O kubernetes-v1.27.2/kubernetes/images/kube-controller-manager-v1.27.2.tar.gz
    curl -O kubernetes-v1.27.2/kubernetes/images/kube-proxy-v1.27.2.tar.gz
    curl -O kubernetes-v1.27.2/kubernetes/images/kube-scheduler-v1.27.2.tar.gz
    curl -O kubernetes-v1.27.2/kubernetes/images/pause-3.9.tar.gz

    cd ..
    mv cri_images /opt/images/cri_images
}

prepare_vm