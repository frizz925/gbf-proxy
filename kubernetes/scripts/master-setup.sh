#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Master node setup script need to be run as root."
    echo "Re-running master node setup script using sudo..."
    sudo -HE /bin/bash "$0" $@
    exit $?
fi

if [ -z "$PUBLIC_IP" ]; then
    if [ -n "$1" ]; then
        PUBLIC_IP="$1"
    elif [ -n "$COREOS_PUBLIC_IPV4" ]; then
        PUBLIC_IP="$COREOS_PUBLIC_IPV4"
    elif [ -n "$PUBLIC_IFACE" ]; then
        echo "Public IP address argument not provided. Using provided network interface IP address..."
        PUBLIC_IP=$(ip -f inet -o addr show $PUBLIC_IFACE | awk '{ print $4 }' | cut -d/ -f1 | head -n1)
    else
        echo "IP address argument not provided. Using default IP address..."
        GATEWAY_IP=$(ip route | awk '/^default/{ print $3 }')
        PUBLIC_IP=$(ip route get $GATEWAY_IP | awk 'NR==1{ print $5 }')
    fi
fi

if [ -z "$PRIVATE_IP" ]; then
    if [ -n "$2" ]; then
        PRIVATE_IP="$2"
    elif [ -n "$COREOS_PRIVATE_IPV4" ]; then
        PRIVATE_IP="$COREOS_PRIVATE_IPV4"
    elif [ -n "$PRIVATE_IFACE" ]; then
        echo "Private IP address argument not provided. Using provided network interface IP address..."
        PRIVATE_IP=$(ip -f inet -o addr show $PRIVATE_IFACE | awk '{ print $4 }' | cut -d/ -f1 | head -n1)
    fi
fi

if [ -z "$KUBEADM_APISERVER_ADDRESS" ]; then
    if [ -n "$PRIVATE_IP" ]; then
        KUBEADM_APISERVER_ADDRESS="$PRIVATE_IP"
    else
        KUBEADM_APISERVER_ADDRESS="$PUBLIC_IP"
    fi
fi

EXTRA_SANS="$3"
if [ -z "$EXTRA_SANS" ]; then
    EXTRA_SANS="$PUBLIC_IP"
fi

if [ -z "$KUBEADM_EXTRA_SANS" ]; then
    KUBEADM_EXTRA_SANS="$EXTRA_SANS"
else
    KUBEADM_EXTRA_SANS="$KUBEADM_EXTRA_SANS,$EXTRA_SANS"
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
echo "kubeadm extra SANs: ${KUBEADM_EXTRA_SANS}"
if [ -n "$KUBEADM_EXTRA_ARGS" ]; then
    echo "kubeadm extra args: ${KUBEADM_EXTRA_ARGS}"
fi
kubeadm init --apiserver-advertise-address="$KUBEADM_APISERVER_ADDRESS" \
    --apiserver-cert-extra-sans="$KUBEADM_EXTRA_SANS" \
    $KUBEADM_EXTRA_ARGS
echo "Kubernetes initialized."
