# Connect an upstream service

Once you have [created an API](deploy-an-api.md) and mocked its responses, you are ready to implement the services and connect them to the API. 
This section explains how you would connect your services to Kusk-gateway. 

## **1. Deploy a Service**

Let's deploy a hello-world Deployment. Create `deployment.yaml` file:

```sh 
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
spec:
  selector:
    matchLabels:
      app: hello-world
  template:
    metadata:
      labels:
        app: hello-world
    spec:
      containers:
      - name: hello-world
        image: aabedraba/kusk-hello-world:1.0
        resources:
          limits:
            memory: "128Mi"
            cpu: "500m"
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world-svc
spec:
  selector:
    app: hello-world
  ports:
  - port: 8080
    targetPort: 8080
```

And apply it with: 

```sh
kubectl apply -f deployment.yaml
```

## **2. Update the API Manifest to Connect the Service to the Gateway**

Once you have finished implementing and deploying the service, you will need to stop the mocking of the API endpoint and connect the service to the gateway. 

Stop the API mocking by deleting the `mocking` section from the `openapi.yaml` file: 

```diff
...
- mocking: 
-  enabled: true
...
```

Add the `upstream` policy to the top of the `x-kusk` section of the `openapi.yaml` file, which contains the details of the service we just created:

```yaml
x-kusk:
 upstream:
  service:
    name: hello-world-svc
    namespace: default
    port: 8080
```

## **3. Apply the Changes**

```
kubectl apply -f api.yaml
```

## **4. Test the API**

Get the External IP of Kusk-gateway as indicated in [installing Kusk-gateway section](./installation/#get-the-gateways-external-ip) and run:

```
$ curl 104.198.194.37/hello
Hello from an implemented service!
```

Now you have successfully deployed an API! 

## Next Steps

The approach from this "Getting Started" section of the documentation follows a [design-first](https://kubeshop.io/blog/from-design-first-to-automated-deployment-with-openapi) approach where you deployed the API first, mocked the API to and later implemented the services and connected them to the API.

Check out the [available OpenAPI extensions](../guides/working-with-extension.md) to see all the features that you can enable in your gateway through OpenAPI. And, if you want, connect with us on [Discord](https://discord.gg/6zupCZFQbe) to tell us about your experience!