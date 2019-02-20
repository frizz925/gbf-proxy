#!/bin/bash
set -e

PROJECT_DIR="/tmp/gbf-proxy"
TARBALL_PATH="/tmp/gbf-proxy.tar.gz"
IMAGE_NAME="gbf-proxy:latest"

if [ ! -d $PROJECT_DIR ]; then
    echo "Creating project directory..."
    mkdir -p $PROJECT_DIR
fi

if [ ! -e $TARBALL_PATH ]; then
    echo "Tarball file not found at '${TARBALL_PATH}'" >&2
    exit 1
fi

echo "Extracting tarball..."
tar -C $PROJECT_DIR -f $TARBALL_PATH -xz

if [ -n "$(docker images -q $IMAGE_NAME)" ]; then
    echo "Removing existing docker image..."
    docker rmi $IMAGE_NAME
fi

echo "Building Docker image..."
docker build -qt $IMAGE_NAME $PROJECT_DIR
echo "Docker image built."
