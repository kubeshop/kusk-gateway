# syntax = docker/dockerfile:experimental
FROM --platform=$BUILDPLATFORM docker.io/golang:1.19 as builder

WORKDIR /go/src

COPY . .

ARG TARGETARCH
ARG TARGETOS
ARG TELEMETRY_TOKEN
ARG VERSION
# https://skaffold.dev/docs/workflows/debug/#go-runtime-go
ARG SKAFFOLD_GO_GCFLAGS

# Build
RUN \
  --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
  go build -mod=vendor -v -ldflags "-X 'github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$TELEMETRY_TOKEN' -X 'github.com/kubeshop/kusk-gateway/pkg/build.Version=$VERSION'" -gcflags="${SKAFFOLD_GO_GCFLAGS}" -o main cmd/manager/main.go

# Directory for the files created and used by the manager, to be copied for static rootless images since we don't have shell to create it there
RUN mkdir -m 00755 /opt/manager

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM --platform=$BUILDPLATFORM gcr.io/distroless/static:nonroot

ENV GOTRACEBACK=single

WORKDIR /

COPY --from=builder /go/src/main .
COPY --from=builder --chown=65532:65532 /opt/manager /opt/manager

USER 65532:65532

ENTRYPOINT ["/main"]
