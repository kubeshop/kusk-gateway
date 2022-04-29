<!-- Add banner here -->

# Kusk Gateway

<!-- Add buttons here -->

Kusk Gateway is a self-service API gateway powered by [OpenAPI](https://www.openapis.org/) and [Envoy](https://www.envoyproxy.io/).

Kusk Gateway is for you if:

- You or your team develop REST APIs running in Kubernetes
- You embrace a contract-first approach to developing your APIs using OpenAPI or Swagger
- You want to ramp-up time when deploying a new REST api to a cluster and you don't want to spend lots of time configuring ingress controllers that require a dedicated Ops Engineer
- You want your REST API endpoints traffic to be observable and controllable with the easy to use settings

Kusk Gateway has a unique way to the configuration among other API gateways as it configures itself through the metadata defined in your OpenAPI or Swagger document.

You can apply your API definition like any other Kubernetes resource using our custom-made Kusk API CustomResourceDefinition.

# Table of contents

- [Kusk Gateway](#kusk-gateway)
- [Table of contents](#table-of-contents)
- [Get Started](#get-started)
  - [Installation](#installation)
  - [Usage](#usage)
    - [API CRD Example](#api-crd-example)
  - [Custom Resources](#custom-resources)
  - [Roadmap](#roadmap)
  - [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Contribute](#contribute)
- [License](#license)

# Get Started

See the [architecture document](docs/reference/architecture.md) for an overview of the Kusk Gateway architecture

## Installation

[(Back to top)](#table-of-contents)

Kusk Gateway can be installed on any cloud or bare metal Kubernetes cluster.

If you want to quickly setup and evaluate the Kusk Gateway, then please use the [installation instructions](docs/getting-started/installation.md).

For the quick and impatient:

```sh

# Install Kubeshop Helm repo and update it
helm repo add kubeshop https://kubeshop.github.io/helm-charts && helm repo update

# Install the Kusk Gateway with CRDs into kusk-system namespace.
# We need to wait for the kusk-gateway-manager deployment to finish the setup for the next step.
helm install kusk-gateway kubeshop/kusk-gateway -n kusk-system --create-namespace &&\
kubectl wait --for=condition=available --timeout=600s deployment/kusk-gateway-manager -n kusk-system

# Install the "default" EnvoyFleet Custom Resource, which will be used by the Kusk Gateway
# to create Envoy Fleet Deployment and Service with the type LoadBalancer
helm install kusk-gateway-envoyfleet-default kubeshop/kusk-gateway-envoyfleet -n kusk-system

```

## Usage

[(Back to top)](#table-of-contents)

Kusk Gateway configures itself via the [API CRD](docs/reference/customresources/api.md) that contains your embedded Swagger or OpenAPI document.

See also [x-kusk extension documentation](docs/reference/extension.md) and [Custom Resources](docs/reference/customresources/index.md) for the guidelines on how to add the necessary routing information to your OpenAPI file and Kusk Gateway.

After that all that's required is to apply it as you would any other Kubernetes resource.

### API CRD Example

```yaml
apiVersion: gateway.kusk.io/v1
kind: API
metadata:
  name: httpbin-sample
spec:
  # service name, namespace and port should be specified inside x-kusk annotation
  spec: |
    swagger: '2.0'
    info:
      title: httpbin.org
      description: API Management facade for a very handy and free online HTTP tool.
      version: '1.0'
    x-kusk:
      validation:
        request:
          enabled: true # enable automatic request validation using OpenAPI spec
      upstream:
        service:
          name: httpbin
          namespace: default
          port: 8080
      path:
        # allows to serve under /api prefix
        prefix: "/api"
        # removes prefix when sending to upstream service
        rewrite:
          pattern: "^/api"
          substitution: ""
    paths:
      "/get":
          get:
            description: Returns GET data.
            operationId: "/get"
            responses: {}
      "/delay/{seconds}":
        get:
          description: Delays responding for n–10 seconds.
          operationId: "/delay"
          parameters:
          - name: seconds
            in: path
            description: ''
            required: true
            type: string
            default: 2
            enum:
            - 2
          responses: {}
      ...
```

## Custom Resources

[(Back to top)](#table-of-contents)

See [Custom Resources](https://kubeshop.github.io/kusk-gateway/reference/customresources/) for more information on the Custom Resources that Kusk Gateway supports.

## Roadmap

[(Back to top)](#table-of-contents)

For the list of the currently supported and planned features please check the [Roadmap](https://kubeshop.github.io/kusk-gateway/contributing/roadmap/).

## Troubleshooting

[(Back to top)](#table-of-contents)

See the [Troubleshooting](https://kubeshop.github.io/kusk-gateway/troubleshooting/) for how to troubleshoot the Kusk Gateway problems.

# Development

[(Back to top)](#table-of-contents)

See our [Development document](https://kubeshop.github.io/kusk-gateway/contributing/development/) for how to develop Kusk Gateway.

# Contribute

[(Back to top)](#table-of-contents)

- Check out our [Contributor Guide](https://github.com/kubeshop/.github/blob/main/CONTRIBUTING.md) and
  [Code of Conduct](https://github.com/kubeshop/.github/blob/main/CODE_OF_CONDUCT.md)
- Fork/Clone the repo and make sure you can run it as shown above
- Check out open [issues](https://github.com/kubeshop/kusk-gateway/issues) here on GitHub
- Get in touch with the team by starting a discussion on [GitHub](https://github.com/kubeshop/kusk-gateway/discussions) or on our [Discord Server](https://discord.gg/uNuhy6GDyn).
- or open an issue of your own that you would like to contribute to the project.

# License

[(Back to top)](#table-of-contents)

[MIT](https://mit-license.org/)
