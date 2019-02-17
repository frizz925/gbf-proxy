#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Setup script need to be run as root."
    echo "Re-running setup script using sudo..."
    sudo /bin/bash "$0"
    exit $?
fi

CNI_VERSION="v0.7.4"
CRICTL_VERSION="v1.13.0"
K8S_BINARIES=(kubeadm kubelet kubectl)

echo "Getting Kubernetes stable release version..."
K8S_RELEASE=$(curl -sSL https://dl.k8s.io/release/stable.txt)
echo "Using Kubernetes version ${K8S_RELEASE}"

echo "Getting private IP..."
PRIVATE_IP=$(ip -f inet -o addr show eth1 | head -n 1 | awk '{ print $4 }' | cut -d/ -f1)
echo "Using private IP ${PRIVATE_IP}"

check_service_active() {
    systemctl is-active "$1" --quiet
    return $!
}

check_service_enabled() {
    systemctl list-unit-files | grep "$1" | grep -q enabled
    return $!
}

if ! check_service_active docker.service; then
    echo "Starting Docker service..."
    systemctl start docker.service
fi
if ! check_service_enabled docker.service; then
    echo "Enabling Docker service..."
    systemctl enable docker.service
fi

download_cni() {
    echo "Downloading CNI..."
    curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-amd64-${CNI_VERSION}.tgz" | tar -C /opt/cni/bin -xz
    echo "CNI downloaded."
}

if [ ! -d /opt/cni ]; then
    mkdir -p /opt/cni/bin
    download_cni
fi

download_cri() {
    echo "Downloading CRI..."
    curl -L "https://github.com/kubernetes-incubator/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz" | tar -C /opt/bin -xz
    echo "CRI downloaded."
}

if [ ! -d /opt/bin ]; then
    mkdir -p /opt/bin
    download_cri
fi

download_k8s_binary() {
    BINARY_NAME="$1"
    BINARY_PATH="$2"
    echo "Downloading Kubernetes binary: ${BINARY_NAME}..."
    curl -L "https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/amd64/${BINARY_NAME}" -o $BINARY_PATH
    chmod +x $BINARY_PATH
    echo "Kubernetes binary downloaded: ${BINARY_NAME}."
}

for kb in ${K8S_BINARIES[@]}; do
    BINARY_PATH="/opt/bin/$kb"
    if [ ! -f $BINARY_PATH ]; then
        download_k8s_binary "$kb" "$BINARY_PATH"
    fi
done

download_k8s_unit_files() {
    echo "Downloading Kubernetes systemd unit files..."
    curl -L "https://raw.githubusercontent.com/kubernetes/kubernetes/${K8S_RELEASE}/build/debs/kubelet.service" | sed "s:/usr/bin:/opt/bin:g" > /etc/systemd/system/kubelet.service
    curl -L "https://raw.githubusercontent.com/kubernetes/kubernetes/${K8S_RELEASE}/build/debs/10-kubeadm.conf" | sed "s:/usr/bin:/opt/bin:g" > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
    echo "Kubernetes systemd unit files downloaded."
}

if [ ! -f /etc/default/kubelet ]; then
    echo "Writing extra configurations for kubelet..."
    echo "KUBELET_EXTRA_ARGS=--node-ip=${PRIVATE_IP}" > /etc/default/kubelet
fi

if [ ! -d /etc/systemd/system/kubelet.service.d ]; then
    mkdir -p /etc/systemd/system/kubelet.service.d
    download_k8s_unit_files
fi

if ! check_service_active kubelet.service || ! check_service_enabled kubelet.service; then
    echo "Enabling and starting kubelet..."
    systemctl enable --now kubelet.service
fi

echo "Pulling Kubernetes images..."
kubeadm config images pull

echo "Initializing Kubernetes..."
kubeadm init --apiserver-advertise-address=$PRIVATE_IP --pod-network-cidr=10.244.0.0/16
