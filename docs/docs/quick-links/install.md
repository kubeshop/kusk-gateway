# Install 

Kusk CLI is available for macOS, Linux and Windows. It is used to configure Kusk Gateway, which is Kusk's Ingress Controller for Kubernetes.  

You'll also need:
- Docker
- A Kubernetes cluster

:::note
If you are missing an installation method you'd like us to support, please [open an issue in the Github repository](https://github.com/kubeshop/kusk-gateway/issues/new?assignees=&labels=kind%2Ffeature&template=feature_request.md&title=) and we will try to address it right away!
:::

## MacOS

```sh
brew install kubeshop/kusk/kusk
```

If you don't use `brew`, you can also download Kusk CLI with: 

```sh 
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

## Linux 

```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

_Support for APT is coming soon._

## Windows

For Windows installation you can either download the [latest release binary](https://github.com/kubeshop/kusk-gateway/releases/latest) or use the following command ([`go` binary](https://go.dev/doc/install)  needed):

```sh 
go install -x github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```

_Support for Chocolatey coming soon._



