#!/bin/bash
set -e

if [ -d /tmp/gbf-proxy ]; then
    rm -rf /tmp/gbf-proxy
fi
mkdir -p /tmp/gbf-proxy

echo "Extracting archive..."
tar -C /tmp/gbf-proxy -f /tmp/gbf-proxy.tar.gz -xz
echo "Archive extracted."

cd /tmp/gbf-proxy
if [ -n "$(docker images -q gbf-proxy:latest)" ]; then
    docker rmi gbf-proxy:latest
fi

echo "Building Docker image..."
docker build -qt gbf-proxy:latest .
echo "Docker image built."
