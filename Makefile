# Image URL to use all building/pushing image targets
TAG ?= latest
IMG ?= vmware/kafka-slo-exporter

CGO_ENABLED ?= 0
EXTRA_GO_LDFLAGS ?= ""
BUILDARCH ?= amd64
GO_LDFLAGS := -s -w $(EXTRA_GO_LDFLAGS)

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
	go build -o bin/kafka-slo-monitoring 

.PHONY: run-consumer
run-consumer: fmt vet
	go run main.go consumer

.PHONY: run-producer
run-producer: fmt vet
	go run main.go producer

.PHONY: run-app
run-app: fmt vet
	go run main.go app

.PHONY: build-image
build-image: vendor ## Build docker image with the manager.
	docker build -t ${IMG}:${TAG} $(DOCKER_BUILD_ARGS) .

.PHONY: push-image
push-image: build-image ## Push docker image with the manager.
	docker push ${IMG}:${TAG}
	docker push ${IMG}:latest

.PHONY: start-environ
start-environ:
	 docker-compose -f compose.yaml up

.PHONY: godoc
godoc: ## Use gomarkdoc to generate documentation for the whole project
	gomarkdoc --output 'docs/{{.Dir}}/README.md' ./cmd/...
	gomarkdoc --output 'docs/{{.Dir}}/README.md' ./config/...
	gomarkdoc --output 'docs/{{.Dir}}/README.md' ./pkg/...