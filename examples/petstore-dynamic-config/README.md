# Testing Envoy with Dynamic configuration

This directory provides local testing configuration with Envoy as frontend proxy and petstore application as a backend.

The example completely maps "/" in Envoy to petstore, will proxy to Petstore without changes.

There are 1 petstore application, 1 go-control-plane applicatin and 2 front-envoy each having configuration with different ENVOY_CLUSTER_ID.

Each front-envoy will generate configuration from envoy.yaml.tmpl with different Envoy cluster name and Node ID (based on ENVOY_CLUSTER + HOSTNAME) and using go-control-plane endpoints registered in .env, which is DRY configuration available for usage in docker-compose.yaml.

On the start, go-control-plane will be built and started with hardcoded Envoy Cluster, Routes and Listener configuration that points to petstore application.
Once Front Envoy starts, it will connect with GRPC to go-control-plane with its NodeID and Cluster (Envoy Cluster) in request.
This information is used by Plane to select the necessary Envoy configuration snapshot and register the node by NodeID in SnapshotCache for watching - any updates for that NodeID SnapshotCache will be sent back with GRPC to the Envoy to provide configuration updates.

To run:

```shell
docker-compose up
```

Envoy frontends will be available on *http://172.21.0.5:8080* (Cluster1) and *http://172.21.0.6:8080* (Cluster2) while backend could be reached on http://172.21.0.3:8080 .

Envoy management interface is available on http://172.23.0.5:19000,  http://172.23.0.6:19000, there one can verify what configuration it has in config_dump.

To test:

```shell
curl -v -X GET 'http://172.21.0.5:8080/api/v3/pet/1' -H 'accept: application/json'
```


## Development

You can start go-control-plane in IDE (in whatever way it needs to be), change .env file endpoints information and once the stack is started Envoy will connect to locally running go-control-plane.