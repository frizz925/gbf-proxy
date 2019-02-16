#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Teardown script need to be run as root."
    echo "Re-running teardown script using sudo..."
    sudo /bin/bash "$0"
    exit $?
fi

if [ -f /opt/bin/kubeadm ]; then
    echo "Resetting Kubernetes..."
    /opt/bin/kubeadm reset -f
fi

echo "Clearing iptable rules..."
iptables -F
iptables -t nat -F
iptables -t mangle -F
iptables -X
