#!/bin/bash
set -e

NETWORKING_ADDON="$1"
if [ -z "$NETWORKING_ADDON" ]; then
    echo "Networking add-on not specified. Assuming Weave add-on."
    NETWORKING_ADDON="weave"
fi

FLANNEL_VERSION="v0.11.0"
CALICO_VERSION="v3.5"
CILIUM_VERSION="v1.4.0"

K8S_VERSION="1.13"

if [ "$NETWORKING_ADDON" = "flannel" ]; then
    echo "Removing Flannel add-on..."
    kubectl delete -f "https://raw.githubusercontent.com/coreos/flannel/${FLANNEL_VERSION}/Documentation/kube-flannel.yml" || true
    echo "Flannel add-on removed."
elif [ "$NETWORKING_ADDON" = "calico" ]; then
    echo "Removing Calico add-on..."
    kubectl delete -f "https://docs.projectcalico.org/${CALICO_VERSION}/getting-started/kubernetes/installation/hosted/etcd.yaml" || true
    kubectl delete -f "https://docs.projectcalico.org/${CALICO_VERSION}/getting-started/kubernetes/installation/hosted/calico.yaml" || true
    echo "Calico add-on removed."
elif [ "$NETWORKING_ADDON" = "cilium" ]; then
    echo "Removing Cilium add-on..."
    kubectl delete -f "https://raw.githubusercontent.com/cilium/cilium/${CILIUM_VERSION}/examples/kubernetes/${K8S_VERSION}/cilium.yaml" || true
    echo "Cilium add-on removed."
elif [ "$NETWORKING_ADDON" = "weave" ]; then
    echo "Removing Weave add-on..."
    kubectl delete -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')" || true
    echo "Weave add-on removed."
else
    echo "Unknown networking addon '$NETWORKING_ADDON'" >&2
    exit 1
fi
