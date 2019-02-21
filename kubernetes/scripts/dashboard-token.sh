#!/bin/bash
set -e

SCRIPTS_DIR=$(realpath $(dirname $0))
DEPLOYMENT_DIR=$(realpath $SCRIPTS_DIR/../deployment)
DASHBOARD_USER=$(kubectl get secret -n kube-system | grep admin-user | awk '{ print $1 }')
if [ -z "$DASHBOARD_USER" ]; then
    exit 1
else
    kubectl describe secret -n kube-system $DASHBOARD_USER | awk '/token:/ { print $2 }'
fi
