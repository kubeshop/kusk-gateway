#!/usr/bin/env bash

set -e

k3d registry delete reg

k3d cluster delete local-k8s
