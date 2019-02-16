#!/bin/bash
set -e

CWD=$(pwd)
PROJECT_DIR="$1"

cleanup() {
    cd $CWD
}
trap cleanup EXIT

cd $PROJECT_DIR/golang
make deps
make test
make build-linux
