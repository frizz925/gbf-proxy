#!/bin/bash
set -e

for f in $(find . -type f -iname '*.yaml' | sort); do
    kubectl apply -f $f
done
