# API Autodeploy

Kusk gateway supports API auto deployment. 

Any REST API will typically consist of a Kubernetes pod(s) that is running the server and a Kubernetes service pointed to it.

## It just works
To expose a pre-existing REST API in the cluster, you would just have to execute this:

```sh
kubectl apply -f svc.yaml
```
Or edit an existing Kubernetes service to add the annotation `kusk-gateway/openapi-url` with a URL to the location of your OpenAPI

## Under the hood

Let's explain what is going on.

There are several convenience annotation that will allow users to easily expose their REST API through Kusk gateway.

Assuming that the user has already set up a deployment that is running their REST API server code and with the selector `my-api`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-api
  annotations:
    # NOTE: we need a sleeker URL for this
    kusk-gateway/openapi-url: https://gist.githubusercontent.com/jasmingacic/082849b29d0e06e5f018a66f4cd49ec3/raw/e91c94cc82e7591031399e0d8c563d28a62de460/openapi.yaml

    # OPTIONAL annotations
    # sets the request path prefix that your API will be reachable at via envoy
    # will default to / if not specified
    kusk-gateway/path-prefix: /my-api
    
    # sets the value that will replace the prefix defined above if defined
    # If you set value to "" then it will remove the prefix before sending
    # the request onto your service which is normally the desired behaviour
    # e.g. path-prefix = /my-api, path-prefix-substitution = ""
    # and thers is a request to /my-api/foo then your service will receieve a 
    # request to /foo as the prefix /my-api is removed and replaced by ""
    kusk-gateway/path-prefix-substitution: ""
    
    # sets the envoyfleet to use. Defaults to default envoyfleet
    # you may wish to set this to a custom envoyfleet if you have multiple
    # envoyfleets in your cluster, one of which, for example, is private to 
    # the cluster, should you wish not to expose the API to the internet
    kusk-gateway/envoy-fleet: my-private-fleet
spec:
  type: ClusterIP
  selector:
    app: my-api # aforementioned selector
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

  * this holds true for the other annotations too. If the corresponding x-kusk settings are present in the OpenAPI spec then they will be used and not overwritten. 


Upcoming features:
- Refresh interval
- Versioning
- ...

