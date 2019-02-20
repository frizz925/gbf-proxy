#!/bin/bash
set -e
CWD=$(pwd)
SCRIPT_PID=$$

# Testing using golang
cd "$(dirname $0)/../golang"
echo "Running tests with coverage..."
go test -race -coverprofile=coverage.txt -covermode=atomic ./...
echo "OK"

# Testing using external tool
# Setup
CACHE_ADDRESS="127.0.0.1:8001"
CONTROLLER_ADDRESS="127.0.0.1:8000"
PROXY_ADDRESS="127.0.0.1:8088"
ASSET_URL="http://game-a3.granbluefantasy.jp/assets/1548846641/js/config.js"
BIN_DIR="/tmp"
BIN_NAME="gbf-proxy-$(date +%s)"
BIN_EXEC="$BIN_DIR/$BIN_NAME"

echo "Building binary..."
go build -race -o "$BIN_EXEC" -v

cleanup() {
    killall $BIN_NAME
    [ -e "$BIN_EXEC" ] && rm $BIN_EXEC
    cd "$CWD"
}
trap cleanup EXIT


run() {
    $BIN_DIR/$BIN_NAME $@
}

request() {
    curl -fsSL -x http://$PROXY_ADDRESS $@
}

echo "Spinning up cache service at $CACHE_ADDRESS"
run cache $CACHE_ADDRESS 2> /dev/null &
CACHE_PID=$!
sleep 1

echo "Spinning up controller service at $CONTROLLER_ADDRESS..."
run controller $CONTROLLER_ADDRESS -c $CACHE_ADDRESS -w example.org:80 --web-hostname example.org 2> /dev/null &
CONTROLLER_PID=$!
sleep 1

echo "Spinning up proxy service at $PROXY_ADDRESS..."
run proxy $PROXY_ADDRESS $CONTROLLER_ADDRESS 2> /dev/null &
PROXY_PID=$!
sleep 1

sleep 3
pkill -0 -P $CACHE_PID > /dev/null 2>&1 || (echo "Cache service fails to run!" >&2 && exit 1)
pkill -0 -P $CONTROLLER_PID > /dev/null 2>&1 || (echo "Controller service fails to run!" >&2 && exit 1)
pkill -0 -P $PROXY_PID > /dev/null 2>&1 || (echo "Proxy service fails to run!" >&2 && exit 1)

# Test game server
printf "Testing game server... "
request game.granbluefantasy.jp | grep -q "グランブルーファンタジー" && echo "OK" || (echo "FAIL!" && exit 1)
# Test game assets server
printf "Testing game assets server... "
request $ASSET_URL > /dev/null && echo "OK" || (echo "FAIL!" && exit 1)
# Test game assets caching
printf "Testing game assets caching... "
request $ASSET_URL > /dev/null && echo "OK" || (echo "FAIL!" && exit 1)
# Test static web
printf "Testing static web... "
request example.org | grep -q "Example Domain" && echo "OK" || (echo "FAIL!" && exit 1)
# Test forbidden
printf "Testing forbidden host... "
request github.com 2> /dev/null && echo "Not forbidden!" && exit 1 || echo "OK"
