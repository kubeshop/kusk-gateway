#!/usr/bin/env sh
set -euf

docker build --no-cache \
  -t kubeshop/kusk-ext-authz-http-service:v0.0.1 \
  -t kubeshop/kusk-ext-authz-http-service:latest \
  -t kusk-ext-authz-http-service:latest \
  -t kusk-ext-authz-http-service:v0.0.1 \
  --file ./http-service/Dockerfile .

docker push docker.io/kubeshop/kusk-ext-authz-http-service:latest
