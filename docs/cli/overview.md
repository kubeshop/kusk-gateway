# Kusk CLI

Kusk is a CLI tool designed to help you manage common tasks related to Kusk Gateway.

The CLI supports the following commands:

- `install` - Installs Kusk Gateway and all its components with a single command - [Read more](install-cmd.md).  
- `api generate` - Generates Kusk Gateway API resources from OpenAPI - [Read more](generate-cmd.md).

## **Installation**

### **Homebrew**

```
brew install kubeshop/kusk/kusk
```

### **Using Golang Installation**

```
go install github.com/kubeshop/kusk@latest
```

To install a particular version replace `latest` with the version number.

You can get a list of the available Kusk Gateway versions from our [releases page](https://github.com/kubeshop/kusk/releases).

### **Install Script**
Install `kusk` into `/usr/local/bin/kusk`:

```sh
bash < <(curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk/main/scripts/install.sh)
```

### **From Source**
```
git clone git@github.com:kubeshop/kusk.git && \
cd kusk && \
go install
```

### **Alternative Installation Method - Manual**

If you don't like automatic scripts you can install the CLI manually:

1. Download binary with version of your choice (recent one is recommended).
2. Upack it (tar -zxvf kusk_0.1.0_Linux_arm64.tar.gz).
3. Move it to a location in the PATH. For example `mv kusk_0.1.0_Linux_arm64/kusk /usr/local/bin/kusk`.

For Windows, download the binary from [here](https://github.com/kubeshop/kusk/releases), unpack the binary and add it to `%PATH%`. 

## **Updating**
### **Homebrew**

```
brew upgrade kubeshop/kusk/kusk
```

### **Latest Release on GitHub**

```
go install github.com/kubeshop/kusk@$VERSION
```

### **From Source**

Inside of the kusk repository directory:

```
git clone https://github.com/kubeshop/kusk.git
```
