# Kusk Custom Resources

Kusk Gateway defines a number of Kubernetes CRDs for managing its configuration. These are installed as part of the
Kusk Gateway installation process.

Kusk Gateway uses the following CRDs:

* [EnvoyFleet](envoyfleet.md) - For managing Envoy deployments.
* [API](api.md) - For using an OpenAPI definition to configure Gateway behaviour.
* [StaticRoute](staticroute.md) - For exposing static content through Kusk Gateway.
