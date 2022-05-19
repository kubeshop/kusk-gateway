#!/usr/bin/env bash

kind create cluster --image kindest/node:v1.24.0 --wait 5m --name kusk --config ./cluster.yaml
