# Traffic splitting 

Traffic splitting in Kusk allows for traffic to be routed to 2 or more targets for the same API. The traffic to the API or API path will be split between services running different version of the same service.

## How does Traffic splitting work?

Setting property `x-kusk.upstreams` allows users to configure several endpoints for the API to which traffic will be split according to the value of `service.weight`.

```yaml
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  upstreams:
    service:
      name: simple-api-servicev1
      namespace: default
      weight: 50
    service:    
      name: simple-api-servicev2
      namespace: default
      weight: 50
..
```

The sum of the weights must be equal to 100. In this example traffic will be split equaly between services.

To debug easily Kusk will set response header `x-kusk-weighted-cluster:[service_name]` for each weighted service. 