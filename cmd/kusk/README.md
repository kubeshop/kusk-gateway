# Kusk CLI

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=shields)](http://makeapullrequest.com)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kubeshop/kusk-gateway)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kubeshop/kusk-gateway)
![GitHub Release Date](https://img.shields.io/github/release-date/kubeshop/kusk-gateway)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/kubeshop/kusk-gateway/Go?label=Go)

----

Kusk is a CLI tool designed to help you manage common tasks required when running Kusk Gateway.

Currently we support the following commands:

- `install` - installs Kusk Gateway and all its components with a single command. (Requires a helm installation)
- `api generate` - for creating Kusk Gateway API resources from your OpenAPI specification document.

---

# Table of contents

- [Usage](#usage)
  - [install](#install)
  - [api generate](#api-generate)
  - [dashboard](#dashboard)
- [Installation](#installation)
- [Updating](#updating)
- [Uninstallation](#uninstallation)
- [Contributing](#contributing)
- [License](#license)

# Usage

## Install

### Flags

|         Flag         |                                                     Description                                                     | Required? |
| :------------------: | :-----------------------------------------------------------------------------------------------------------------: | :-------: |
|       `--name`       | the prefix of the name to give to the helm releases for each of the kusk gateway components (default: kusk-gateway) |     ❌     |
| `--namespace` / `-n` |  the namespace to install kusk gateway into. Will create the namespace if it doesn't exist (default: kusk-system)   |     ❌     |
|   `--no-dashboard`   |                               when set, will not install the kusk gateway dashboard.                                |     ❌     |
|      `--no-api`      |                      when set, will not install the kusk gateway api. implies --no-dashboard.                       |     ❌     |
|  `--no-envoy-fleet`  |                                     when set, will not install any envoy fleets                                     |     ❌     |

### Examples

```sh
$ kusk install

Will install kusk-gateway, a public (for your APIS) and private (for the kusk dashboard and api)
envoy-fleet, api, and dashboard in the kusk-system namespace using helm.

$ kusk install --name=my-release --namespace=my-namespace

Will create a helm release named with --name in the namespace specified by --namespace.

$ kusk install --no-dashboard --no-api --no-envoy-fleet

Will install kusk-gateway, but not the dashboard, api, or envoy-fleet.
```

## Api generate

Generate accepts your OpenAPI spec file as input either as a local file or a URL pointing to your file
and generates a Kusk Gateway compatible API resource that you can apply directly into your cluster.

Configuration of the API resource is done via the x-kusk extension.

If the OpenAPI spec doesn't have a top-level x-kusk annotation set, it will add them for you and set
the upstream service, namespace and port to the flag values passed in respectively and set the rest of the settings to defaults.
This is enough to get you started

If the x-kusk extension is already present, it will override the the upstream service, namespace and port to the flag values passed in respectively
and leave the rest of the settings as they are.

You must specify the name of the envoyfleet you wish to use to expose your API. This is because Kusk Gateway could be managing more than one.
In the future, we will add the notion of a default envoyfleet which kusk gateway will use when none is specified. i.e. kusk-gateway-envoy-fleet.

If you do not specify the envoyfleet namespace, it will default to kusk-system.

Sample usage

No name specified

```sh
kusk api generate \
  -i spec.yaml \
  --envoyfleet.name kusk-gateway-envoy-fleet \
  --envoyfleet.namespace kusk-system
```

In the above example, kusk will use the openapi spec info.title to generate a manifest name and leave the existing
x-kusk extension settings

No api namespace specified

```sh
kusk api generate \
  -i spec.yaml \
  --name httpbin-api \
  --upstream.service httpbin \
  --upstream.port 8080 \
  --envoyfleet.name kusk-gateway-envoy-fleet
```

In the above example, as --namespace isn't defined, it will assume the default namespace.

Namespace specified

```sh
kusk api generate \
  -i spec.yaml \
  --name httpbin-api \
  --upstream.service httpbin \
  --upstream.namespace my-namespace \
  --upstream.port 8080 \
  --envoyfleet.name kusk-gateway-envoy-fleet
```

OpenAPI spec at URL

```sh
kusk api generate \
    -i https://raw.githubusercontent.com/$ORG_OR_USER/$REPO/myspec.yaml \
    --name httpbin-api \
    --upstream.service httpbin \
    --upstream.namespace my-namespace \
    --upstream.port 8080 \
    --envoyfleet.name kusk-gateway-envoy-fleet
```

This will fetch the OpenAPI document from the provided URL and generate a Kusk Gateway API resource

### Flags

|          Flag          |                                             Description                                             | Required? |
| :--------------------: | :-------------------------------------------------------------------------------------------------: | :-------: |
|        `--name`        | the name to give the API resource e.g. --name my-api. Otherwise taken from OpenAPI info title field |     ❌     |
|  `--namespace` / `-n`  | the namespace of the API resource e.g. --namespace my-namespace, -n my-namespace (default: default) |     ❌     |
|     `--in` / `-i`      |       file path or URL to OpenAPI spec file to generate mappings from. e.g. --in apispec.yaml       |     ✅     |
|  `--upstream.service`  |                                 name of upstream Kubernetes service                                 |     ❌     |
| `--upstream.namespace` |                          namespace of upstream service (default: default)                           |     ❌     |
|   `--upstream.port`    |                       port that upstream service is exposed on (default: 80)                        |     ❌     |
|  `--envoyfleet.name`   |                               name of envoyfleet to use for this API                                |     ✅     |
| `envoyfleet.namespace` |                  namespace of envoyfleet to use for this API. Default: kusk-system                  |     ❌     |

### Example

Take a look at the [`./examples/httpbin-spec.yaml`](./examples/httpbin-spec.yaml)

```sh
kusk api generate -i ./examples/httpbin-spec.yaml --name httpbin-api --upstream.service httpbin --upstream.port 8080 --envoyfleet.name kusk-gateway-envoy-fleet
```

The output should contain the following x-kusk extension at the top level

```sh
...
x-kusk:
  cors: {}
  path:
    rewrite:
      pattern: ""
      substitution: ""
  upstream:
    service:
	name: httpbin
	namespace: default
	port: 8080
```

## dashboard
Access the kusk dashboard. kusk dashboard will start a port-forward session on port 8080 to the envoyfleet
serving the dashboard and will open the dashboard in the browser. By default this is kusk-gateway-private-envoy-fleet.kusk-system.

The flags --envoyfleet.namespace and --envoyfleet.name can be used to change the envoyfleet.

### Flags
|           Flag           |                                         Description                                          | Required? |
| :----------------------: | :------------------------------------------------------------------------------------------: | :-------: |
|      `--kubeconfig`      |                                 absolute path to kube config                                 |     ❌     |
|   `--envoyfleet.name`    | kusk gateway dashboard envoy fleet service name. (default: kusk-gateway-private-envoy-fleet) |     ❌     |
| `--envoyfleet.namespace` |         kusk gateway dashboard envoy fleet service namespace. (default: kusk-system)         |     ❌     |
|    `--external-port`     |                    external port to access dashboard at. (default: 8080)                     |     ❌     |

### Examples

```sh
$ kusk dashboard
```

Opens the kusk gateway dashboard in the browser by exposing the default private envoy fleet on port 8080

```sh
$ kusk dashboard --envoyfleet.namespace=other-namespace --envoyfleet.name=other-envoy-fleet
```

Specify other envoyfleet and namespace that is serving the dashboard

```sh
$ kusk dashboard --external-port=9090
```

Expose dashboard on port 9090

```sh
$ kusk dashboard --kubeconfig=/path/to/kube/config
```

Specify path to kube config. $HOME/.kube/config is used by default.

# Installation

## Homebrew

```sh
brew install kubeshop/kusk/kusk
```

## Go install the latest release on Github

```sh
go install github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```

To install a particular version: replace `latest` with the version number

You can get a list of the available kusk versions from our [releases page](https://github.com/kubeshop/kusk-gateway/releases)

## Easy install script

This will install `kusk` into `/usr/local/bin/kusk`

```sh
$ curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | sh
...
kusk installed in /usr/local/bin/kusk
```

## From source

```sh
git clone git@github.com:kubeshop/kusk-gateway.git && \
  cd kusk-gateway/cmd/kusk && \
  go install
```

## Alternative installation method (manual)

If you don't like automatic scripts you can always use the manual install:

1. Download binary with version of your choice (recent one is recommended).
2. Unpack it (`tar -zxvf kusk_0.1.0_Linux_arm64.tar.gz`).
3. Move it to a location in the `PATH`. For example `mv kusk_0.1.0_Linux_arm64/kusk /usr/local/bin/kusk`.

For Windows, download the binary from [here](https://github.com/kubeshop/kusk-gateway/releases), unpack the binary and add it to `%PATH%`.

# Updating

## Homebrew

```sh
brew upgrade kubeshop/kusk/kusk
```

## Latest release on Github

```sh
go install github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```

## From source

Inside of the kusk repository directory

```sh
git pull && go install
```

# CLI Reference 

For detailed command line reference visit [docs](docs/kusk.md)
# Contributing

Your contributions are always welcome! Please have a look at [How to contribute](https://github.com/kubeshop/.github/blob/main/CONTRIBUTING.md) first.

### Development

```sh
make
./kusk --help
```

# License

[The MIT License](https://mit-license.org/).

Please have a look at the [`./LICENSE`](./LICENSE) for more details.
