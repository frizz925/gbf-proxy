#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Teardown script need to be run as root."
    echo "Re-running teardown script using sudo..."
    sudo -HE /bin/bash "$0"
    exit $?
fi

if [ -f /opt/bin/kubeadm ]; then
    echo "Resetting Kubernetes..."
    /opt/bin/kubeadm reset -f
    echo "Kubernetes reset."
fi

if [ -f /etc/systemd/system/kubelet.service ]; then
    echo "Stopping kubelet..."
    systemctl stop kubelet
    echo "Kubelet stopped."
fi

DOCKER_CONTAINERS=$(docker ps -aq || printf '')
if [ -n "$DOCKER_CONTAINERS" ]; then
    echo "Removing docker containers..."
    docker rm -f $DOCKER_CONTAINERS
    echo "Docker containers removed."
fi

echo "Stopping docker..."
systemctl stop docker
echo "Docker stopped."

for d in /var/lib/cni /var/lib/kubelet /etc/cni; do
    if [ -d $d ]; then
        echo "Removing directory: $d..."
        rm -rf $d
        echo "Directory removed: $d."
    fi
done

for i in cni0 flannel.1 docker0 weave cilium_host cilium_net cilium_vxlan; do
    if [ -e /sys/class/net/$i ]; then
        echo "Removing network interface: $i..."
        ifconfig $i down
        echo "Network interface removed: $i."
    fi
done

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
