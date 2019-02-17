#!/bin/bash
set -e

CWD=$(pwd)
SCRIPT_DIR=$(dirname "$0")
FILES_DIR=$(realpath "$SCRIPT_DIR/../files")
PROJECT_DIR=$(realpath "$1")

cleanup() {
    cd "$CWD"
}
trap cleanup EXIT

cd "$PROJECT_DIR/golang"
make clean
make deps
make test
make build-linux

tar -czf "$FILES_DIR/gbf-proxy.tar.gz" Dockerfile bin
