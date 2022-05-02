# Connect a Service

Once you have [created an API](deploy-an-api.md) and mocked its responses, you would proceed to implement the services and connect them to the API. 
This section explains how you would connect your services to Kusk-gateway. 

## 1. Deploy your service

Create a `deployment.yaml` file

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

And apply it with 

```sh
kubectl apply -f deployment.yaml
```

## 2. Update the API manifest to connect the service to the gateway

Once you have finished implementing and deploying the service, you will need to stop the mocking of the API endpoint and connect the service to the gateway. 

Stop the API mocking by deleting the `mocking` section: 

```diff
...
- mocking: 
-  enabled: true
...
```

Add the `upstream` details, which are the details of the service we just created, under the `x-kusk` section to connect the gateway to the service. 

```diff
...
x-kusk:
+ upstream:
+  service:
+   name: hello-world-svc
+   namespace: default
+   port: 8080
...
```

## 3. Apply the changes

```
kubectl apply -f api.yaml
```

## 4. Test the API

Get the External IP of Kusk-gateway as indicated in [installing Kusk-gateway section](../installation/#2-get-the-gateways-external-ip) and run:

```
$ curl EXTERNAL_IP/hello
Hello from an implemented service!
```

And now you have successfully deployed an API! The approach from this "Getting Started" section of the documentation follows a [design-first](https://kubeshop.io/blog/from-design-first-to-automated-deployment-with-openapi) approach where you deployed the API first, mocked the API to and later implemented the services and connected them to the API.

You can also check the [Automatic API Deployment](reference/automatic-api-deployment.md) for a code-first (or service-first) approach. 