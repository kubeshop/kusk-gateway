# 5. Connecting an Application

In the previous section, you've mocked the API to provide fake results so developers can start working on it. 

In this section, you will deploy an application to your Kubernetes cluster and learn how to access it using Kusk Gateway. 

## Deploy a sample application

Deploy the following sample web server that has a `/hello` route:

```sh 
kubectl create deployment hello-world --image=kubeshop/kusk-hello-world:v1.0.0
kubectl expose deployment hello-world --name hello-world-svc --port=8080
```

```sh title="Expected output:"
deployment.apps/hello-world created
service/hello-world-svc exposed
```

## Update the OpenAPI definition

First, disable Kusk's mocking of the API by delete the `mocking` section from the `api.yaml` file:

```diff
...
- mocking:
-  enabled: true
...
```

Now use the `upstream` policy with the details of the service we just created, this tells Kusk that about our deployed application:

```yaml
x-kusk:
 upstream:
  service:
    name: hello-world-svc
    namespace: default
    port: 8080
```

The resulting file should look like this:

```yaml title="api.yaml"
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  upstream:
    service:
      name: hello-world-svc
      namespace: default
      port: 8080
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

## Deploy the updates to Kusk Gateway

```sh 
kusk deploy -i api.yaml
```
```sh title="Expected output:"
🎉 successfully parsed api.yaml
✅ initiallizing deployment to fleet kusk-gateway-envoy-fleet
🎉 api.gateway.kusk.io/simple-api updated
```
:::note
You can watch the changed to the OpenAPI definition and automatically deploy it by using the watch feature: 

```sh
kusk deploy -i api.yaml -w
```
::: 

## Test the API

Again, let's get the IP address of the gateway by running: 

```sh title="Expected output:"
123.45.67.89
```

And now test the API using `curl`: 

```sh
curl 123.45.67.89/hello
```
```sh title="Expected output:"
{"message":"Hello from an implemented service!"}
```

This response is served from the deployed application. You have successfully deployed an application to Kusk Gateway!