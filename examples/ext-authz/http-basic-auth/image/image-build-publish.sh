#!/usr/bin/env sh
set -euf

docker build --no-cache \
  -t kubeshop/ext-authz-http-basic-auth:v0.1.0 \
  -t kubeshop/ext-authz-http-basic-auth:latest \
  --file ./Dockerfile \
  .

docker push docker.io/kubeshop/ext-authz-http-basic-auth:latest
docker push docker.io/kubeshop/ext-authz-http-basic-auth:v0.1.0
