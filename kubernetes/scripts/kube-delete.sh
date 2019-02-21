#!/bin/bash
set -e

for f in $(find . -type f -iname '*.yaml' | sort -r); do
    kubectl delete -f $f || true
done
