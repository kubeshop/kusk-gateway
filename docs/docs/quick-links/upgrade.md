# Upgrade

To upgrade Kusk Gateway in your cluster, you will need to install the new version of Kusk CLI first and then upgrade Kusk Gateway with it.

Check the [Helm upgrade guide](./helm-install.md) in case you are using Helm.

## 1. Update Kusk CLI

#### MacOS

```sh
brew upgrade kubeshop/kusk/kusk
```

or if you don't use `brew` you can upgrade directly with:

```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

#### Linux
Install script
```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

APT (Debian/Ubuntu)
1. Update your local repository index:
```sh
sudo apt-get update
```

2. Upgrade Kusk:
```sh
sudo apt-get upgrade -y kusk
```

#### Windows
Chocolatey

Please run the command from an elevated command shell.
```sh
choco upgrade kusk -y
```
For other ways of installation, you can download the [latest release binary](https://github.com/kubeshop/kusk-gateway/releases/latest) or use the following command ([`go` binary](https://go.dev/doc/install)  needed):

```sh
go install -x github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```

## 2. Update Kusk Gateway in your cluster

```sh
kusk cluster upgrade
```

This command will update Kusk Gateway and its components so you can use the latest features and have all the security measures up to date.
