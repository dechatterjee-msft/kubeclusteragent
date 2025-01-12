
#!/usr/bin/env bash
upgrade_K8s() {
     # ---------FILE SPLIT --------- #

    #call function to load containerd

    echo "loading containerd..."
    load_containerd_images

    echo "containerd images uploaded successfully..."

    echo "copy coredns yaml.."
    kubectl get cm -n kube-system coredns -o yaml> coredns_cm.yaml

    echo "deleting coredns"
    kubectl delete cm coredns -n kube-system

    echo "upgrade starting"
    kubeadm upgrade apply v1.27.2 -f --ignore-preflight-errors=all

    echo "upgrade done"
    kubectl apply -f coredns_cm.yaml 

    #To check upgrade is successful
}

load_containerd_images(){
   distroRoot="/opt"
   imagesRoot="$distroRoot/images"
   cd   $imagesRoot/cri_images_update
   gzip --decompress *.tar.gz
   for f in *
   do
        ctr -n=k8s.io images import $f
        echo "successfully imported $f to containerd"
   done
}

upgrade_K8s