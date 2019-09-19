GO = GOOS=linux GOARCH=amd64 GO111MODULE=on go
SERVICE_NAME = tempo-device-service

.PHONY: run all

all: build image run

build: server.go
	 $(GO) build -o server

image: Dockerfile
	docker build -t $(SERVICE_NAME) .

run:
	docker run --rm -it --net host $(SERVICE_NAME) -p 9001
