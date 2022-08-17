# Installing Kusk Gateway
## **1. Install Kusk CLI** 

To install Kusk CLI, you will need the following tools available in your terminal:

- [helm](https://helm.sh/docs/intro/install/) command-line tool
- [kubectl](https://kubernetes.io/docs/tasks/tools/) command-line tool

```sh
# MacOS 
brew install kubeshop/kusk/kusk

# Linux
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

## **2. Install Kusk Gateway**

Use the Kusk CLIs [install command](../cli/install-cmd.md) to install Kusk Gateway components in your cluster. 

```sh
kusk install
```