#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Network setup script need to be run as root."
    echo "Re-running network setup script using sudo..."
    sudo /bin/bash "$0" $@
    exit $?
fi

NETWORKING_ADDON="$1"
if [ -z "$NETWORKING_ADDON" ]; then
    echo "Networking add-on not specified. Assuming Weave add-on."
    NETWORKING_ADDON="weave"
fi

FLANNEL_VERSION="v0.11.0"
CALICO_VERSION="v3.5"
CILIUM_VERSION="v1.4.0"

K8S_VERSION="1.13"
export KUBECONFIG=/etc/kubernetes/admin.conf

echo "Enabling bridged traffic to iptables' chains..."
sysctl net.bridge.bridge-nf-call-iptables=1

if [ "$NETWORKING_ADDON" = "flannel" ]; then
    echo "Installing Flannel add-on..."
    kubectl apply -f "https://raw.githubusercontent.com/coreos/flannel/${FLANNEL_VERSION}/Documentation/kube-flannel.yml"
    echo "Flannel add-on installed."
elif [ "$NETWORKING_ADDON" = "calico" ]; then
    echo "Installing Calico add-on..."
    kubectl apply -f "https://docs.projectcalico.org/${CALICO_VERSION}/getting-started/kubernetes/installation/hosted/etcd.yaml"
    kubectl apply -f "https://docs.projectcalico.org/${CALICO_VERSION}/getting-started/kubernetes/installation/hosted/calico.yaml"
    echo "Calico add-on installed."
elif [ "$NETWORKING_ADDON" = "cilium" ]; then
    echo "Installing Cilium add-on..."
    kubectl apply -f "https://raw.githubusercontent.com/cilium/cilium/${CILIUM_VERSION}/examples/kubernetes/${K8S_VERSION}/cilium.yaml"
    echo "Cilium add-on installed."
elif [ "$NETWORKING_ADDON" = "weave" ]; then
    echo "Installing Weave add-on..."
    kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')"
    echo "Weave add-on installed."
else
    echo "Unknown networking addon '$NETWORKING_ADDON'" >&2
    exit 1
fi
