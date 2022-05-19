#!/usr/bin/env sh
set -ef

if [ -n "${DEBUG}" ]; then
  set -x
fi

KIND_VERSION="${KIND_VERSION:=v0.13.0}"
KUBECTL_VERSION="${KUBECTL_VERSION:=v1.24.0}"
CLUSTER_NAME="${CLUSTER_NAME:=kgw}"
IMAGE="${IMAGE:=kindest/node:v1.22.0@sha256:b8bda84bb3a190e6e028b1760d277454a72267a5454b57db34437c34a588d047}"
# IMAGE="${IMAGE:=kindest/node:v1.16.4@sha256:b91a2c2317a000f3a783489dfb755064177dbc3a0b2f4147d50f04825d016f55}"

if which go >/dev/null 2>&1; then
  echo "go version=$(go version)"
else
  echo "go missing"
  exit 1
fi

echo

# https://kind.sigs.k8s.io/docs/user/quick-start/
if which kind >/dev/null 2>&1; then
  echo "kind version=$(kind version)"
else
  go install "sigs.k8s.io/kind@${KIND_VERSION}"
  echo "kind version=$(kind version)"
fi

echo

if which kubectl >/dev/null 2>&1; then
  echo "kubectl version=$(kubectl version --client --short)"
else
  echo "creating ${HOME}/.bin/"
  mkdir -pv ~/.bin/

  echo

  if [ "$(uname)" = 'Linux' ]; then
    curl --progress-bar --output "${HOME}/.bin/kubectl" --location "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
  fi
  if [ "$(uname)" = 'Darwin' ]; then
    curl --progress-bar --output "${HOME}/.bin/kubectl" --location "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/darwin/amd64/kubectl"
  fi
  chmod +x "${HOME}/.bin/kubectl"

  echo

  # shellcheck disable=SC2016
  echo 'add "${HOME}/.bin" to path, i.e., "export PATH="${HOME}/.bin:${PATH}"'
  echo "or"
  # shellcheck disable=SC2016
  echo '"export PATH="${HOME}/.bin:${PATH}" >> ~/.profile'

  echo
  export PATH="${HOME}/.bin:${PATH}"
  echo "kubectl version=$(kubectl version --client --short)"
fi

echo

CONTAINER_ENGINE=""
if which podman >/dev/null 2>&1; then
  export KIND_EXPERIMENTAL_PROVIDER=podman
  CONTAINER_ENGINE="$(which podman)"
  echo "podman version=$(podman version)"
  echo "exporting 'KIND_EXPERIMENTAL_PROVIDER=podman'"
fi

echo

if which docker >/dev/null 2>&1; then
  export DOCKER_CLI_EXPERIMENTAL=enabled
  CONTAINER_ENGINE="$(which docker)"
  echo "docker version=$(kind version)"
  echo "exporting 'DOCKER_CLI_EXPERIMENTAL=enabled'"
fi

echo
echo "engine=${CONTAINER_ENGINE}"
echo

while kind get clusters | grep "${CLUSTER_NAME}"; do
  echo
  echo "deleting ${CLUSTER_NAME}"
  kind delete cluster --name "${CLUSTER_NAME}"
  sleep 2s
done

TEMP_DIR="$(mktemp -d)"
cat <<EOF >"${TEMP_DIR}/config-kind.yaml"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    image: ${IMAGE}
  - role: worker
    image: ${IMAGE}
EOF

echo
cat "${TEMP_DIR}/config-kind.yaml"
echo

# kind create cluster --name "${CLUSTER_NAME}" --config "${TEMP_DIR}/config-kind.yaml" --image 'kindest/node:v1.20.0' --wait 8m
kind create cluster --name "${CLUSTER_NAME}" --config "${TEMP_DIR}/config-kind.yaml" --image "${IMAGE}" --wait 8m

echo
echo "testing cluster context=kind-${CLUSTER_NAME}"
echo

set -x
kubectl cluster-info --context "kind-${CLUSTER_NAME}"
kubectl run busybox -it --rm --image=busybox --restart=Never -- /bin/sh -c 'echo "running in kind - hostname=$(hostname)"'
