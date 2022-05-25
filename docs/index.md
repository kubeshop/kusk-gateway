# Kusk Gateway

Kusk Gateway is a self-service API gateway powered by [OpenAPI](https://www.openapis.org/) and [Envoy](https://www.envoyproxy.io/)!

Kusk Gateway enables you to design and configure your APIs from single OpenAPI manifest and you can:

- **Rapidly prototype your APIs** by mocking your API responses, allowing your teams to instantly start building on top of your APIs without your services being implemented.
- Protect your endpoints with **automatic request and response validations**.
- Configure critical policies like request timeouts and CORs with **no coding required**.
- **Centrally control your APIs** from an Open Source dashboard.

Kusk Gateway is for you if:

- You or your team develop REST APIs running in Kubernetes.
- You embrace a contract-first approach to developing your APIs using OpenAPI or Swagger.
- You want to ramp-up time when deploying a new REST api to a cluster and you don't want to spend lots of time configuring ingress controllers that require a dedicated Ops Engineer.
- You want your REST API endpoints traffic to be observable and controllable with the easy to use settings.

Kusk Gateway has a unique way to the configuration among other API gateways as it configures itself through the metadata defined in your OpenAPI or Swagger document.

You can apply your API definition like any other Kubernetes resource using our custom-made Kusk Gateway [API](customresources/api.md) CustomResourceDefinition (CRDs).

Other [Custom Resources](customresources/overview.md) are used to configure the Envoy Fleet which implements the gateway and specifies additional routing configurations.

You can check the supported and planned features at Kusk Gateway [Roadmap](contributing/roadmap.md).

Proceed with our [Installation](getting-started/installation.md) instructions for installing to the generic Kubernetes cluster

Once you have Kusk Gateway installed, feel free to check out the [Deploy an API](getting-started/deploy-an-api.md) section to get started with Kusk Gateway.
