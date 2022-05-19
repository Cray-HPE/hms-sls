FROM artifactory.algol60.net/docker.io/library/alpine:3.15

RUN set -x \
    && apk -U upgrade \
    && apk add --no-cache \
        bash \
        curl \
        jq

COPY load-sls.sh /src/app/load-sls.sh
COPY data/hardware /src/app/data/hardware
COPY data/networks /src/app/data/networks

WORKDIR /src/app
# Run as nobody
RUN chown -R 65534:65534 /src
USER 65534:65534

# this is inherited from the hms-test container
CMD [ "/src/app/load-sls.sh" ]