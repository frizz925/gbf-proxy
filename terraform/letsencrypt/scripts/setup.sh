#!/bin/bash
set -e

CERTBOT_IMAGE_NAME="certbot/certbot"
NGINX_IMAGE_NAME="nginx:1.15-alpine"

CERTBOT_CONTAINER_NAME="certbot"
NGINX_CONTAINER_NAME="certbot-nginx"

DEFAULT_WEBROOT=/var/www/certbot
DEFAULT_CERTS_DIR=/etc/letsencrypt
REQUIRED_ENVS=(LETSENCRYPT_EMAIL LETSENCRYPT_DOMAINS)

if [ $EUID -ne 0 ]; then
    echo "Setup script need to be run as root" >&2
    exit 1
fi

docker_check_container() {
    test -n "$(docker ps -aqf name=$1)"
    return $?
}

docker_check_image() {
    test -n "$(docker images -q $1)"
    return $?
}

check_env() {
    test -n "$(printenv $1)"
    return $?
}

cleanup() {
    echo "Cleaning things up..."
    for NAME in $CERTBOT_CONTAINER_NAME $NGINX_CONTAINER_NAME; do
        if docker_check_container $NAME; then
            echo "Removing Docker container: ${NAME}..."
            docker rm -f $NAME > /dev/null
            echo "Docker container removed: ${NAME}."
        fi
    done
}

for ENV in ${REQUIRED_ENVS[@]}; do
    if ! check_env $ENV; then
        echo "Missing required environment variable: $ENV" >&2
        exit 1
    fi
done

if [ -z "$LETSENCRYPT_WEBROOT" ]; then
    LETSENCRYPT_WEBROOT="$DEFAULT_WEBROOT"
fi

if [ -z "$LETSENCRYPT_CERTS_DIR" ]; then
    LETSENCRYPT_CERTS_DIR="$DEFAULT_CERTS_DIR"
fi

cleanup
trap cleanup EXIT

for DIR in $LETSENCRYPT_WEBROOT $LETSENCRYPT_CERTS_DIR; do
    if [ ! -d $DIR ]; then
        echo "Creating directory: ${DIR}..."
        mkdir -p $DIR
        echo "Directory created: ${DIR}."
    fi
done

for IMAGE in $CERTBOT_IMAGE_NAME $NGINX_IMAGE_NAME; do
    if ! docker_check_image $IMAGE; then
        echo "Pulling the required Docker image: ${IMAGE}..."
        docker pull $IMAGE > /dev/null
        echo "Pulled Docker image: ${IMAGE}."
    fi
done

EMAIL_ARG="-m ${LETSENCRYPT_EMAIL}"
WEBROOT_ARG="-w ${DEFAULT_WEBROOT}"
DOMAIN_ARGS=""
for DOMAIN in $LETSENCRYPT_DOMAINS; do
    DOMAIN_ARGS="$DOMAIN_ARGS -d $DOMAIN"
done
CERTBOT_ARGS="certonly -n --webroot --agree-tos $EMAIL_ARG $WEBROOT_ARG $DOMAIN_ARGS"

echo "Email: $LETSENCRYPT_EMAIL"
echo "Domains: $LETSENCRYPT_DOMAINS"
echo "Webroot: $LETSENCRYPT_WEBROOT"
echo "Certs dir: $LETSENCRYPT_CERTS_DIR"
echo "Certbot args: $CERTBOT_ARGS"
if [ -t 0 ]; then
    read -p "Are these settings correct? (y/n) " -n 1 CONFIRM
    echo
    if [ "$CONFIRM" != "y" ]; then
        echo "Cancelled."
        exit 1
    fi
fi

echo "Starting Nginx web server container..."
docker run --name $NGINX_CONTAINER_NAME -d \
    -v $LETSENCRYPT_WEBROOT:/usr/share/nginx/html:ro \
    -p 80:80 \
    $NGINX_IMAGE_NAME > /dev/null

echo "Starting certbot container..."
docker run --name $CERTBOT_CONTAINER_NAME \
    -v $LETSENCRYPT_WEBROOT:$DEFAULT_WEBROOT \
    -v $LETSENCRYPT_CERTS_DIR:$DEFAULT_CERTS_DIR \
    $CERTBOT_IMAGE_NAME $CERTBOT_ARGS

echo "Done."
