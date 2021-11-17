# Kusk Gateway

Kusk Gateway is a self-service API gateway powered by [OpenAPI](https://www.openapis.org/) and [Envoy](https://www.envoyproxy.io/)!

Kusk Gateway is for you if:
- You or your team develop REST APIs running in Kubernetes
- Embrace a contract-first approach to developing your APIs using OpenAPI or Swagger
- You don't want to spend lots of time configuring hard-to-work-with ingress controllers that require a dedicated Ops Engineer

Kusk Gateway configures itself through the metadata defined in your OpenAPI or Swagger document.

You can apply your API definition like any other Kubernetes resource using our custom-made Kusk API CustomResourceDefinition.