#!/bin/bash
set -e

CWD=$(pwd)
SCRIPT_DIR=$(dirname "$0")
FILES_DIR=$(realpath "$SCRIPT_DIR/../files")
PROJECT_DIR=$(realpath "$1")
TAR_PATH="$FILES_DIR/gbf-proxy.tar.gz"

cleanup() {
    cd "$CWD"
}
trap cleanup EXIT

cd "$PROJECT_DIR/golang"
make clean
make deps
make test
make build-linux

if [ -f "$TAR_PATH" ]; then
    rm "$TAR_PATH"
fi
tar -czf "$TAR_PATH" Dockerfile bin/gbf-proxy-linux-amd64
