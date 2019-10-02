SERVICE_NAME = tempo-device-service
VERSION=1.0.0

GO = GOOS=linux GOARCH=amd64 GO111MODULE=auto go
GOFLAGS = -ldflags "-X github.impcloud.net/RSP-Inventory-Suite/$(SERVICE_NAME)/main.Version=$(VERSION)"

RUN_PROFILE=docker
RUN_FLAGS=--rm -it -p 9001:9001 -p 49993:49993

.PHONY: default run

default: build image run

build: $(SERVICE_NAME)
$(SERVICE_NAME):
	 $(GO) build $(GOFLAGS) -o $@ .

image: Dockerfile go.mod go.sum
	docker build -t $(SERVICE_NAME):$(VERSION) .

run:
	docker run $(RUN_FLAGS) $(SERVICE_NAME) --profile="$(RUN_PROFILE)"
