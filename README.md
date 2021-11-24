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
  - [Usage](#usage)
  - [Custom Resources](#custom-resources)
- [Development](#development)
- [Contribute](#contribute)
- [License](#license)

# Get Started

See the [architecture document](docs/arch.md) for an overview of the Kusk Gateway architecture

## Installation

[(Back to top)](#table-of-contents)

See our [Installation document](https://kubeshop.github.io/kusk-gateway/installation/) for how to install Kusk Gateway with Helm or how to get kusk gateway running locally.

## Usage

[(Back to top)](#table-of-contents)

Kusk Gateway configures itself via the API CRD that contains your embedded Swagger or OpenAPI document.

See [x-kusk extension documentation](docs/extension.md) for the guidelines on how to add the necessary routing information to your OpenAPI file.

After that all that's required is to apply it as you would any other Kubernetes resource.

The easiest way to get started is to use our httpbin example, found in `examples/httpbin`.

`kubectl apply -f examples/httpbin`

Grab the loadbalancer IP

`external_ip=$(kubectl -n kusk-system get svc kusk-envoy-svc-default --template="{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}")`

Curl the httpbin service

```
❯ curl http://$external_ip:8080/
{
  "args": {},
  "headers": {
    "Accept": "*/*",
    "Host": "192.168.64.2:8080",
    "User-Agent": "curl/7.64.1",
    "X-Envoy-Expected-Rq-Timeout-Ms": "15000",
    "X-Envoy-Original-Path": "/"
  },
  "origin": "172.17.0.1",
  "url": "http://192.168.64.2:8080/get"
}
```

### API CRD Format

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

See [httpbin API Resource](examples/httpbin/httpbin_v1_api.yaml) for a full example

## Custom Resources

[(Back to top)](#table-of-contents)

See [Custom Resources](https://kubeshop.github.io/kusk-gateway/customresources/) for how to develop Kusk Gateway.

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
