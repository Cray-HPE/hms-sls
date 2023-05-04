# MIT License
#
# (C) Copyright [2021-2022] Hewlett Packard Enterprise Development LP
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

# import config.
# You can change the default config with `make cnf="config_special.env" build`
cnf ?= config.env
include $(cnf)
export $(shell sed 's/=.*//' $(cnf))

ifndef NAME
$(error NAME is not set.  Please review and copy config.env.default to config.env and try again)
endif

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help build-image touch.build-image

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# Set the basic docker build command

DOCKER_BUILD_CMD=docker build --build-arg REGISTRY_HOST=${REGISTRY_HOST} ${DOCKER_ARGS}

all : image unittest ct snyk ct_image

image: touch.build-image .build-image ## Build the service image

.build-image: Dockerfile 
	${DOCKER_BUILD_CMD} --no-cache --pull --tag '${NAME}:${VERSION}' .
	${DOCKER_BUILD_CMD} --target builder --tag '${NAME}-build:${VERSION}' .
	@touch .build-image

test-image: image  ## Build the image for running tests
	${DOCKER_BUILD_CMD} --tag '${NAME}-test:${VERSION}' -f Dockerfile.testing
	

unittest: image ## Run the unittests^Wintegration tests
	./runUnitTest.sh

snyk: ## Run Snyk
	./runSnyk.sh

ct: ## Run Conformance Tests
	./runCT.sh

ct_image:
	docker build --no-cache -f test/ct/Dockerfile test/ct/ --tag hms-sls-hmth-test:${VERSION}

sls-loader-helper:
	docker build --no-cache -f test/ct/Dockerfile.load-sls.Dockerfile test/ct/ --tag hms-sls-loader:${VERSION}

touch.build-image:
	$(eval timestamp=$(shell docker inspect -f '{{ range $$i, $$e := split .Metadata.LastTagTime "T" }}{{if eq $$i 0}}{{range $$j, $$v := split $$e "-"}}{{$$v}}{{end}}{{else}}{{$$f := printf "%.8s" $$e}}{{range $$j, $$g := split $$f ":"}}{{if lt $$j 2}}{{$$g}}{{else}}.{{$$g}}{{end}}{{end}}{{end}}{{end}}' $(DOCKER_TAG) 2>/dev/null ))
	@if [ "A$(timestamp)A" = "AA" ] ; then rm -f .build-image ; else touch -t $(timestamp) .build-image ; fi
