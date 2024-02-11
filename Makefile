# Makefile

# Project-specific variables
BINARY_NAME=urlshortener
DOCKER_IMAGE_NAME=urlshortener
DOCKER_TAG=latest

# Go-related variables
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test -v
GO_CLEAN=$(GO_CMD) clean

# Docker-related variables
DOCKER_CMD=docker
DOCKER_BUILD=$(DOCKER_CMD) build
DOCKER_PUSH=$(DOCKER_CMD) push

# Build the Go binary
.PHONY: build
build:
	$(GO_BUILD) -o $(BINARY_NAME) .

# Run tests
.PHONY: test
test:
	$(GO_TEST) ./...

# Clean up
.PHONY: clean
clean:
	$(GO_CLEAN)
	rm -f $(BINARY_NAME)
	rm -fr ./result

# Build the Docker image
.PHONY: docker
docker:
	$(DOCKER_BUILD) -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

# Build the Docker image
.PHONY: nix-build
nix-build:
	nix-build default.nix

# Build the Docker image
.PHONY: nix-docker
nix-docker:
	$(DOCKER_BUILD) -f NixDockerfile -t nix$(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .
