# Image URL to use all building/pushing image targets
TAG ?= latest
IMG ?= vmwaresaas.jfrog.io/vdp/source/vdp-kafka-monitoring

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

ifdef CI_JOB_TOKEN
	GITLAB_TOKEN := $${CI_JOB_TOKEN}
else
	GITLAB_TOKEN := $(shell vdpctl git::gitlab-token)
endif
DOCKER_BUILD_ARGS := --build-arg GITLAB_TOKEN=$(GITLAB_TOKEN)

DEV_ENV_WORK_DIR := /go/src/${PKG}
CURRENT_DIR=$(shell pwd)
PROJECT_DIR := $(shell dirname $(abspath $(firstword $(MAKEFILE_LIST))))

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: test
fmt: ## Run go fmt against code.
	go test ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

##@ Build

.PHONY: build
build: tidy fmt vet  ## Build manager binary.
	go build -o bin/vdp-kafka-monitoring

.PHONY: run
run: fmt vet
	go run main.go producer

.PHONY: build-image
build-image: vendor ## Build docker image with the manager.
	docker build -t ${IMG}:${TAG} $(DOCKER_BUILD_ARGS) .

.PHONY: push-image
push-image: ## Push docker image with the manager.
	docker push ${IMG}:${TAG}


vendor: go.mod go.sum
	go mod vendor
