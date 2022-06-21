# Build the manager binary
FROM docker.io/golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
ARG GOPROXY
ENV GOPROXY=$GOPROXY
RUN go mod download

ARG TELEMETRY_TOKEN

# Copy the go source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags "-X github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$TELEMETRY_TOKEN" -o manager cmd/manager/main.go

# Directory for the files created and used by the manager, to be copied for static rootless images since we don't have shell to create it there
RUN mkdir -m=00755 /opt/manager

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder --chown=65532:65532 /opt/manager /opt/manager

USER 65532:65532

ENTRYPOINT ["/manager"]
