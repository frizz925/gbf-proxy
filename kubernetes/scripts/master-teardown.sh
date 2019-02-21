#!/bin/bash
set -e

SCRIPTS_DIR=$(dirname $0)
GO_TEMPLATE='{{range $_, $n := .items}}{{$n.metadata.name}} {{end}}'
for n in $(kubectl get nodes -o go-template="$GO_TEMPLATE"); do
    bash $SCRIPTS_DIR/node-teardown.sh $n
done
