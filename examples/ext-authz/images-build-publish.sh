#!/usr/bin/env bash
set -euf

BASIC_AUTH_IMAGE_NAME="kusk-ext-authz-http-service-basic-auth"
BEARER_TOKEN_IMAGE_NAME="kusk-ext-authz-http-service-bearer-token"

docker build --no-cache \
  -t kubeshop/${BASIC_AUTH_IMAGE_NAME}:v0.0.1 \
  -t kubeshop/${BASIC_AUTH_IMAGE_NAME}:latest \
  -t ${BASIC_AUTH_IMAGE_NAME}:latest \
  -t ${BASIC_AUTH_IMAGE_NAME}:v0.0.1 \
  --file ./basic-auth/http-service/Dockerfile \
  .

docker build --no-cache \
  -t kubeshop/${BEARER_TOKEN_IMAGE_NAME}:v0.0.1 \
  -t kubeshop/${BEARER_TOKEN_IMAGE_NAME}:latest \
  -t ${BEARER_TOKEN_IMAGE_NAME}:latest \
  -t ${BEARER_TOKEN_IMAGE_NAME}:v0.0.1 \
  --file ./bearer-token/http-service/Dockerfile \
  .

docker push docker.io/kubeshop/${BASIC_AUTH_IMAGE_NAME}:latest
docker push docker.io/kubeshop/${BASIC_AUTH_IMAGE_NAME}:v0.0.1

docker push docker.io/kubeshop/${BEARER_TOKEN_IMAGE_NAME}:latest
docker push docker.io/kubeshop/${BEARER_TOKEN_IMAGE_NAME}:v0.0.1
