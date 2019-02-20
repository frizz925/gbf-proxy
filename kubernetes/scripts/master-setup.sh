#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Master node setup script need to be run as root."
    echo "Re-running master node setup script using sudo..."
    sudo -HE /bin/bash "$0" $@
    exit $?
fi

IP_ADDRESS="$1"
if [ -z "$IP_ADDRESS" ]; then
    echo "IP address argument not provided. Using default IP address..."
    GATEWAY_IP=$(ip route | awk '/^default/{ print $3 }')
    IP_ADDRESS=$(ip route get $GATEWAY_IP | awk 'NR==1{ print $5 }')
fi

if ! systemctl is-active docker.service --quiet; then
    echo "Starting Docker service..."
    systemctl start docker.service
    echo "Docker service started."
fi

echo "Using IP address ${IP_ADDRESS}"
echo "Initializing Kubernetes..."
if [ -n "$KUBEADM_EXTRA_ARGS" ]; then
    echo "kubeadm extra args: ${KUBEADM_EXTRA_ARGS}"
fi
kubeadm init --apiserver-advertise-address=$IP_ADDRESS $KUBEADM_EXTRA_ARGS
echo "Kubernetes initialized."
