#!/bin/bash
set -e

if [ $EUID -ne 0 ]; then
    echo "Setup script need to be run as root."
    echo "Re-running setup script using sudo..."
    sudo -HE /bin/bash "$0"
    exit $?
fi

if [ -z "$PRIVATE_IFACE" ]; then
    PRIVATE_IFACE="eth1"
fi
echo "Using network interface: ${PRIVATE_IFACE}"

CNI_VERSION="v0.7.4"
CRICTL_VERSION="v1.13.0"
K8S_BINARIES=(kubeadm kubelet kubectl)
HOSTNAME=$(hostname)

CNI_TARBALL_FILE="cni-plugins-amd64-${CNI_VERSION}.tgz"
CRICTL_TARBALL_FILE="crictl-${CRICTL_VERSION}-linux-amd64.tar.gz"
CRICTL_TARBALL_SHA256="9bdbea7a2b382494aff2ff014da328a042c5aba9096a7772e57fdf487e5a1d51 $CRICTL_TARBALL_FILE"

echo "Getting Kubernetes stable release version..."
K8S_RELEASE=$(curl -fsSL https://dl.k8s.io/release/stable.txt)
echo "Using Kubernetes version ${K8S_RELEASE}"

echo "Getting private IP..."
if [ -n "$COREOS_PRIVATE_IPV4" ]; then
    PRIVATE_IP="$COREOS_PRIVATE_IPV4"
else
    PRIVATE_IP=$(ip -f inet -o addr show $PRIVATE_IFACE | head -n 1 | awk '{ print $4 }' | cut -d/ -f1)
fi
if [ -z "$PRIVATE_IP" ]; then
    echo "Can't determine the private IP address. Exiting!" >&2
    exit 1
fi
echo "Using private IP: ${PRIVATE_IP}"

check_service_active() {
    systemctl is-active "$1" --quiet
    return $?
}

check_service_enabled() {
    systemctl list-unit-files | grep "$1" | grep -q enabled
    return $?
}

if ! check_service_active docker.service; then
    echo "Starting Docker service..."
    systemctl start docker.service
fi
if ! check_service_enabled docker.service; then
    echo "Enabling Docker service..."
    systemctl enable docker.service
fi

check_and_extract() {
    echo "Checking file integrity..."
    cd $(dirname "$1")
    sha256sum -c <(printf "$3")
    EXIT_CODE=$?
    cd - > /dev/null 2>&1

    if [ $EXIT_CODE -eq 0 ]; then
        echo "File integrity check passed."
        echo "Extracting tarball archive..."
        tar -xzf "$1" -C "$2"
        echo "Tarball archive extracted."
        return 0
    else
        echo "File integrity check failed."
        rm -f "$1"
        return 1
    fi
}

download_cni() {
    echo "Downloading CNI tarball archive..."
    curl -fsSL "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/${CNI_TARBALL_FILE}" -o "$1"
    echo "CNI tarball archive downloaded."
}

extract_cni() {
    if [ ! -f "$1" ]; then
        download_cni "$1"
    fi

    SHA256_HASH=$(curl -fsSL "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/${CNI_TARBALL_FILE}.sha256")
    if ! check_and_extract "$1" "$2" "$SHA256_HASH"; then
        echo "Retrying..."
        extract_cni $@
    fi
}

if [ ! -d /opt/cni/bin ]; then
    mkdir -p /opt/cni/bin
fi
extract_cni /tmp/$CNI_TARBALL_FILE /opt/cni/bin

download_cri() {
    echo "Downloading CRI..."
    curl -fsSL "https://github.com/kubernetes-incubator/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz" -o "$1"
    echo "CRI downloaded."
}

extract_cri() {
    if [ ! -f "$1" ]; then
        download_cri "$1"
    fi

    if ! check_and_extract "$1" "$2" "$CRICTL_TARBALL_SHA256"; then
        echo "Retrying..."
        extract_cri $@
    fi
}

if [ ! -d /opt/bin ]; then
    mkdir -p /opt/bin
fi
extract_cri /tmp/$CRICTL_TARBALL_FILE /opt/bin

download_k8s_binary() {
    BINARY_NAME="$1"
    BINARY_PATH="$2"
    echo "Downloading Kubernetes binary: ${BINARY_NAME}..."
    curl -fsSL "https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/amd64/${BINARY_NAME}" -o $BINARY_PATH
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
    curl -fsSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${K8S_RELEASE}/build/debs/kubelet.service" | sed "s:/usr/bin:/opt/bin:g" > /etc/systemd/system/kubelet.service
    curl -fsSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${K8S_RELEASE}/build/debs/10-kubeadm.conf" | sed "s:/usr/bin:/opt/bin:g" > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
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

if ! grep -q $HOSTNAME /etc/hosts; then
    echo "Adding host DNS into /etc/hosts..."
    echo "$PRIVATE_IP $HOSTNAME" >> /etc/hosts
    echo "Host DNS added."
fi
