#!/bin/bash
set -e

DEPLOYMENTS=(gbf-proxy gbf-proxy-controller gbf-proxy-cache)
IMAGE_NAME=gbf-proxy
IMAGE_TAG=$(git describe --always --long --dirty)

for d in ${DEPLOYMENTS[@]}; do
    kubectl set image deployments/$d $d=$IMAGE_NAME:$IMAGE_TAG
done

WEB_NAME=gbf-proxy-web
kubectl set image depoyments/$WEB_NAME $WEB_NAME=$WEB_NAME:$IMAGE_TAG

