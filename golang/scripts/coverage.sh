#!/bin/sh
set -e
PKG_EXC_PATTERN="(golang$|cmd)"

cd $(dirname $0)/..
COVER_PKGS=$(go list ./... | grep -v -E "$PKG_EXC_PATTERN" | tr '\n' ',')
go test -race -coverprofile=coverage.txt \
    -covermode=atomic \
    -coverpkg=$COVER_PKGS ./...
cd - >/dev/null
