#!/bin/bash
set -e

SCRIPTS_DIR=$(realpath $(dirname $0))
DEPLOYMENTS_DIR=$(realpath $SCRIPTS_DIR/../deployments)
PROJECT_DIR=$(realpath $SCRIPTS_DIR/../..)

TLS_CERT_PATH=$PROJECT_DIR/certs/fullchain.pem
TLS_PRIVATE_KEY_PATH=$PROJECT_DIR/certs/privkey.pem

create_configmap() {
    CONFIGMAP_NAME="$1"
    CONFIGMAP_FILE="$2"
    echo "Creating configmap: ${CONFIGMAP_NAME}..."
    kubectl create configmap $CONFIGMAP_NAME --from-file="$CONFIGMAP_FILE"
    echo "Configmap created: ${CONFIGMAP_NAME}."
}

create_secrets() {
    echo "Creating secrets..."
    kubectl create secret generic gbf-proxy \
        --from-file=dhparam.pem=<(openssl dhparam -dsaparam 4096) \
        --from-file=tls.key=$1 \
        --from-file=tls.crt=$2
    echo "Secrets created."
}

create_configmap redis $PROJECT_DIR/redis
create_configmap nginx $PROJECT_DIR/nginx
create_secrets $TLS_PRIVATE_KEY_PATH $TLS_CERT_PATH

echo "Deploying application..."
cd $DEPLOYMENTS_DIR
bash $SCRIPTS_DIR/kube-apply.sh
cd - > /dev/null
echo "Application deployed."
