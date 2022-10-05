#!/usr/bin/env sh
set -ef

if [ -n "${DEBUG}" ]; then
  set -x
fi

_sudo() {
  [ "$(id -u)" -eq 0 ] || set -- command sudo "$@"
  "$@"
}

_detect_arch() {
  case $(uname -m) in
    amd64 | x86_64)
      echo "x86_64"
      ;;
    arm64 | aarch64)
      echo "arm64"
      ;;
    i386)
      echo "i386"
      ;;
    *)
      echo "Unsupported processor architecture"
      return 1
      ;;
  esac
}

_detect_os() {
  case $(uname) in
    Linux)
      echo "Linux"
      ;;
    Darwin)
      echo "macOS"
      ;;
    Windows)
      echo "Windows"
      ;;
  esac
}

_download_url() {
  local arch="$(_detect_arch)"
  local os="$(_detect_os)"
  local version=$kusk_VERSION

  if [ -z "$kusk_VERSION" ]; then
    version=$(curl -s https://api.github.com/repos/kubeshop/kusk-gateway/releases/latest 2> /dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  fi

  local trailedVersion=$(echo $version | tr -d v)
  echo "https://github.com/kubeshop/kusk-gateway/releases/download/${version}/kusk_${trailedVersion}_${os}_${arch}.tar.gz"
}

echo "Downloading Kusk from URL: $(_download_url)"
curl --progress-bar --output kusk.tar.gz -SLf "$(_download_url)"
tar -xzf kusk.tar.gz kusk
rm kusk.tar.gz

install_dir=$1
if [ "$install_dir" != "" ]; then
  mkdir -p "$install_dir"
  mv kusk "${install_dir}/kusk"
  echo "Kusk installed in ${install_dir}"
  exit 0
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "Sudo rights are needed to move the binary to /usr/local/bin, please type your password when asked"
  _sudo mv kusk /usr/local/bin/kusk
else
  mv kusk /usr/local/bin/kusk
fi

echo "Kusk installed in /usr/local/bin/kusk"
