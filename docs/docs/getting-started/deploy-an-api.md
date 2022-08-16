# Deploy an API

Now that you've installed Kusk Gateway, let's have a look at how you can use OpenAPI to configure the operational and functional parts of your API.

## **1. Create an OpenAPI definition**

Create the file `openapi.yaml`

```yaml
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

## **2. Deploy the API**

```sh
kusk api generate -i openapi.yaml | kubectl apply -f -
```

## **3. Get the Gateway's External IP**

The `kusk-gateway-envoy-fleet` LoadBalancer is the default entry point of the gateway. Copy the External-IP and have it handy for the next steps!

<pre>
kubectl get service -n kusk-system kusk-gateway-envoy-fleet
<br />
<br />
NAME                       TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)                      AGE
<br />
kusk-gateway-envoy-fleet   LoadBalancer   10.100.15.213   <b>104.198.194.37</b>   80:31833/TCP,443:3083
</pre>

## **4. Test the API**

Given we have enabled gateway-level mocks, we don't need to implement the services to be able to test the API.

Get the External IP of Kusk-gateway as indicated in [installing Kusk-gateway section](./installation/#get-the-gateways-external-ip).


```sh
$ curl 104.198.194.37/hello
Hello world!
```

In the [next section](connect-a-service-to-the-api.md), we'll cover how to connect your service to Kusk-gateway.
