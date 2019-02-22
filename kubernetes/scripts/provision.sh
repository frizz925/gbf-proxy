#!/bin/bash
set -e

CWD=$(pwd)
SCRIPTS_DIR=$(realpath $(dirname "$0"))
FILES_DIR=$(realpath "$SCRIPTS_DIR/../files")
PROJECT_DIR=$(realpath "$1")

BIN_FILENAME="gbf-proxy-alpine-linux-amd64"
TAR_FILENAME="gbf-proxy.tar.gz"
WEB_TAR_FILENAME="gbf-proxy-web.tar.gz"
VERSION_FILENAME="gbf-proxy-version"
TAR_PATH="$FILES_DIR/$TAR_FILENAME"
WEB_TAR_PATH="$FILES_DIR/$WEB_TAR_FILENAME"
VERSION_PATH="$FILES_DIR/$VERSION_FILENAME"

cleanup() {
    cd "$CWD"
}
trap cleanup EXIT

cd "$PROJECT_DIR/golang"
VERSION=$(make version | awk 'NR==1{ print $1 }' || printf latest)

echo "Building Granblue Proxy (version: ${VERSION})..."
make clean
make deps
make test
make build-alpine-linux
echo "Granblue Proxy built."

echo "Creating Granblue Proxy tarball..."
if [ -f "$TAR_PATH" ]; then
    rm -f "$TAR_PATH"
fi
tar -czf "$TAR_PATH" Dockerfile bin/$BIN_FILENAME
echo "Granblue Proxy tarball created."

echo "Creating static web tarball..."
cd "$PROJECT_DIR"
if [ -f "$WEB_TAR_PATH" ]; then
    rm -f "$WEB_TAR_PATH"
fi
tar -czf "$WEB_TAR_PATH" web-docker web
echo "Static web tarball created."

echo $(VERSION) > $VERSION_PATH
