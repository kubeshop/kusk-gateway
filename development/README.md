# Local development with Envoy, Kusk-gateway and OpenAPI files

This directory provides local development configuration with kusk-gatway, Envoy as frontend proxy and petstore application as a backend.

There are 1 petstore application and 1 front-envoy.

Preliminary steps:

```shell
# From the project root
cp development/.env.example ./.env
```

For the development change PROJECT_ROOT/.env file to point GO_CONTROL_PLANE_ADDRESS and GO_CONTROL_PLANE_PORT variables to ip address and port your kusk-gateway is listening on.
This will allow Envoy instance to connect to your application in IDE.

Front-envoy will generate configuration from envoy.yaml.tmpl with "default" Envoy cluster name and Node ID based on ENVOY_CLUSTER + HOSTNAME.

Kusk-gateway will consume OpenAPI file, passed with "--in" parameter and will switch to "local" mode that will skip Kubernetes connection.

Once Front Envoy starts, it will connect to kusk-gateway with GRPC with its NodeID and Cluster ("default") fields specified and kusk-gateway will provide generated configuration.

To run:

```shell
# From the project root
docker-compose up
```

Envoy frontends will be available on *http://172.21.0.5:8080* (Cluster1) and *http://172.21.0.6:8080* (Cluster2) while backend (petstore app) could be reached on http://172.21.0.3:8080 .

On MacOS, the frontends are available on *http://localhost:8080* (Cluster1) and *http://localhost:8081* (Cluster2)

Envoy management interface is available on *http://172.21.0.5:19000*,  *http://172.21.0.6:19000*, there one can verify what configuration it has in config_dump.

On MacOS, the Envoy management interface is available on *http://localhost:19000* and *http://localhost:19001*  

To test (depends on configured variables in your OpenAPI file):

```shell
# Linux
curl -v -X GET 'http://172.21.0.5:8080/api/v3/pet/findByStatus?status=available' -H 'accept: application/json'

# MacOS
curl -v -X GET 'http://localhost:8080/api/v3/pet/findByStatus?status=available' -H 'accept: application/json'
```

For the convenience you can use provided petshop-openapi-short-with-kusk.yaml file in this directory.
