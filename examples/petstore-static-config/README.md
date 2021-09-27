# Testing Envoy as Ingress

This directory provides local testing configuration with Envoy as frontend proxy and petstore  application as a backend.

Envoy configuration is done accordingly to cut-off Petstore OpenAPI file with *x-kusk* extension configuration.

To run:

```shell
docker-compose up
```

Envoy frontend will be availlable on *localhost:8080* while backend could be reached on http://172.21.0.3:8080 .

To test:

```shell
curl -v -X GET 'http://localhost:8080/pets_prefix/api/v3/pet/1' -H 'accept: application/json'
```
