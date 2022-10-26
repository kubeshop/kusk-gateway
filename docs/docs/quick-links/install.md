# Install

Kusk CLI is available for macOS, Linux and Windows. It is used to configure Kusk Gateway, which is Kusk's Ingress Controller for Kubernetes.

You'll also need:
- Docker
- A Kubernetes cluster

## 1. Install Kusk CLI

:::note
If you are missing an installation method you'd like us to support, please [open an issue in the Github repository](https://github.com/kubeshop/kusk-gateway/issues/new?assignees=&labels=kind%2Ffeature&template=feature_request.md&title=) and we will try to address it right away!
:::

#### MacOS

```sh
brew install kubeshop/kusk/kusk
```

If you don't use `brew`, you can also download Kusk CLI with:

```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

#### Linux
Install Script
```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

APT (Debian/Ubuntu)
1. Download our public GPG key, and add it to the trusted keys:
```sh
wget -qO - https://repo.kubeshop.io/key.pub | sudo apt-key add -
```

2. Add our repository to your apt sources:
```sh
echo "deb https://repo.kubeshop.io/kusk linux main" | sudo tee -a /etc/apt/sources.list
```

3. Update your local package index:
```sh
sudo apt-get update
```

4: Install Kusk
```sh
sudo apt-get install -y kusk
```

#### Windows
Install with Chocolatey

Please run the commands from an elevated command shell.

1. Add our repository URL to the list of Chocolatey sources:
```sh
choco source add --name=kubeshop_repo --source=https://chocolatey.kubeshop.io/chocolatey
```

2. Install `kusk`:
```sh
choco install kusk -y
```

For other ways of installation, you can download the [latest release binary](https://github.com/kubeshop/kusk-gateway/releases/latest) or use the following command ([`go` binary](https://go.dev/doc/install)  needed):

```sh
go install -x github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```

_Support for Chocolatey coming soon._

## 2. Install Kusk Gateway in your cluster

Once you've installed Kusk CLI, you can now install Kusk in your cluster:

```sh
kusk cluster install
```

This will install Kusk Gateway and all its components including Kusk Dashboard.

You can also install Kusk Gateway in your cluster [using `helm`](https://docs.kusk.io/quick-links/helm-install).
