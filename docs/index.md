# Kusk Gateway

Kusk Gateway is a self-service API gateway powered by [OpenAPI](https://www.openapis.org/) and [Envoy](https://www.envoyproxy.io/)!

Kusk Gateway is for you if:

- You or your team develop REST APIs running in Kubernetes
- Embrace a contract-first approach to developing your APIs using OpenAPI or Swagger
- You don't want to spend lots of time configuring hard-to-work-with ingress controllers that require a dedicated Ops Engineer

Kusk Gateway is configured through the metadata defined in your OpenAPI or Swagger document.

You can apply your API definition like any other Kubernetes resource using our custom-made Kusk Gateway [API](customresources/api.md) CustomResourceDefinition (CRDs).

Other [Custom Resources](customresources/index.md) are used to configure the Envoy Fleet which implements the gateway and specify additional routing configurations.

Proceed with our [Installation](installation.md) instructions for installing to the generic Kubernetes cluster or use these [local installation instructions](local-installation.md) to evaluate the Kusk Gateway locally with Minikube.
Once you have Kusk Gateway installed, feel free to check out the [TodoMVC](todomvc.md) example to deploy backend and frontend with Kusk Gateway.
