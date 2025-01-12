#!/usr/bin/env bash


get_version(){

    component_type="$1"
    component_name="$2"
    version=$(kubectl get pod -n kube-system -l "$component_type"="$component_name" -o=jsonpath='{.items[0].spec.containers[0].image}{"\n"}')
    # version_print=$(echo "$version" | grep -o 'v[0-9]\+\.[0-9]\+')
    version_print=$(echo "$version" | awk -F':' '{print $2}')
    # echo "$version_print"
    if [ "$component_name" == "kube-dns" ]; then
        component_name="core_dns"
    fi
    echo "$component_name version: $version_print"
}

echo
echo "--------- CURRENT VERSIONS -------------"
echo
get_version "component" "kube-apiserver"
get_version "component" "kube-controller-manager"
get_version "k8s-app" "kube-proxy"
get_version "component" "kube-scheduler"
get_version "k8s-app" "kube-dns"
echo
echo "-------------------------------------------"
echo