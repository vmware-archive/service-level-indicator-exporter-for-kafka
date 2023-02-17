# Build the manager binary
FROM golang:alpine as builder

ARG BUILDARCH
ARG CGO_ENABLED
ARG GO_LDFLAGS

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
COPY config.yaml config.yaml
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer

# Copy the go source
COPY cmd/ cmd/
COPY config/ config/
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=$CGO_ENABLED GOOS=linux GOARCH=$BUILDARCH && go build -ldflags $GO_LDFLAGS -a -o kafka-slo-monitoring 
RUN echo 'nonroot:x:1000:2000::/home/nonroot:/dev/null' > /tmp/passwd

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM scratch
WORKDIR /
COPY --from=builder /tmp/passwd /etc/passwd
COPY --from=builder /workspace/kafka-slo-monitoring /kafka-slo-monitoring
COPY --from=builder /workspace/config.yaml /config.yaml

USER 1000

ENTRYPOINT ["/kafka-slo-monitoring","app"]
