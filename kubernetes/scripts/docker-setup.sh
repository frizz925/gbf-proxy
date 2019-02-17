#!/bin/bash

if [ -d /tmp/gbf-proxy ]; then
    rm -rf /tmp/gbf-proxy
fi
mkdir -p /tmp/gbf-proxy
tar -C /tmp/gbf-proxy -f /tmp/gbf-proxy.tar.gz -xz

cd /tmp/gbf-proxy
if docker images -q gbf-proxy:latest > /dev/null 2>&1; then
    docker rmi gbf-proxy:latest
fi
docker build -qt gbf-proxy:latest .
