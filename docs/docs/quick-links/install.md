# Install Kusk CLI

Kusk CLI is available for macOS, Linux and Windows. 

The Kusk CLI is used to configure Kusk Gateway, the Ingress Controller for Kubernetes.

**System requirements:**
- [kubectl](https://kubernetes.io/docs/tasks/tools/) 
- [docker](https://docs.docker.com/desktop/)

## 1. Install Kusk CLI

:::note
If you are missing an installation method you'd like us to support, please [open an issue in the Github repository](https://github.com/kubeshop/kusk-gateway/issues/new?assignees=&labels=kind%2Ffeature&template=feature_request.md&title=) and we will try to address it right away!
:::

### MacOS

**Using `brew`**:

```sh
brew install kubeshop/kusk/kusk
```

**Using install script:**

```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

### Linux

**Using install script:**

```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

**Using APT (Debian/Ubuntu):**

```sh
wget -qO - https://repo.kubeshop.io/key.pub | sudo apt-key add -
echo "deb https://repo.kubeshop.io/kusk linux main" | sudo tee -a /etc/apt/sources.list
sudo apt-get update
sudo apt-get install -y kusk
```

### Windows

```sh
choco source add --name=kubeshop_repo --source=https://chocolatey.kubeshop.io/chocolatey
choco install kusk -y
```

### Other installation methods

For other ways of installation, you can download the [latest release binary](https://github.com/kubeshop/kusk-gateway/releases/latest) or use the following command ([`go` binary](https://go.dev/doc/install)  needed):

```sh
go install -x github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```