# API autopilot

Kusk gateway supports API auto deployment. 

Any REST API will typically consist out of a Kubernetes pod that is running the server and a Kubernetes service pointed to it.

## It just works
To expose preexisting REST API in the cluster you would just have to execute this:

```sh
kubectl apply -f svc.yaml
```
Or edit existing Kubernetes service to add annotation `kusk-gateway/openapi-url` with URL to the location of your OpenAPI

## Under the hood

Let's explain what is going on.

We added a convenience method that will allow users to easily expose their REST API through Kusk gateway by using `kusk-gateway/openapi-url` annotation, and here is how.

Assuming that the user has already set up a pod that is running REST API server code and the pod name is `todo-backend`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: todo-backend
  annotations:
    kusk-gateway/openapi-url: https://gist.githubusercontent.com/jasmingacic/082849b29d0e06e5f018a66f4cd49ec3/raw/e91c94cc82e7591031399e0d8c563d28a62de460/openapi.yaml 
    #NOTE: we need a sleeker URL for this
spec:
  type: ClusterIP
  selector:
    app: todo-backend # aforementioned pod name
  ports:
    - port: 3000
      targetPort: http
```

Once the service is deployed to a cluster Kusk gateway will pick it up and create an API resource.
Annotation `kusk-gateway/openapi-url` contains URL to OpenAPI spec for the user. Provided API may or may not contain `x-kusk` [extension](extension.md) so the reconciler will check if `x-kusk` extension is present and:
   * if not present controller will add it and point upstream to the newly created service 
```yaml
x-kusk:
    upstream:
        name: todo-backend
        namespace: default
```
  * or if the extension is present it will check if it contains `upstream` property configured. If not it will add it to the extension otherwise it will take OpenAPI definition as is and create API resource. 


Upcoming features:
- Refresh interval
- Versioning
- ...

