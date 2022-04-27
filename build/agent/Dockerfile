# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
ARG GOPROXY
ENV GOPROXY=$GOPROXY
RUN go mod download

# Copy the go source
COPY . .

ARG ANALYTICS_API_KEY 
ARG ANALYTICS_TRACKING_ID 

# Build
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags "-X github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$TELEMETRY_TOKEN" -o agent cmd/agent/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/agent .
USER 65532:65532

ENTRYPOINT ["/agent"]
