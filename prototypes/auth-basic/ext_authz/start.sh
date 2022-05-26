#!/usr/bin/env sh
set -ef

export FRONT_ENVOY_YAML=config/http-service.yaml
docker-compose up --build --abort-on-container-exit
