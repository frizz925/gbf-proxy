#!/bin/bash
set -e

CWD=$(pwd)
SCRIPT_DIR=$(dirname "$0")
FILES_DIR=$(realpath "$SCRIPT_DIR/../files")
PROJECT_DIR=$(realpath "$1")

TAR_FILENAME="gbf-proxy.tar.gz"
BIN_FILENAME="gbf-proxy-alpine-linux-amd64"
TAR_PATH="$FILES_DIR/$TAR_FILENAME"

cleanup() {
    cd "$CWD"
}
trap cleanup EXIT

cd "$PROJECT_DIR/golang"
make clean
make deps
make test
make build-alpine-linux

if [ -f "$TAR_PATH" ]; then
    rm -f "$TAR_PATH"
fi
tar -czf "$TAR_PATH" Dockerfile bin/$BIN_FILENAME
