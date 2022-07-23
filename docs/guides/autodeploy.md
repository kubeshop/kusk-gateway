# Automatic API Deployment

Kusk Gateway supports automatic configuration and deployment of an API based on kusk-gateway annotations on 
your Kubernetes Service resource, allowing you to totally automate the deployment of your API to Kusk Gateway whenever
you deploy the corresponding Service to your cluster.

The following annotations are available:

| Name                                    | Description                                                                   | Optional |
|:----------------------------------------|:------------------------------------------------------------------------------|:--------:|
| `gateway.kusk.io/openapi-url`              | The absolute URL to the OpenAPI definition to deploy.                          |          |
| `gateway.kusk.io/envoy-fleet`              | Which EnvoyFleet to use.                                                       |    X     |
| `gateway.kusk.io/path-prefix:`             | The path where your API will be hosted.                                         |    X     |
| `gateway.kusk.io/path-prefix-substitution` | What to substitute the prefix with when forwarding the request to the service. |    X     |

For example, assuming that you have set up a deployment that is running your REST API, you could deploy 
the following Kubernetes Service: 

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-api
  annotations:
    gateway.kusk.io/openapi-url: https://some-resolvablehost-name/path-to-openapi.yaml
spec:
  type: ClusterIP
  selector:
    app: my-api 
  ports:
    - port: 3000
      targetPort: http
```

Deploying this manifest with kubectl (`kubectl apply -f svc.yaml`) will make Kusk Gateway automatically: 

- Read the OpenAPI definition from the `openapi-url` annotation.
- Add corresponding `x-kusk.upstream` [extensions](../../reference/extension/#upstream) to route API requests to this Service (if not already present).
- Create an [API resource](../customresources/api.md) for the OpenAPI definition and deploy it to Kusk Gateway.

If you want to customize the mapping and/or envoy-fleet used by the API, add these as annotations:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-api
  annotations:
    gateway.kusk.io/openapi-url: https://gist.githubusercontent.com/jasmingacic/082849b29d0e06e5f018a66f4cd49ec3/raw/e91c94cc82e7591031399e0d8c563d28a62de460/openapi.yaml
    gateway.kusk.io/path-prefix: /my-api
    gateway.kusk.io/path-prefix-substitution: ""
    gateway.kusk.io/envoy-fleet: my-private-fleet
spec:
  type: ClusterIP
  selector:
    app: my-api # aforementioned selector
  ports:
    - port: 3000
      targetPort: http
```

