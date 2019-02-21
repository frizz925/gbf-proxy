#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Master node setup script need to be run as root."
    echo "Re-running master node setup script using sudo..."
    sudo -HE /bin/bash "$0" $@
    exit $?
fi

if [ -z "$KUBEADM_APISERVER_ADDRESS" ]; then
    if [ -n "$1" ]; then
        KUBEADM_APISERVER_ADDRESS="$1"
    elif [ -n "$LOCAL_IFACE" ]; then
        echo "IP address argument not provided. Using provided network interface IP address..."
        KUBEADM_APISERVER_ADDRESS=$(ip -f inet -o addr show $LOCAL_IFACE | awk '{ print $4 }' | cut -d/ -f1 | head -n1)
    else
        echo "IP address argument not provided. Using default IP address..."
        GATEWAY_IP=$(ip route | awk '/^default/{ print $3 }')
        KUBEADM_APISERVER_ADDRESS=$(ip route get $GATEWAY_IP | awk 'NR==1{ print $5 }')
    fi
fi

if ! systemctl is-active docker.service --quiet; then
    echo "Starting Docker service..."
    systemctl start docker.service
    echo "Docker service started."
fi

echo "Pulling Kubernetes images..."
kubeadm config images pull

echo "Using IP address ${IP_ADDRESS}"
echo "Initializing Kubernetes..."
if [ -n "$KUBEADM_EXTRA_ARGS" ]; then
    echo "kubeadm extra args: ${KUBEADM_EXTRA_ARGS}"
fi
kubeadm init --apiserver-advertise-address=$KUBEADM_APISERVER_ADDRESS $KUBEADM_EXTRA_ARGS
echo "Kubernetes initialized."
