#!/bin/bash
set -e

NODE_NAME="$1"
if [ -z "$NODE_NAME" ]; then
    echo "Node name argument required." >&2
    exit 1
fi

if kubectl get nodes $NODE_NAME -o name > /dev/null 2>&1; then
  echo "Removing node: $NODE_NAME..."
  kubectl drain $NODE_NAME --delete-local-data --force --ignore-daemonsets
  kubectl delete node $NODE_NAME
  echo "Node removed: $NODE_NAME."
fi
