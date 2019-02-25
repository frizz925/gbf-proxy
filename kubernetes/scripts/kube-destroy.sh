#!/bin/bash
set -e

SCRIPTS_DIR=$(realpath $(dirname $0))
DEPLOYMENTS_DIR=$(realpath $SCRIPTS_DIR/../deployments)
PROJECT_DIR=$(realpath $SCRIPTS_DIR/../..)

delete_configmap() {
    CONFIGMAP_NAME="$1"
    echo "Deleting configmap: ${CONFIGMAP_NAME}..."
    kubectl delete configmap $CONFIGMAP_NAME || true
    echo "Configmap deleted: ${CONFIGMAP_NAME}."
}

delete_secrets() {
    echo "Deleting secrets..."
    kubectl delete secret gbf-proxy || true
    echo "Secrets deleted."
}

echo "Destroying application..."
cd $DEPLOYMENTS_DIR
bash $SCRIPTS_DIR/kube-delete.sh
cd - > /dev/null
echo "Application destroyed."

delete_configmap redis
delete_configmap nginx
delete_secrets
