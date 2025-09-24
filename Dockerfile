# MIT License
#
# (C) Copyright [2019-2022,2024-2025] Hewlett Packard Enterprise Development LP
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

# Dockerfile for building HMS SLS.

# Build base just has the packages installed we need.
FROM artifactory.algol60.net/docker.io/library/golang:1.24-alpine AS build-base

RUN set -ex \
    && apk -U upgrade \
    && apk add build-base

FROM build-base AS base

RUN go env -w GO111MODULE=auto

# Copy all the necessary files to the image.
COPY cmd $GOPATH/src/github.com/Cray-HPE/hms-sls/v2/cmd
COPY internal $GOPATH/src/github.com/Cray-HPE/hms-sls/v2/internal
COPY vendor $GOPATH/src/github.com/Cray-HPE/hms-sls/v2/vendor
COPY pkg $GOPATH/src/github.com/Cray-HPE/hms-sls/v2/pkg

### Build Stage ###

FROM base AS builder

# Now build
RUN set -ex \
    && go build -v -o sls github.com/Cray-HPE/hms-sls/v2/cmd/sls \
    && go build -v -o sls-init github.com/Cray-HPE/hms-sls/v2/cmd/sls-init \
    && go build -v -o sls-loader github.com/Cray-HPE/hms-sls/v2/cmd/sls-loader \
    && go build -v -o sls-migrator github.com/Cray-HPE/hms-sls/v2/cmd/sls-migrator \
    && go build -v -o sls-s3-downloader github.com/Cray-HPE/hms-sls/v2/cmd/sls-s3-downloader \
    && go build -v -o sls-benchmark github.com/Cray-HPE/hms-sls/v2/cmd/sls-benchmark

### Final Stage ###

FROM artifactory.algol60.net/csm-docker/stable/docker.io/library/alpine:3.22
LABEL maintainer="Hewlett Packard Enterprise"
STOPSIGNAL SIGTERM
EXPOSE 8376

# Default to latest schema version, this is overridden in the versioned chart.
ENV SCHEMA_VERSION=latest

RUN set -ex \
    && apk -U upgrade \
    && apk add --no-cache \
        curl \
        jq \
        bind-tools \
    && mkdir -p /persistent_migrations \
    && chmod 777 /persistent_migrations \
    && mkdir -p /sls && chown 65534:65534 /sls

# Copy files necessary for running/setup
COPY migrations /migrations
COPY configs /configs
COPY entrypoint.sh /

# Get sls from the builder stage.
COPY --from=builder /go/sls /usr/local/bin
COPY --from=builder /go/sls-init /usr/local/bin
COPY --from=builder /go/sls-loader /usr/local/bin
COPY --from=builder /go/sls-migrator /usr/local/bin
COPY --from=builder /go/sls-s3-downloader /usr/local/bin
COPY --from=builder /go/sls-benchmark /usr/local/bin

# nobody 65534:65534
USER 65534:65534

ENTRYPOINT ["/entrypoint.sh"]
CMD ["sls"]
