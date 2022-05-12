# OpenAPI extension Reference

Kusk Gateway comes with an [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to accommodate everything within
an OpenAPI spec to make that a real source of truth for operational behaviour of your API.

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
```

Check out the [OpenAPI Extension Guide](../guides/working-with-extension.md) to learn how it can be used to configure operational aspects
of your API.

## Available properties

### disabled

This boolean property allows you to disable the corresponding path/operation, allowing you to "hide" internal operations
from being published to end users.

When set to true at the top level all paths will be hidden; you will have to override specific paths/operations with
`disabled: false` to make those operations visible.

### hosts

This string array property configures hosts (i.e. `Host` HTTP header) list the Gateway will listen traffic for. Wildcard hosts are supported in the suffix or prefix form, exclusively, i.e.:

- *.example.org
- example.*

Read more in the [guide on routing](../guides/routing/#using-hosts-for-multi-hosting-scenarios)


### cors

The `cors` object sets properties for configuring [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for your API.

| Name             | Description                                                     |
|:-----------------|:----------------------------------------------------------------|
| `origins`        | list of HTTP origins accepted by the configured operations      |
| `methods`        | list of HTTP methods accepted by the configured operations      |
| `headers`        | list of HTTP headers accepted by the configured operations      |
| `expose_headers` | list of HTTP headers exposed by the configured operations       |
| `credentials`    | boolean flag for requiring credentials                          |
| `max_age`        | indicates how long results of a preflight request can be cached |

Read more in the [guide on CORS](../guides/cors.md)

### qos

Options for configuring QoS settings, such as retries and timeouts.

| Name              | Description                               |
|:------------------|:------------------------------------------|
| `retries`         | maximum number of retries (0 by default)  |
| `request_timeout` | total request timeout (in seconds)        |
| `idle_timeout`    | timeout for idle connections (in seconds) |

Read more in the [guide on timeouts](../guides/timeouts.md)

### websocket

An optional boolean field defines whether to enable handling of "Upgrade: websocket" and other related to Websocket HTTP headers in the request to create a Websocket tunnel to the backend. By default false, don't handle Websockets.

### upstream

This setting configures where the traffic goes. `service` and `host` are available and are mutually exclusive.
The `upstream` settings is mutually exclusive with `redirect` setting.

`service` is a reference to a Kubernetes Service inside the cluster, while `host` can reference any hostname, even
outside the cluster.

See the [Guide on Routing](../guides/routing.md) to learn more about this functionality.

#### rewrite

Additionally, `upstream` has an optional object `rewrite`. It allows to modify the URL of the request before forwarding
it to the upstream service.

| Name                   | Description                     |
|:-----------------------|---------------------------------|
| `rewrite.pattern`      | regular expression              |
| `rewrite.substitution` | regular expression substitution |

#### service

The service object sets the target Kubernetes service to receive traffic, it contains the following properties:

| Name        | Description                                      |
|:------------|:-------------------------------------------------|
| `namespace` | the namespace containing the upstream Service    |
| `name`      | the upstream Service's name                      |
| `port`      | the upstream Service's port. Default value is 80 |

#### host

The host object sets the target host to receive traffic, it contains the following properties:

| Name       | Description                      |
|:-----------|:---------------------------------|
| `hostname` | the hostname to route traffic to |
| `port`     | target port to route traffic to  |

Note: `service` and `host` are mutually exclusive since they define the same thing (the upstream host to route to).

### path

The path object contains the following properties to configure service endpoints paths:

| Name     | Description                                                                              |
|:---------|------------------------------------------------------------------------------------------|
| `prefix` | Prefix for the route  ( i.e. /your-prefix/here/rest/of/the/route ). Default value is "/" |

If `upstream.rewrite` option is not specified then the upstream service will receive the request "as is" with this prefix
still appended to the URL. If the upstream application doesn't know about this path, usually `404` is returned.

See the [Guide on Routing](../guides/routing.md) to learn more about this functionality.

### redirect

Configures where to redirect request to. Redirect and upstream options are mutually exclusive.

| Name                         | Description                                                                 |
|:-----------------------------|-----------------------------------------------------------------------------|
| `scheme_redirect`            | redirect scheme (http / https)                                              |
| `host_redirect`              | host to redirect to                                                         |
| `port_redirect`              | port to redirect to                                                         |
| `path_redirect`              | path to redirect to                                                         |
| `rewrite_regex.pattern`      | regular expression (mutually exclusive with path_redirect)                  |
| `rewrite_regex.substitution` | regular expression substitution                                             |
| `strip_query`                | boolean, configures whether to strip the query from the URL (default false) |
| `response_code`              | redirect response code (301, 302, 303, 307, 308)                            |

See the [Guide on Routing](../guides/routing.md) to learn more about this functionality.

### validation

The validation objects contains the following properties to configure automatic request validation:

| Name                         | Description                               |
|:-----------------------------|-------------------------------------------|
| `validation.request.enabled` | boolean flag to enable request validation |

See the [Guide on Validation](../guides/validation.md) to learn more about this functionality.

Note: currently `mocking` is incompatible with the `validation` option, the configuration deployment will fail if both are enabled.

### mocking

The validation objects contains the following properties to configure automatic request validation:

| Name                 | Description                    |
|:---------------------|--------------------------------|
| `mocking.enabled`    | boolean flag to enable mocking |

See the [Guide on Mocking](../guides/mocking.md) to learn more about this functionality.

Note: currently `mocking` is incompatible with the `validation` option, the configuration deployment will fail if both are enabled.

### rate_limit

The rate_limit object contains the following properties to configure request rate limiting:

| Name                 | Description                    |
|:---------------------|--------------------------------|
| `rate_limit.requests_per_unit`    | how many requests API can handle per unit of time. |
| `rate_limit.unit`                 | unit of time, can be one of the following: second, minute, hour . |
| `rate_limit.per_connection`       | boolean flag, that specifies whether the rate limiting, should be applied per connection or in total. |
| `rate_limit.response_code`        | HTTP response code, which is returned when rate limiting. Typically 429, Too Many Requests. |


Note: currently `rate_limiting` is applied per Envoy process, which means that if you have more than a single Envoy deployed the total request capacity will be bigger than specified in the extension.
