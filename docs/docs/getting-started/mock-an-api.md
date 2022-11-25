# 4. Mock an API

In this section you will deploy a mocked API.

A mocked API is an API that returns fake API responses. Kusk reads the OpenAPI schema and can provide those mocked results, without the need of having an implementation of the API. 

A mock API is helpful when, for example, a frontend team does not want to get blocked by the backend team's API implementation, so the frontend team can start working on using the API when scaffolding their work.

## Create an OpenAPI definition

To configure Kusk Gateway, you will need to create an OpenAPI definition of your API. Once created, you can add the gateway configurations as OpenAPI extensions (similar to code annotations). The Kusk OpenAPI extension starts with `x-kusk`.

For mocking to work, you will need to have an `example` field under `Content-Type` section, in this case under `application/json`. 

```yaml title="api.yaml"
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  mocking:
    enabled: true
paths:
  /hello:
    get:
      responses:
        "200":
          description: A simple hello world!
          content:
            application/json:
              example: 
                message: Hello from a mocked response!
              schema:
                type: object
                properties:
                  message:
                    type: string
```

## Apply the OpenAPI definition to Kusk

```sh
kusk deploy -i api.yaml
```
```sh title="Expected output:"
🎉 successfully parsed api.yaml
✅ initiallizing deployment to fleet kusk-gateway-envoy-fleet
🎉 api.gateway.kusk.io/simple-api created
```

## Test the API 

First, you'll need to get the IP address of the API by running the following command: 

```sh
kusk ip
```
```sh title="Expected output:"
123.45.67.89
```

And now test the API using `curl`: 

```sh
curl 123.45.67.89/hello
```
```sh title="Expected output:"
{"message":"Hello from a mocked response!"}
```

:::info
If you're running a local cluster with Minikube you might not have an IP address when running `kusk ip` and you might find the following message:

```sh
EnvoyFleet doesn't have an External IP address assigned yet. Try port-forwarding by running:

 kubectl port-forward svc/kusk-gateway-envoy-fleet -n kusk-system 8080:80
```

You should run the suggested command and when using `curl` you would need to run it as follows:

```sh
curl localhost:8080/hello
```
:::
