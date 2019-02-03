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
echo "Spinning up controller service at port localhost:8000..."
go run ./main.go controller localhost:8000 &
CONTROLLER_PID=$!

echo "Spinning up proxy service at port localhost:8080..."
go run ./main.go proxy localhost:8080 localhost:8000 &
PROXY_PID=$!

cleanup() {
    echo "Killing servers..."
    pkill -9 -P $CONTROLLER_PID > /dev/null 2>&1 || exit 0
    pkill -9 -P $PROXY_PID > /dev/null 2>&1 || exit 0
    echo "Done."
}
trap cleanup EXIT

sleep 3
pkill -0 -P $CONTROLLER_PID > /dev/null 2>&1 || (echo "Controller service fails to run!" >&2 && exit 1)
pkill -0 -P $PROXY_PID > /dev/null 2>&1 || (echo "Proxy service fails to run!" >&2 && exit 1)

# Test allowed
printf "Testing allowed host... "
curl -six http://localhost:8080 game.granbluefantasy.jp | grep -q "グランブルーファンタジー" && echo "OK"
# Test forbidden
printf "Testing forbidden host... "
curl -six http://localhost:8080 github.com | grep -q "403 Forbidden" && echo "OK"

cd "$CWD"
