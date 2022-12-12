# syntax = docker/dockerfile:experimental
FROM --platform=$BUILDPLATFORM docker.io/golang:1.18 as builder

WORKDIR /go/src

COPY . .

ARG TELEMETRY_TOKEN
ARG VERSION
ARG TARGETARCH
ARG TARGETOS
# https://skaffold.dev/docs/workflows/debug/#go-runtime-go
ARG SKAFFOLD_GO_GCFLAGS

RUN \
  --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
  go build -mod=vendor -v -ldflags "-X github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$TELEMETRY_TOKEN -X github.com/kubeshop/kusk-gateway/pkg/build.Version=$VERSION" -o kusk-gateway-api cmd/api-server/main.go

FROM --platform=$BUILDPLATFORM gcr.io/distroless/static:nonroot

ENV GOTRACEBACK=single

COPY --from=builder --chown=65532:65532 /go/src/kusk-gateway-api ./
EXPOSE 8080/tcp

USER 65532:65532

ENTRYPOINT ["./kusk-gateway-api"]
