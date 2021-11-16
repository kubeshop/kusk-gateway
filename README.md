<!-- Add banner here -->

# Kusk Gateway

<!-- Add buttons here -->

Kusk Gateway is a self-service API gateway powered by [OpenAPI](https://www.openapis.org/) and [Envoy](https://www.envoyproxy.io/)!

Kusk Gateway is for you if:
- You or your team develop REST APIs running in Kubernetes
- Embrace a contract-first approach to developing your APIS using OpenAPI or Swagger
- You don't want to spend lots of time configuring hard-to-work-with ingress controllers that require a dedicated Ops Engineer

Kusk Gateway configures itself through the metadata defined in your OpenAPI or Swagger document.

You can apply your API definition like any other Kubernetes resource using our custom-made Kusk API CustomResourceDefinition.

See our [announcement blog post](...) for full details!

# Table of contents
- [Get Started](#get-started)
  - [Installation](#installation)
  - [Usage](#usage)
- [Development](#development)
- [Contribute](#contribute)
- [License](#license)

# Get Started
## Installation
[(Back to top)](#table-of-contents)

See our [Installation document](https://kubeshop.github.io/kusk-gateway/installation/) for how to install Kusk Gateway with Helm or how to get kusk gateway running locally.

## Usage
[(Back to top)](#table-of-contents)

### API CRD Format
```
apiVersion: gateway.kusk.io/v1
kind: API
metadata:
  name: httpbin-sample
spec:
  # service name and port should be specified inside x-kusk annotation
  spec: |
    $YOUR_API_DOCUMENT_HERE
```

See [httpbin API Resource](examples/httpbin/httpbin_v1_api.yaml) for a full example


TODO(#65) - provide link to fill x-kusk documentation

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

