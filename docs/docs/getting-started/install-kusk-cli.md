# 1. Install Kusk CLI 

To install Kusk CLI, you will need the following tools available in your system:

- [kubectl](https://kubernetes.io/docs/tasks/tools/) 
- [docker](https://docs.docker.com/desktop/)

**MacOS**
```sh
brew install kubeshop/kusk/kusk
```

**Linux**

```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

**Windows (Chocolatey)**

```sh
choco source add --name=kubeshop_repo --source=https://chocolatey.kubeshop.io/chocolatey
choco install kusk -y
```

:::note

For **other installation methods**, check the [extended installation page](../quick-links/install.md).

:::