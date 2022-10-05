# Install Kusk CLI

Kusk is a CLI tool designed to help you manage common tasks related to Kusk Gateway.

The CLI supports the following commands:

- `install` - Installs Kusk Gateway and all its components with a single command - [Read more](install-cmd.md).
- `deploy` - Deploys your API and configures Kusk Gateway using OpenAPI - [Read more](deploy-cmd.md).
- `ip` - Provides the IP address of the default Kusk LoadBalancer
- `generate` - Generates Kusk Gateway API resources from OpenAPI - [Read more](generate-cmd.md).
- `dashboard` - Opens a port-forward to access Kusk Dashboard  - [Read more](dashboard-cmd.md).

## **Installation**

### **Homebrew**

```
brew install kubeshop/kusk/kusk
```

### **Using Golang Installation**

```
go install github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```

To install a particular version replace `latest` with the version number.

You can get a list of the available Kusk Gateway versions from our [releases page](https://github.com/kubeshop/kusk-gateway/releases).

### **Install Script**
Install `kusk` into `/usr/local/bin/kusk`:

```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

### **From Source**
```
git clone git@github.com:kubeshop/kusk.git && \
cd kusk && \
go install
```

### **Alternative Installation Method - Manual**

If you prefer installing the CLI manually:

1. Download [the latest binary from Github](https://github.com/kubeshop/kusk-gateway/releases/).
2. Upack it (`tar -zxvf kusk_1.2.3_Linux_x86_64.tar.gz`).
3. Move it to a location in the PATH. For example `mv kusk_0.1.0_Linux_arm64/kusk /usr/local/bin/kusk`.

For Windows, unpack the binary and add it to `%PATH%`. 

## **Updating**
### **Homebrew**

```
brew upgrade kubeshop/kusk/kusk
```

### **Latest Release on GitHub**

```
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

### **Using go**

```
go install -x github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```

### **From Source**

Inside of the Kusk repository directory:

```
git clone https://github.com/kubeshop/kusk-gateway.git
cd cmd/kusk
go install
```
