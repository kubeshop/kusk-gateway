# Kusk Dashboard

Kusk Gateway includes a browser-based dashboard for inspecting deployed APIs, Envoy Fleets and Static Routes. For APIs,
it is also possible to dig into the underlying OpenAPI definition and make ad-hoc requests to the API using the 
embedded Swagger UI. 

The dashboard is installed as part of the [Kusk Installation](../cli/install-cmd.md) and published using Kusk Gateway
via dedicated Envoy Fleet, Static Route and API resources (which you can see in the dashboard). 

After installing Kusk Gateway, access the dashboard by running:

```shell
kubectl port-forward -n kusk-system svc/kusk-gateway-private-envoy-fleet 8080:80
open http://localhost:8080/
```

This will open the dashboard, showing the kusk-gateway-api API Resource deployed through Kusk Gateway (the API that the dashboard itself uses):

![Kusk Dashboard APIs](images/dashboard-apis.png)

Use the tabs on the top left to see deployed EnvoyFleets and StaticRoutes. The following EnvoyFleets are deployed by 
default:

![Kusk Dashboard Fleets](images/dashboard-fleets.png)

Use the following Static Route for exposing the dashboard itself:

![Kusk Dashboard StaticRoutes](images/dashboard-staticroutes.png)

Selecting any object will open a details panel to the right showing corresponding information - as shown in the screenshot 
above for the dashboard Static Route.

## **Setting the API Endpoint**

By default, the dashboard uses a backend API exposed on the same host as the dashboard itself with an '/api' path prefix.
If you host the dashboard or its API differently, you can configure the API endpoint using the Settings button in the 
bottom left, which opens the following dialog:

![Kusk Dashboard Settings](images/settings.png)

You can specify any absolute URL or relative path here - just make sure that CORS is configured accordingly if you specify
a different host than the default one.


