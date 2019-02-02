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
echo "Spinning up server at port :8000..."
go run ./main.go 8000 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"
sleep 3
pkill -0 -P $SERVER_PID > /dev/null 2>&1 || (echo "Server fails to run!" >&2 && exit 1)

# Test allowed
printf "Testing allowed host... "
curl -six http://localhost:8000 game.granbluefantasy.jp | grep -q "グランブルーファンタジー" && echo "OK"
# Test forbidden
printf "Testing forbidden host... "
curl -six http://localhost:8000 github.com | grep -q "403 Forbidden" && echo "OK"

# Teardown
echo "Killing server..."
pkill -9 -P $SERVER_PID > /dev/null 2>&1

# Done
echo "Done."
cd "$CWD"
