# Apache v2 license
#  Copyright (C) <2019> Intel Corporation
#
#  SPDX-License-Identifier: Apache-2.0
#
SERVICE_NAME=tempo-device-service
MODULE_NAME?=$(shell go list -m)
VERSION?=$(shell cat ./VERSION)

GO=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go
GOFLAGS:=-ldflags "-X $(MODULE_NAME)/cmd/main.Version=$(VERSION)"
TAGS?=$(VERSION) dev latest
LABELS?="git_sha=$(shell git rev-parse HEAD)"
RUN_FLAGS=--rm -it -P

BUILD_ARGS=--build-arg http_proxy=$(http_proxy) \
			--build-arg https_proxy=$(https_proxy)

.PHONY: default run image clean

default: image run

# create a local build of the service
build: $(SERVICE_NAME)
$(SERVICE_NAME):
	 $(GO) build $(GOFLAGS) -o $@ $(MODULE_NAME)/cmd
	 chmod 0700 $@

# delete a local build, if it exists
clean:
	-rm $(SERVICE_NAME)

# build the service in a Dockerfile and turn it into an image
image:
	docker build $(BUILD_ARGS) -t $(SERVICE_NAME):$(VERSION) .

# run the Docker version of the service
run:
	docker run $(RUN_FLAGS) \
		--name=$(SERVICE_NAME)_$(VERSION) \
		$(SERVICE_NAME):$(VERSION)
