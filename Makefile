GO = GOOS=linux GOARCH=amd64 GO111MODULE=on go
SERVICE_NAME = tempo-device-service
GOFLAGS = -ldflags "-X github.impcloud.net/RSP-Inventory-Suite/tempo-device-service/cmd/main.Version=1.0.0"

.PHONY: run default cmd/tempo-device-service

default: build image run

build: cmd/tempo-device-service
cmd/tempo-device-service:
	 $(GO) build $(GOFLAGS) -o $@ ./cmd

image: Dockerfile go.mod go.sum
	docker build -t $(SERVICE_NAME) .

run:
	docker run --rm -it -p 9001:9001 -p 49993:49993 $(SERVICE_NAME) --profile=dev
