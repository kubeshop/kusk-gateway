<!-- Add banner here -->

# Kusk Gateway

<!-- Add buttons here -->

Kusk Gateway is a self-service API gateway powered by [OpenAPI](https://www.openapis.org/) and [Envoy](https://www.envoyproxy.io/)!

Kusk Gateway is for you if:
- You or your team develop REST APIs running in Kubernetes
- Embrace a contract-first approach to developing your APIs using OpenAPI or Swagger
- You don't want to spend lots of time configuring hard-to-work-with ingress controllers that require a dedicated Ops Engineer

Kusk Gateway configures itself through the metadata defined in your OpenAPI or Swagger document.

You can apply your API definition like any other Kubernetes resource using our custom-made Kusk API CustomResourceDefinition.

# Table of contents
- [Get Started](#get-started)
  - [Installation](#installation)
  - [Installation to the Local Kubernetes cluster with Minikube](#installation)
  - [Usage](#usage)
  - [Troubleshooting](#troubleshooting)
  - [Custom Resources](#custom-resources)
- [Development](#development)
- [Contribute](#contribute)
- [License](#license)

# Get Started

See the [architecture document](docs/arch.md) for an overview of the Kusk Gateway architecture

## Installation

[(Back to top)](#table-of-contents)

If you want to quickly setup and evaluate the Kusk Gateway, then please use the [local installation instructions with Minikube](docs/local-installation.md).

Otherwise see our [Installation document](https://kubeshop.github.io/kusk-gateway/installation/) for how to install the Kusk Gateway to Kubernetes.

For the quick and impatient, [Jetstack Cert-Manager](https://cert-manager.io/docs/installation/) must be installed in the cluster and then:

```sh

# Install Kubeshop Helm repo and update it
helm repo add kubeshop https://kubeshop.github.io/helm-charts && helm repo update

# Install the Kusk Gateway with CRDs into kusk-system namespace.
# We need to wait for the kusk-gateway-manager deployment to finish the setup for the next step.
helm install kusk-gateway kubeshop/kusk-gateway -n kusk-system --create-namespace &&\
kubectl wait --for=condition=available --timeout=600s deployment/kusk-gateway-manager -n kusk-system

# Install the "default" EnvoyFleet Custom Resource, which will be used by the Kusk Gateway
# to create Envoy Fleet Deployment and Service with the type LoadBalancer
helm install kusk-gateway-default-envoyfleet kubeshop/kusk-gateway-envoyfleet -n kusk-system

```

## Usage

[(Back to top)](#table-of-contents)

Kusk Gateway configures itself via the [API CRD](docs/customeresources/api.md) that contains your embedded Swagger or OpenAPI document.

The easiest way to get started is to go through [ToDoMVC example](docs/todomvc.md).

See also [x-kusk extension documentation](docs/extension.md) and [Custom Resources](docs/customresources/index.md) for the guidelines on how to add the necessary routing information to your OpenAPI file and Kusk Gateway.

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

See [Custom Resources](https://kubeshop.github.io/kusk-gateway/customresources/) for how to develop Kusk Gateway.

## Troubleshooting

[(Back to top)](#table-of-contents)

See the [Troubleshooting](https://kubeshop.github.io/kusk-gateway/troubleshooting/) for how to troubleshoot the Kusk Gateway problems.

# Development

[(Back to top)](#table-of-contents)

See our [Development document](https://kubeshop.github.io/kusk-gateway/development/) for how to develop Kusk Gateway.

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
