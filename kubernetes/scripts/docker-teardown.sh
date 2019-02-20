#!/bin/bash

PROJECT_DIR="/tmp/gbf-proxy"

if [ -d $PROJECT_DIR ]; then
    echo "Removing project directory..."
    rm -rf $PROJECT_DIR
fi
