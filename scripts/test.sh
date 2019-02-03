#!/bin/bash
set -e
CWD=$(pwd)

# Testing using golang
cd "$(dirname $0)/golang"
echo "Running test with coverage..."
make coverage
echo "OK"

# Testing using external tool
# Setup
CONTROLLER_ADDRESS="localhost:8000"
PROXY_ADDRESS="localhost:8088"

echo "Spinning up controller service at $CONTROLLER_ADDRESS..."
go run ./main.go controller $CONTROLLER_ADDRESS &
CONTROLLER_PID=$!

echo "Spinning up proxy service at $PROXY_ADDRESS..."
go run ./main.go proxy $PROXY_ADDRESS $CONTROLLER_ADDRESS &
PROXY_PID=$!

cleanup() {
    echo "Killing servers..."
    pkill -9 -P $CONTROLLER_PID > /dev/null 2>&1 || true
    pkill -9 -P $PROXY_PID > /dev/null 2>&1 || true
    echo "Done."
}
trap cleanup EXIT

sleep 3
pkill -0 -P $CONTROLLER_PID > /dev/null 2>&1 || (echo "Controller service fails to run!" >&2 && exit 1)
pkill -0 -P $PROXY_PID > /dev/null 2>&1 || (echo "Proxy service fails to run!" >&2 && exit 1)

# Test allowed
printf "Testing allowed host... "
curl -six http://$PROXY_ADDRESS game.granbluefantasy.jp | grep -q "グランブルーファンタジー" && echo "OK"
# Test forbidden
printf "Testing forbidden host... "
curl -six http://$PROXY_ADDRESS github.com | grep -q "403 Forbidden" && echo "OK"

cd "$CWD"
