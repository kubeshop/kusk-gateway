# Quickstart

Now that you've installed Kusk Gateway, let's have a look at how you can use OpenAPI to configure the operational and functional parts of your API.

### **1. Create an API Manifest**

Create the file `api.yaml`

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: API
metadata:
  name: hello-world
spec: 
  fleet:
    name: kusk-gateway-envoy-fleet
    namespace: kusk-system
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
                  example: Hello from a mocked response!
```

Kusk Gateway relies on OpenAPI to define your APIs and configure the gateway, all in one place, using the `x-kusk` extension.

In this example, we have defined a simple `/hello` endpoint and configured the gateway (under `x-kusk` section) enabling CORS and API mocking.

### **2. Deploy the Gateway Configuration**

```sh
kubectl apply -f api.yaml
```

### **3. Test the API**

Given we have enabled gateway-level mocks, we don't need to implement the services to be able to test the API.

Get the External IP of Kusk-gateway as indicated in [installing Kusk-gateway section](../installation/#2-get-the-gateways-external-ip).

And query the `/hello` endpoint

```sh
$ curl EXTERNAL_IP/hello
Hello world!
```

In the [next section](connect-a-service-to-the-api.md), we'll cover how to connect your service to Kusk-gateway.

### **Read More**

- Kusk Gateway [API manifest](../customresources/api.md).
- The [x-kusk extension](../guides/working-with-extension.md).
- [Mocking of APIs](../guides/mocking.md) with Kusk.