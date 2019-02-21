#!/bin/bash
set -e

build_project() {
    PROJECT_NAME="$1"
    PROJECT_DIR="/tmp/$PROJECT_NAME"
    TARBALL_PATH="/tmp/$PROJECT_NAME.tar.gz"
    IMAGE_NAME="${PROJECT_NAME}:latest"
    DOCKERFILE_PATH="$PROJECT_DIR/$2/Dockerfile"

    echo "Building project: ${PROJECT_NAME}..."
    if [ ! -d $PROJECT_DIR ]; then
        echo "Creating project directory: ${PROJECT_DIR}..."
        mkdir -p $PROJECT_DIR
    fi

    if [ ! -e $TARBALL_PATH ]; then
        echo "Tarball file not found at '${TARBALL_PATH}'" >&2
        exit 1
    fi

    echo "Extracting tarball: ${TARBALL_PATH}..."
    tar -C $PROJECT_DIR -f $TARBALL_PATH -xz

    if [ -n "$(docker images -q $IMAGE_NAME)" ]; then
        echo "Removing existing docker image: ${IMAGE_NAME}..."
        docker rmi $IMAGE_NAME
    fi

    echo "Building Docker image: ${IMAGE_NAME}..."
    docker build -qt $IMAGE_NAME -f $DOCKERFILE_PATH $PROJECT_DIR
    echo "Docker image built: ${IMAGE_NAME}."
    echo "Project built: ${PROJECT_NAME}."
}

build_project gbf-proxy
build_project gbf-proxy-web web-docker
