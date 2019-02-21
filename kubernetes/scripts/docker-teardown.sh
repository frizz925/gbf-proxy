#!/bin/bash

for d in gbf-proxy gbf-proxy-web; do
    PROJECT_DIR="/tmp/$d"
    if [ -d $PROJECT_DIR ]; then
        echo "Removing project directory..."
        rm -rf $PROJECT_DIR
    fi
done
