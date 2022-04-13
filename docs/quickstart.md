# Quickstart

Now that you've installed Kusk, let's have a look of how you can use OpenAPI to configure the operational and functional parts of your API.

### 1. Create your API manifest

Create the file `api.yaml`

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: API
metadata:
  name: hello-world
spec: 
  spec: |
    openapi: 3.0.0
    info:
      title: simple-api
      version: 0.1.0
    x-kusk:
      cors:
        origins:
          - "*"
        methods:
          - GET
      mocking: 
        enabled: true
    paths:
      /hello:
        get:
          responses:
            '200':
              description: A simple hello world!
              content:
                text/plain:
                  schema:
                    type: string
                  example: Hello world!
```

Kusk-gateway relies on OpenAPI to define your APIs and configure the gateway all in one place.

In this example we have defined a simple `/hello` endpoint and configured the gateway (under `x-kusk` section) enabling CORS and API mocking.

### 2. Deploy the gateway configuration

```sh
kubectl apply -f api.yaml
```

### 3. Test your API

Given we have enable mocks, we don't need to implement the services to be able to test the API.

Get the External IP of Kusk-gateway

```sh
kubectl get svc -l "app.kubernetes.io/component=envoy-svc" --namespace kusk-system
```

```sh
NAME                      TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
kusk-gateway-envoyfleet   LoadBalancer   10.109.135.106   127.0.0.1    80:31079/TCP,443:32524/TCP   53s
```

And query the `/hello` endpoint

```sh
$ curl 127.0.0.1/hello
Hello world!
```

In the next section, we'll cover how to connect your service to Kusk-gateway.