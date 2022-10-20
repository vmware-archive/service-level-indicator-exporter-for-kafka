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
test: ## Run go test against code.
	go test ./... -count=1 -v

.PHONY: test-ci
test-ci: ## Run go test against code.
	go test --tags=ci ./... -count=1 -v

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: vendor
vendor: go.mod go.sum
	go mod vendor

##@ Build

.PHONY: build
build: tidy fmt vet  ## Build manager binary.
	go build -o bin/vdp-kafka-monitoring

.PHONY: run-consumer
run-consumer: fmt vet
	go run main.go consumer

.PHONY: run-producer
run-producer: fmt vet
	go run main.go producer

.PHONY: build-image
build-image: vendor ## Build docker image with the manager.
	docker build -t ${IMG}:${TAG} $(DOCKER_BUILD_ARGS) .

.PHONY: push-image
push-image: build-image ## Push docker image with the manager.
	docker push ${IMG}:${TAG}

.PHONY: start-environ
start-environ:
	 docker-compose -f compose.yaml up

.PHONY: kind-test
kind-test:
	vdpctl kind::create::gitlab
	docker pull vmwaresaas.jfrog.io/vdp/source/confluentinc/cp-kafka:7.0.1
	docker pull vmwaresaas.jfrog.io/vdp/source/confluentinc/cp-zookeeper:7.0.1
	kind load docker-image --name vdp-$(hostname | md5sum | cut -c1-6) vmwaresaas.jfrog.io/vdp/source/confluentinc/cp-kafka:7.0.1
	kind load docker-image --name vdp-$(hostname | md5sum | cut -c1-6) vmwaresaas.jfrog.io/vdp/source/confluentinc/cp-zookeeper:7.0.1
	kubectl apply -f resources/kafka-slim.yaml
	echo "Sleeping 60 seconds to start environ for e2e tests"
	sleep 60
	kubectl describe po kafka-0
	kubectl port-forward kafka-0 9092:9092 &
	make test-ci
	vdpctl kind::delete::gitlab
