# TodoMVC step by step example

This [example](/examples/todomvc) will show you how to deploy a famous [TodoMVC](https://todomvc.com/) website using Kusk Gateway.

Please refer to the [installation](/docs/installation.md) guide on how to get Kusk Gateway Manager installed into your Kubernetes cluster.

We chose the [TodoBackend](http://www.todobackend.com/) implementation for an example. The website consists of a Go-powered
[backend](/examples/todomvc/backend) (with an OpenAPI [specification](/examples/todomvc/todospec.yaml) provided) and a NodeJS [frontend](/examples/todomvc/frontend).

In order to let the website communicate with the backend, one would need to properly configure [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).
Luckily, Kusk Gateway Manager allows you to do that right in your OpenAPI specification file using **x-kusk** [extension](/docs/extension.md).

## Deploy services

First, deploy the backend and frontend services:
```
kubectl apply -f https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/todomvc/backend.yaml
kubectl apply -f https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/todomvc/frontend.yaml
```

## Configure the API (backend)

Once deployed, configure the [upstream](/docs/extension.md#upstream) so that Kusk knows where to route traffic to:
```yaml
x-kusk:
  upstream:
    service:
      namespace: default
      name: todo-backend
      port: 3000
```

Then, in order to have website calling the API, configure [CORS](/docs/extension.md#cors):
```yaml
x-kusk:
  cors:
    origins:
      - '*'
    methods:
      - POST
      - PATCH
      - DELETE
      - PUT
      - GET
      - OPTIONS
    headers:
      - Content-Type
    credentials: true
    max_age: 86200
  upstream:
    service:
      namespace: default
      name: todo-backend
      port: 3000
```

Now, in order to apply it to our cluster, you need to envelope it in a [API](/docs/customresources/api.md) CRD.
You can do it either manually:
```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: API
metadata:
  name: todo
spec:
  spec: |
    # your spec goes here
```
Or by using [kgw](https://github.com/kubeshop/kgw) CLI tool:
```
kgw api generate -i https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/todomvc/todospec.yaml --name todo > kusk-backend-api.yaml
```

You can see the result you should get [here](/examples/todomvc/kusk-backend-api.yaml).

Apply it to the cluster:
```
kubectl apply -f https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/todomvc/kusk-backend-api.yaml
```

or pipe directly from **kgw** CLI - you can even do it in your CI/CD:
```
kgw api generate -i https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/todomvc/todospec.yaml --name todo | kubectl apply -f -

```

## Configure the frontend

In order to configure access to services that do not have an OpenAPI specification,
Kusk Gateway employs [StaticRoute](/docs/customresources/staticroute.md) CRD.

Create a `kusk-frontend-route.yaml` [file](/examples/todomvc/kusk-frontend-route.yaml):

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: StaticRoute
metadata:
  name: todo-frontend
spec:
  # should work with localhost, example.org
  hosts: [ "localhost", "*"]
  paths:
  # Root goes to frontend service
    /: 
       get:
        route:
         upstream:
            service:
              namespace: default
              name: todo-frontend
              port: 3000
```

And apply it to the cluster:
```
kubectl apply -f https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/todomvc/kusk-frontend-route.yaml
```

## Test access
We assume that you have followed the [installation instructions](/docs/installation.md) and have determined the external IP of EnvoyFleet Service:

```
export EXTERNAL_IP=192.168.64.2
```

Now, open the frontend in your browser: (http://192.168.64.2:8080/) and put `http://192.168.64.2:8080/todos` as your backend endpoint:
![todobackend url prompt](todobackend-prompt.png)

You should now see the TodoMVC app running against your backend, with Kusk Gateway delivering traffic to it via EnvoyFleet:
![result](result.png)