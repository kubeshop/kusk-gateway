# Testing Envoy with Dynamic configuration

This directory provides local testing configuration with Envoy as frontend proxy and petstore application as a backend.

The example completely maps "/" in Envoy to petstore, i.e. http://127.0.0.1:8080 will proxy to Petstore without changes.

The configuration is done with go-control-plane container using hardcoded values in resource.go.

To run:

```shell
docker-compose up
```

Envoy frontend will be available on *localhost:8080* while backend could be reached on http://172.21.0.3:8080 .

To test:

```shell
curl -v -X GET 'http://localhost:8080/pets_prefix/api/v3/pet/1' -H 'accept: application/json'
```

Envoy management interface is available on http://172.21.0.2:19000.
