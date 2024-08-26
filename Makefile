# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# Define variables
GOLANG          := golang:1.23
ALPINE          := alpine:3.20
LS_APP          := lift-sim
VERSION         := 0.0.1
BASE_IMAGE_NAME := localhost
LS_IMAGE        := $(BASE_IMAGE_NAME)/$(LS_APP):$(VERSION)

# Docker Compose settings
DOCKER_COMPOSE  := docker compose
DC_FILE         := deployments/docker-compose.yaml
ENV_FILE        := ./src/.env.development
BUILD_DATE      := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_REF       := $(VERSION)

# Determine the environment
ENV ?= development
ifeq ($(ENV),production)
    ENV_FILE := ./src/.env.production
else ifeq ($(ENV),ci)
    ENV_FILE := ./src/.env.ci
endif

# Load environment variables from .env file
include $(ENV_FILE)
export

# ==============================================================================
# Building containers

.PHONY: build
build: lift-simulation

.PHONY: lift-simulation
lift-simulation:
	docker build \
		-f deployments/docker/dockerfile.lift-simulation \
		-t $(LS_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg ENV=$(ENV) \
		.

# ==============================================================================
# Running and managing containers

.PHONY: up
up:
	export BASE_IMAGE_NAME=$(BASE_IMAGE_NAME) LS_APP=$(LS_APP) VERSION=$(VERSION) BUILD_DATE=$(BUILD_DATE) && \
	$(DOCKER_COMPOSE) -f $(DC_FILE) --env-file $(ENV_FILE) up

.PHONY: down
down:
	$(DOCKER_COMPOSE) -f $(DC_FILE) down

.PHONY: logs
logs:
	$(DOCKER_COMPOSE) -f $(DC_FILE) logs -f

# ==============================================================================
# Running the application

.PHONY: run
run:
	GO_ENV=$(ENV) go run ./src/cmd/api/main.go

# ==============================================================================
# Testing

.PHONY: test
test:
	GO_ENV=$(ENV) go test ./src/... -v

# ==============================================================================
# Cleaning up

.PHONY: clean
clean:
	$(DOCKER_COMPOSE) -f $(DC_FILE) down -v
	docker rmi $(LS_IMAGE)

# ==============================================================================
# Help

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build            : Build the lift-simulation Docker image"
	@echo "  lift-simulation  : Build only the lift-simulation Docker image"
	@echo "  up               : Build and start the containers using Docker Compose"
	@echo "  down             : Stop and remove the containers"
	@echo "  logs             : View container logs"
	@echo "  run              : Run the application locally"
	@echo "  test             : Run the Go tests"
	@echo "  clean            : Remove containers, volumes, and images"
	@echo ""
	@echo "Use ENV=<environment> to specify the environment (development, production, ci)"
	@echo "Example: make up ENV=production"

# ==============================================================================
# Default

.DEFAULT_GOAL := help
