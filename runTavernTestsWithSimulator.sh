#!/bin/bash
make ct_image
VERSION=$(cat .version)
docker run --rm -it --network hms-simulation-environment_simulation hms-sls-hmth-test:${VERSION} tavern -c /src/app/tavern_global_config_ct_test.yaml -p /src/app/api #/2-disruptive
