#!/usr/bin/env sh
set -ef

docker build --no-cache \
  -t banaioltd/envoy-auth-basic-http-service:v0.0.1 \
  -t banaioltd/envoy-auth-basic-http-service:latest \
  -t quay.io/banaio/envoy-auth-basic-http-service:latest \
  -t quay.io/banaio/envoy-auth-basic-http-service:v0.0.1 \
  -t envoy-auth-basic-http-service:latest \
  --file ./http-service/Dockerfile .

docker push banaioltd/envoy-auth-basic-http-service:latest
