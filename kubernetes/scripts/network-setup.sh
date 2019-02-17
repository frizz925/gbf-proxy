#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Network setup script need to be run as root."
    echo "Re-running network setup script using sudo..."
    sudo /bin/bash "$0"
    exit $!
fi

FLANNEL_VERSION="v0.11.0"
export KUBECONFIG=/etc/kubernetes/admin.conf

echo "Enabling bridged traffic to iptables' chains..."
sysctl net.bridge.bridge-nf-call-iptables=1

echo "Installing Flannel add-on..."
kubectl apply -f "https://raw.githubusercontent.com/coreos/flannel/${FLANNEL_VERSION}/Documentation/kube-flannel.yml"
