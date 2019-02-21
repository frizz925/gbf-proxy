#!/bin/bash
set -e

SCRIPTS_DIR=$(realpath $(dirname $0))
DEPLOYMENTS_DIR=$(realpath $SCRIPTS_DIR/../deployments)
PROJECT_DIR=$(realpath $SCRIPTS_DIR/../..)

create_configmap() {
    CONFIGMAP_NAME="$1"
    CONFIGMAP_FILE="$PROJECT_DIR/$2"
    echo "Creating configmap: ${CONFIGMAP_NAME}..."
    kubectl create configmap $CONFIGMAP_NAME --from-file="$CONFIGMAP_FILE"
    echo "Configmap created: ${CONFIGMAP_NAME}."
}

create_configmap redis redis/redis.conf
create_configmap nginx nginx/nginx.conf

echo "Deploying application..."
cd $DEPLOYMENTS_DIR
bash $SCRIPTS_DIR/kube-apply.sh
cd - > /dev/null
echo "Application deployed."
