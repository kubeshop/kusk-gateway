# OpenAPI Extension Reference

Kusk Gateway comes with an [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to accommodate everything within
an OpenAPI spec to create a source of truth for the operational behaviour of your API.

The `x-kusk` extension has the following structure:

```yaml
x-kusk:
  hosts:
    - example.com
  
  disabled: false
  
  validation:
    request:
      enabled: true # enable automatic request validation using OpenAPI definition
  mocking:
    enabled: true # Enables mocking of the responses using examples in OpenAPI responses definition.
  upstream: # upstream and redirect are mutually exclusive
    host: # host and service are mutually exclusive
      hostname: example.com
      port: 80
    service: # host and service are mutually exclusive
      namespace: default
      name: petstore
      port: 8000
    rewrite:
      pattern: 'regular_expression'
      substitution: 'substitution'
      
  redirect: # upstream and redirect are mutually exclusive
    scheme_redirect: https
    host_redirect: example.org
    port_redirect: 8081
      
    path_redirect: /index.html # path_redirect and rewrite_regex are mutually exclusive
    rewrite_regex: # path_redirect and rewrite_regex are mutually exclusive
      pattern: 'regular_expression'
      substitution: 'substitution'
        
    response_code: 308
    strip_query: true
        
        
  path:
    prefix: /api
          
  qos:
    retries: 10
    request_timeout: 60 
    idle_timeout: 30
  cors:
    origins:
      - "*"
    methods:
      - POST
      - GET
      - OPTIONS
    headers:
      - Content-Type
    credentials: true
    expose_headers:
      - X-Custom-Header1
      - X-Custom-Header2
    max_age: 86200
  websocket: true

  rate_limit:
    requests_per_unit: 2
    unit: minute
    per_connection: false
    response_code: 429

  cache:
    enabled: true
    max_age: 10
```

Check out the [OpenAPI Extension Guide](../guides/working-with-extension.md) to configure the operational aspects of your API.

## **Available Properties**

### **Disabled**

This boolean property allows you to disable the corresponding path/operation, "hiding" internal operations from being published to end users.

When set to true at the top level, all paths will be hidden; you will have to override specific paths/operations with
`disabled: false` to make those operations visible.

### **Hosts**

This string array property configures the hosts (i.e. `Host` HTTP header) list the Gateway will listen traffic for. Wildcard hosts are supported in the suffix or prefix form, exclusively, i.e.:

- *.example.org
- example.*

Read more in the [Guide on Routing](/docs/guides/routing.md#using-hosts-for-multi-hosting-scenarios).


### **CORS**

The `CORS` object sets properties for configuring [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for your API.

| Name             | Description                                                     |
|:-----------------|:----------------------------------------------------------------|
| `origins`        | List of HTTP origins accepted by the configured operations.      |
| `methods`        | List of HTTP methods accepted by the configured operations.      |
| `headers`        | List of HTTP headers accepted by the configured operations.      |
| `expose_headers` | List of HTTP headers exposed by the configured operations.       |
| `credentials`    | Boolean flag for requiring credentials.                          |
| `max_age`        | Indicates how long results of a preflight request can be cached. |

Read more in the [guide on CORS](../guides/cors.md).

### **QoS**

Options for configuring QoS settings, such as retries and timeouts.

| Name              | Description                               |
|:------------------|:------------------------------------------|
| `retries`         | Maximum number of retries (0 by default).  |
| `request_timeout` | Total request timeout (in seconds).       |
| `idle_timeout`    | Timeout for idle connections (in seconds). |

Read more in the [Guide on Timeouts](../guides/timeouts.md).

### **Websocket**

An optional boolean field that defines whether to enable handling of "Upgrade: websocket" and other actions related to Websocket HTTP headers in the request to create a Websocket tunnel to the backend. The default value is false -  don't handle Websockets.

### **Upstream**

This setting configures where the traffic goes. `service` and `host` are available and are mutually exclusive.
The `upstream` setting is mutually exclusive with `redirect` setting.

`service` is a reference to a Kubernetes Service inside the cluster, while `host` can reference any hostname, even
outside the cluster.

See the [Guide on Routing](../guides/routing.md) to learn more about this functionality.

#### **Rewrite**

Additionally, `upstream` has an optional object `rewrite`. This allows modification of the URL of the request before forwarding
it to the upstream service.

| Name                   | Description                     |
|:-----------------------|---------------------------------|
| `rewrite.pattern`      | Regular expression.              |
| `rewrite.substitution` | Regular expression substitution. |

#### **Service**

The service object sets the target Kubernetes service to receive traffic. It contains the following properties:

| Name        | Description                                      |
|:------------|:-------------------------------------------------|
| `namespace` | The namespace containing the upstream Service.    |
| `name`      | The upstream Service's name.                      |
| `port`      | The upstream Service's port. Default value is 80. |

#### **Host**

The host object sets the target host to receive traffic. It contains the following properties:

| Name       | Description                      |
|:-----------|:---------------------------------|
| `hostname` | The hostname to route traffic to. |
| `port`     | The target port to route traffic to.  |

Note: `service` and `host` are mutually exclusive since they define the same thing (the upstream host to route to).

### **Path**

The path object contains the following properties to configure service endpoints paths:

| Name     | Description                                                                              |
|:---------|------------------------------------------------------------------------------------------|
| `prefix` | Prefix for the route  ( i.e. /your-prefix/here/rest/of/the/route ). Default value is "/". |

If the `upstream.rewrite` option is not specified, the upstream service will receive the request "as is" with this prefix
still appended to the URL. If the upstream application doesn't know about this path, usually `404` is returned.

See the [Guide on Routing](../guides/routing.md) to learn more about this functionality.

### **Redirect**

Configures where to redirect the request. The redirect and upstream options are mutually exclusive.

| Name                         | Description                                                                 |
|:-----------------------------|-----------------------------------------------------------------------------|
| `scheme_redirect`            | Redirect scheme (http/https).                                              |
| `host_redirect`              | Host to redirect to.                                                        |
| `port_redirect`              | Port to redirect to.                                                         |
| `path_redirect`              | Path to redirect to.                                                         |
| `rewrite_regex.pattern`      | Regular expression (mutually exclusive with path_redirect).                  |
| `rewrite_regex.substitution` | Regular expression substitution.                                             |
| `strip_query`                | Boolean, configures whether to strip the query from the URL (default false). |
| `response_code`              | Redirect response code (301, 302, 303, 307, 308).                            |

See the [Guide on Routing](../guides/routing.md) to learn more about this functionality.

### **Validation**

The validation objects contain the following properties to configure automatic request validation:

| Name                         | Description                               |
|:-----------------------------|-------------------------------------------|
| `validation.request.enabled` | Boolean flag to enable request validation. |

See the [Guide on Validation](../guides/validation.md) to learn more about this functionality.

Note: Currently, `mocking` is incompatible with the `validation` option - the configuration deployment will fail if both are enabled.

### **Mocking**

The validation objects contain the following properties to configure automatic request validation:

| Name                 | Description                    |
|:---------------------|--------------------------------|
| `mocking.enabled`    | Boolean flag to enable mocking. |

See the [Guide on Mocking](../guides/mocking.md) to learn more about this functionality.

Note: Currently `mocking` is incompatible with the `validation` option - the configuration deployment will fail if both are enabled.

### **Rate limiting**

The rate_limit object contains the following properties to configure request rate limiting:

| Name                 | Description                    |
|:---------------------|--------------------------------|
| `rate_limit.requests_per_unit`    | How many requests API can handle per unit of time. |
| `rate_limit.unit`                 | Unit of time, can be one of the following: second, minute, hour . |
| `rate_limit.per_connection`       | Boolean flag, that specifies whether the rate limiting, should be applied per connection or in total. Default: false. |
| `rate_limit.response_code`        | HTTP response code, which is returned when rate limiting. Default: 429, Too Many Requests. |

Note: Currently, rate limiting is applied per Envoy pod - if you have more than a single Envoy pod the total request capacity will be bigger than specified in the rate_limit object. You can check how many Envoy pods you run in the `spec.size` attribute of [EnvoyFleet object](../customresources/envoyfleet.md).


### **Caching**

The cache object contains the following properties to configure HTTP caching:

| Name                 | Description                    |
|:---------------------|--------------------------------|
| `cache.enabled`      | Boolean flag to enable request validation.|
| `cache.max_age`      | Indicates how long (in seconds) results of a request can be cached.  |

Note: current support for caching is experimental. Check out [https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/cache_filter](Envoy documentation) to learn more about how it works.
