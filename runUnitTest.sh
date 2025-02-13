#!/usr/bin/env bash
# MIT License
#
# (C) Copyright [2021,2025] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

test_result=0

# Print executions
set -x

# Need to setup a working DB environment
PROJECT=$RANDOM
NETWORK_NAME="${PROJECT}_default"
INIT_CONTAINER_NAME="${PROJECT}_sls-init_1"

docker compose --project-name $PROJECT -f docker-compose.testing.yaml up -d --build
if [[ $? -ne 0 ]]; then
    echo "Failed to setup environment!"
    exit 1
fi

docker wait $INIT_CONTAINER_NAME

# Build the build base image (if it's not already)
docker build -t cray/sls-base --target base .

# Run the tests.
DOCKER_BUILDKIT=0 docker build --network $NETWORK_NAME -t cray/sls-unit-testing -f Dockerfile.testing --no-cache .
build_result=$?
if [ $build_result -ne 0 ]; then
  echo "Unit tests failed!"
  test_result=$build_result
else
  echo "Unit tests passed!"
fi

# Cleanup.
docker compose --project-name $PROJECT -f docker-compose.testing.yaml down

exit $test_result
