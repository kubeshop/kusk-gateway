# Frontend web applications

Kusk Gatewys deploys APIs but also web applications written in any language and framework. 

Web application do not use OpenAPIs, so in this case to deploy a frontend application, you will need to use Kusk's `StaticRoutes`. 

## Example

To deploy your web application, you will need to build a container of the web app. For the purposes of this demo, we've built an example web application, containerized it and pushed it to a Docker repository. We will use this container in the example. 

### Deploy the web application to Kubernetes

```sh 
kubectl create deployment web-app --image=kubeshop/kusk-web-app-example:v1.0.0
kubectl expose deployment web-app --name web-app-svc --port=3000
```

### Create a Kusk `StaticRoute` to deploy the application

```yaml title="static-route.yaml"
apiVersion: gateway.kusk.io/v1alpha1
kind: StaticRoute
metadata:
  name: web-app-sample
spec:
  fleet:
    name: kusk-gateway-envoy-fleet
    namespace: kusk-system
  upstream:
    service:
      name: web-app-svc
      namespace: default
      port: 3000
```

And apply the Custom Resource containing the deployment details of your web application: 

```sh
kubectl apply -f static-route.yaml
```

```sh title="Expected output"
staticroute.gateway.kusk.io/web-app-sample created
```

### Test that the web application is deployed correctly

Get the IP address information of Kusk: 

```sh  
kusk ip

> 12.34.56.78
```

And visit that URL in the browser: 

![Deployed application test](img/CleanShot%202022-12-06%20at%2020.09.31.png)

That's it, with these simple steps your web application is now deployed in Kubernetes using Kusk!

You'll also be surprise by how easy it is to add OAuth to your web application so easily using Kusk, check the [OAuth guide for web applications](./authentication/oauth2.md).


