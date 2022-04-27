# Custom Resources

Custom Resources is the Kubernetes concept to extend K8s API with third party APIs.

They're deployed as YAML manifests files and are picked up by the Kusk Gateway Manager to configure Envoy settings and routing.

Currently we support the following Custom Resources:

* [EnvoyFleet](envoyfleet.md) - configuration for setting up Envoy Fleet.

* [API](api.md) - OpenAPI based routing configuration.

* [StaticRoute](staticroute.md) - Manually created routing configuration.
