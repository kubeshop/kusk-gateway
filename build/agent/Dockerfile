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

# Copy the go source
COPY . .

ARG GA_ID
ARG GA_SECRET

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags -ldflags "-X github.com/kubeshop/kusk-gateway/pkg/analytics.KuskGAMeasurementID=$GA_ID -X github.com/kubeshop/kusk-gateway/pkg/analytics.KuskGAApiSecret=$GA_SECRET" -o agent cmd/agent/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/agent .
USER 65532:65532

ENTRYPOINT ["/agent"]
