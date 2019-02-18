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
    echo "Kubernetes reset."
fi

echo "Stopping kubelet..."
systemctl stop kubelet
echo "Kubelet stopped."

DOCKER_CONTAINERS=$(docker ps -aq)
if [ -n "$DOCKER_CONTAINERS" ]; then
    echo "Removing docker containers..."
    docker rm -f $DOCKER_CONTAINERS
    echo "Docker containers removed."
fi

echo "Stopping docker..."
systemctl stop docker
echo "Docker stopped."

echo "Removing files..."
[ -d /var/lib/cni ] && rm -rf /var/lib/cni/
[ -d /var/lib/kubelet ] && rm -rf /var/lib/kubelet/*
[ -d /etc/cni ] && rm -rf /etc/cni/
echo "Files removed."

# echo "Removing network interfaces..."
# ifconfig cni0 down
# ifconfig flannel.1 down
# ifconfig docker0 down
# echo "Network interfaces removed."

echo "Clearing iptables' rules..."
iptables -F
iptables -t nat -F
iptables -t mangle -F
iptables -X
echo "iptables' rules cleared."

if [ -n "$(command -v ipvsadm)" ]; then
    echo "Resetting IPVS tables..."
    ipvsadm --clear
    echo "IPVS tables reset."
fi
