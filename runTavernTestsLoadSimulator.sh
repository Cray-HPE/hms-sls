#!/bin/bash
make sls-loader-helper
VERSION=$(cat .version)
docker run --rm -it --network hms-simulation-environment_simulation hms-sls-loader:${VERSION}