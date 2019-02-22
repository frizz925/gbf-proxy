#!/bin/bash
set -e

SCRIPT_DIR=$(dirname $0)
VERSION_FILENAME="gbf-proxy-version"
VERSION_PATH="$SCRIPT_DIR/$VERSION_FILENAME"

if [ -z "$VERSION" ]; then
    if [ -n "$1" ]; then
        VERSION="$1"
    elif [ -e "$VERSION_PATH" ]; then
        VERSION=$(cat "$VERSION_PATH")
    else
        VERSION="latest"
    fi
fi

build_project() {
    PROJECT_NAME="$1"
    PROJECT_DIR="/tmp/$PROJECT_NAME"
    TARBALL_PATH="/tmp/$PROJECT_NAME.tar.gz"
    IMAGE_NAME_LATEST="${PROJECT_NAME}:latest"
    IMAGE_NAME="${PROJECT_NAME}:${VERSION}"
    DOCKERFILE_PATH="$PROJECT_DIR/Dockerfile"

    if [ -n "$(docker images -q $IMAGE_NAME)" ]; then
        echo "Docker image '${IMAGE_NAME}' already exists. Exiting..." >&2
        exit 1
    fi

    if [ -n "$2" ]; then
        DOCKERFILE_PATH="$PROJECT_DIR/$2/Dockerfile"
    fi

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

    echo "Building Docker image: ${IMAGE_NAME}..."
    docker build -qt $IMAGE_NAME -f $DOCKERFILE_PATH $PROJECT_DIR
    echo "Docker image built: ${IMAGE_NAME}."

    if [ -z "$(docker images -q $IMAGE_NAME_LATEST)" ]; then
        echo "Tagging Docker image as latest: ${IMAGE_NAME}..."
        docker tag $IMAGE_NAME $IMAGE_NAME_LATEST
    fi

    echo "Project built: ${PROJECT_NAME}."
}

build_project gbf-proxy
build_project gbf-proxy-web web-docker
