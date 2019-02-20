#!/bin/bash
set -e

TOKEN_CA_PATH=/etc/kubernetes/pki/ca.crt

TOKEN=$(kubeadm token list | awk 'NR==2{ print $1 }')
TOKEN_CA_HASH=$(openssl x509 -in $TOKEN_CA_PATH -pubkey | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | awk '{ print $2 }')

echo "TF_VAR_kube_token=${TOKEN}"
echo "TF_VAR_kube_hash=${TOKEN_CA_HASH}"
