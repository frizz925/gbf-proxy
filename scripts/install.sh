#!/bin/bash

CWD=$(pwd)
cd "$(dirname $0)/../golang"
dep ensure
cd "$CWD"
