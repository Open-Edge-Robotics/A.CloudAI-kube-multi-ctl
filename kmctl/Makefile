# Makefile for building and deploying a Go project

APP_VERSION=0.1.5

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
SERVER_BINARY_NAME=kube_backend
CLIENT_BINARY_NAME=kmctl
CLIENT_CONFIG_FILE=config.yaml
CLIENT_CONFIG_DIR=$(HOME)/.config/$(CLIENT_BINARY_NAME)
MAIN_GO_FILE=main.go

# Docker parameters
DOCKER_REGISTRY=lgecloudroboticstask
DOCKER_IMAGE_NAME=$(SERVER_BINARY_NAME)
DOCKER_TAG=$(shell date +%Y-%m-%d)

# Directories
SERVER_SRC_DIR=./server
CLIENT_SRC_DIR=./client
BUILD_DIR=../build

all: test build

pre-reqs:
ifeq (, $(shell which protoc))
	@echo "Installing protoc..."
	sudo apt update && sudo apt install -y protobuf-compiler
endif
ifeq (, $(shell which protoc-gen-go))
	@echo "Installing protoc-gen-go..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
endif
ifeq (, $(shell which protoc-gen-go-grpc))
	@echo "Installing protoc-gen-go-grpc..."
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
endif
ifeq (, $(shell echo $$PATH | grep -q $(HOME)/go/bin))
	@echo "Adding go/bin to PATH..."
	@echo 'export PATH=$$PATH:$(HOME)/go/bin' >> ~/.bashrc
endif

build: mod proto test
	@echo "Building server and client binaries..."
	export CGO_ENABLED=1
	if [ ! -d "$(BUILD_DIR)" ]; then mkdir -p $(BUILD_DIR); fi
	@echo "Modifying version..."
	sed -i 's|VERSION = .*|VERSION = "$(APP_VERSION)"|' $(CLIENT_SRC_DIR)/cmd/root.go
	@echo "Building server..."
	$(GOBUILD) -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) $(SERVER_SRC_DIR)/$(MAIN_GO_FILE)
	@echo "Building client..."
	$(GOBUILD) -o $(BUILD_DIR)/$(CLIENT_BINARY_NAME) $(CLIENT_SRC_DIR)/$(MAIN_GO_FILE)

test: 
	@echo "Running tests..."
	$(GOTEST) -v ./...

clean: 
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(CLIENT_BINARY_NAME)
	rm -f $(BUILD_DIR)/$(SERVER_BINARY_NAME)
	rm -f $(CLIENT_CONFIG_DIR)/$(CLIENT_CONFIG_FILE)
	sudo rm -f /usr/local/bin/$(CLIENT_BINARY_NAME)
	docker rmi $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest
	docker rmi $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

proto: 
	@echo "Generating protobuf files..."
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/*.proto

mod: 
	@echo "Installing dependencies..."
	$(GOGET) -v ./...
	@echo "Tidying up go.mod..."
	$(GOMOD) tidy

install: build
	@echo "Installing client binaries..."
	sudo cp $(BUILD_DIR)/$(CLIENT_BINARY_NAME) /usr/local/bin
	if [ ! -d "$(CLIENT_CONFIG_DIR)" ]; then mkdir -p $(CLIENT_CONFIG_DIR); fi
	if [ ! -f "$(CLIENT_CONFIG_DIR)/$(CLIENT_CONFIG_FILE)" ]; then cp $(CLIENT_CONFIG_FILE) $(CLIENT_CONFIG_DIR); fi
	@echo "Client installed successfully!"

uninstall: 
	@echo "Uninstalling client..."
	sudo rm -f /usr/local/bin/$(CLIENT_BINARY_NAME)

serve: build
	@echo "Running server..."
	$(BUILD_DIR)/$(SERVER_BINARY_NAME) serve

docker-build:
	@echo "Building docker image..."
	docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .
	@echo "Updating server-deployment.yaml with new image..."
	sed -i 's|image: $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):.*|image: $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)|' server-deployment.yaml

docker-push: docker-build
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

podman-build:
	@echo "Building podman image..."
	podman build -t docker.io/$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest -t docker.io/$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .
	@echo "Updating server-deployment.yaml with new image..."
	sed -i 's|image: $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):.*|image: $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)|' server-deployment.yaml

podman-push: podman-build
	podman push docker.io/$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest
	podman push docker.io/$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

.PHONY: all build clean test serve docker-build docker-push podman-build podman-push proto pre-reqs install uninstall mod