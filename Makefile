SERVICE_NAME=tempo-device-service
MODULE_NAME?=$(shell go list -m)
VERSION?=$(shell cat ./VERSION)

GO=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go
GOFLAGS:=-ldflags "-X $(MODULE_NAME)/cmd/main.Version=$(VERSION)"
TAGS?=$(VERSION) dev latest
LABELS?="git_sha=$(shell git rev-parse HEAD)"

.PHONY: default run

default: build image run

build: $(SERVICE_NAME)
$(SERVICE_NAME):
	 $(GO) build $(GOFLAGS) -o $@ $(MODULE_NAME)/cmd

image: Dockerfile go.mod go.sum
	docker build -t $(SERVICE_NAME):$(VERSION) .

run:
	docker run -p $(SERVICE_NAME):$(VERSION)
