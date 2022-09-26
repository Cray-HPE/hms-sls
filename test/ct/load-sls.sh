#!/bin/bash

# MIT License
#
# (C) Copyright [2022] Hewlett Packard Enterprise Development LP
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

# load-sls.sh; used by runCT.sh to populate SLS with data before running api tests.
echo "Initiating..."

HARDWARE_URL="http://cray-sls:8376/v1/hardware"
HARDWARE_DIR="data/hardware"
NETWORKS_URL="http://cray-sls:8376/v1/networks"
NETWORKS_DIR="data/networks"

# check for SLS hardware data directory
if [[ ! -d ${HARDWARE_DIR} ]] ; then
    echo "Failed to find SLS hardware directory: ${HARDWARE_DIR}"
    exit 1
fi

# check for SLS networks data directory
if [[ ! -d ${NETWORKS_DIR} ]] ; then
    echo "Failed to find SLS networks directory: ${NETWORKS_DIR}"
    exit 1
fi

echo "Loading SLS hardware data..."

# load hardware data into SLS
for HARDWARE_FILE in $(ls ${HARDWARE_DIR}) ; do
    FILE_PATH="${HARDWARE_DIR}/${HARDWARE_FILE}"
    CURL_CMD="curl -s -d @${FILE_PATH} -X POST ${HARDWARE_URL}"
    CURL_OUT=$(eval ${CURL_CMD})
    CURL_RET=$?
    INSERT_CHECK=$(echo "${CURL_OUT}" | grep "inserted new entry")
    if [[ ${CURL_RET} -ne 0 ]] ; then
        echo "Failed to 'POST' SLS hardware data file ${HARDWARE_FILE}, curl returned error code: ${CURL_RET}"
        exit 1
    elif [[ -z ${INSERT_CHECK} ]] ; then
        echo "Failed to 'POST' SLS hardware data file ${HARDWARE_FILE}, SLS did not return expected successful insert message."
        exit 1
    fi
done

echo "Loading SLS network data..."

# load networks data into SLS
for NETWORK_FILE in $(ls ${NETWORKS_DIR}) ; do
    FILE_PATH="${NETWORKS_DIR}/${NETWORK_FILE}"
    CURL_CMD="curl -s -d @${FILE_PATH} -X POST ${NETWORKS_URL}"
    CURL_OUT=$(eval ${CURL_CMD})
    CURL_RET=$?
    if [[ ${CURL_RET} -ne 0 ]] ; then
        echo "Failed to 'POST' SLS networks data file ${NETWORK_FILE}, curl returned error code: ${CURL_RET}"
        exit 1
    fi
done

echo "SLS data loaded, exiting with code: 0"
exit 0
