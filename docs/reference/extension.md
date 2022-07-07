# OpenAPI Extension Reference

Kusk Gateway comes with an [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to accommodate everything within
an OpenAPI spec to create a source of truth for the operational behaviour of your API.

Check out the [OpenAPI Extension Guide](../guides/working-with-extension.md) to configure the operational aspects of your API.

## **Available Properties**

### **Disabled**

This boolean property allows you to disable the corresponding path/operation, "hiding" internal operations from being published to end users.

When set to true at the top level, all paths will be hidden; you will have to override specific paths/operations with
`disabled: false` to make those operations visible.

```yaml
...
  /path:
    x-kusk:
      disabled: true
...
```

### **Hosts**

This string array property configures the hosts (i.e. `Host` HTTP header) list the Gateway will listen traffic for. Wildcard hosts are supported in the suffix or prefix form, exclusively, i.e.:

- *.example.org
- example.*

```yaml
...
x-kusk:
  hosts:
    - onehost.com
...
```

Read more in the [guide on Routing](../guides/routing.md#using-hosts-for-multi-hosting-scenarios).

### **CORS**

The `CORS` object sets properties for configuring [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for your API.

| Name             | Description                                                     |
|:-----------------|:----------------------------------------------------------------|
| `cors.origins`        | List of HTTP origins accepted by the configured operations.      |
| `cors.methods`        | List of HTTP methods accepted by the configured operations.      |
| `cors.headers`        | List of HTTP headers accepted by the configured operations.      |
| `cors.expose_headers` | List of HTTP headers exposed by the configured operations.       |
| `cors.credentials`    | Boolean flag for requiring credentials.                          |
| `cors.max_age`        | Indicates how long results of a preflight request can be cached. |

**Sample:**

```yaml
...
x-kusk:
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
    max_age: 86200
...
```

Read more in the [guide on CORS](../guides/cors.md).

### **QoS**

Options for configuring QoS settings, such as retries and timeouts.

| Name              | Description                               |
|:------------------|:------------------------------------------|
| `qos.retries`         | Maximum number of retries (0 by default).  |
| `qos.request_timeout` | Total request timeout (in seconds).       |
| `qos.idle_timeout`    | Timeout for idle connections (in seconds). |

**Sample:**

```yaml
...
x-kusk:
  qos:
    request_timeout: 60
...
```

Read more in the [guide on Timeouts](../guides/timeouts.md).

### **Websocket**

An optional boolean field that defines whether to enable handling of "Upgrade: websocket" and other actions related to Websocket HTTP headers in the request to create a Websocket tunnel to the backend. The default value is false.

**Sample:**

```yaml
...
x-kusk: 
  websocket: true
...
```

### **Upstream**

This setting configures where the traffic goes. `service` and `host` are available and are mutually exclusive.
The `upstream` setting is mutually exclusive with `redirect` setting.

`service` is a reference to a Kubernetes Service inside the cluster, while `host` can reference any hostname, even
outside the cluster.

See the [guide on Routing](../guides/routing.md) to learn more about this functionality.

#### **Rewrite**

Additionally, `upstream` has an optional object `rewrite`. This allows modification of the URL of the request before forwarding
it to the upstream service.

| Name                   | Description                     |
|:-----------------------|---------------------------------|
| `upstream.rewrite.pattern`      | Regular expression.              |
| `upstream.rewrite.substitution` | Regular expression substitution. |

**Sample:**

```yaml
...
x-kusk: 
  upstream:
    service: 
      ...
    # /foo/bar/... -> to upstream: /bar/...
    rewrite:
      pattern: "^/foo"
      substitution: ""
...
```

#### **Service**

The service object sets the target Kubernetes service to receive traffic. It contains the following properties:

| Name        | Description                                      |
|:------------|:-------------------------------------------------|
| `namespace` | The namespace containing the upstream Service.    |
| `name`      | The upstream Service's name.                      |
| `port`      | The upstream Service's port. Default value is 80. |

**Sample:**

```yaml
...
x-kusk:
  upstream:
    namespace: default
    name: svc-name
    port: 8080
...
```

#### **Host**

The host object sets the target host to receive traffic. It contains the following properties:

| Name       | Description                      |
|:-----------|:---------------------------------|
| `hostname` | The hostname to route traffic to. |
| `port`     | The target port to route traffic to.  |

Note: `service` and `host` are mutually exclusive since they define the same thing (the upstream host to route to).

**Sample:**

```yaml
...
x-kusk:
  upstream:
    hostname: domain
    port: 80
...
```

### **Path**

The path object contains the following properties to configure service endpoints paths:

| Name     | Description                                                                              |
|:---------|------------------------------------------------------------------------------------------|
| `prefix` | Prefix for the route  ( i.e. /your-prefix/here/rest/of/the/route ). Default value is "/". |

If the `upstream.rewrite` option is not specified, the upstream service will receive the request "as is" with this prefix
still appended to the URL. If the upstream application doesn't know about this path, usually `404` is returned.

**Sample:**

```yaml
...
x-kusk:
  prefix: /foo
...
```

See the [guide on Routing](../guides/routing.md) to learn more about this functionality.

### **Redirect**

Configures where to redirect the request. The redirect and upstream options are mutually exclusive.

| Name                         | Description                                                                 |
|:-----------------------------|-----------------------------------------------------------------------------|
| `redirect.scheme_redirect`            | Redirect scheme (http/https).                                              |
| `redirect.host_redirect`              | Host to redirect to.                                                        |
| `redirect.port_redirect`              | Port to redirect to.                                                         |
| `redirect.path_redirect`              | Path to redirect to.                                                         |
| `redirect.rewrite_regex.pattern`      | Regular expression (mutually exclusive with path_redirect).                  |
| `redirect.rewrite_regex.substitution` | Regular expression substitution.                                             |
| `redirect.strip_query`                | Boolean, configures whether to strip the query from the URL (default false). |
| `redirect.response_code`              | Redirect response code (301, 302, 303, 307, 308).                            |

**Sample:**

```yaml
...
x-kusk:
  redirect:
    scheme_redirect: https
    host_redirect: thenewhost.com
    response_code: 302
...
```

See the [guide on Routing](../guides/routing.md) to learn more about this functionality.

### **Validation**

The validation objects contain the following properties to configure automatic request validation:

| Name                         | Description                               |
|:-----------------------------|-------------------------------------------|
| `validation.request.enabled` | Boolean flag to enable request validation. |

See the [guide on Validation](../guides/validation.md) to learn more about this functionality.

Note: Currently, `mocking` is incompatible with the `validation` option - the configuration deployment will fail if both are enabled.

**Sample:**

```yaml
...
x-kusk:
  validation:
    request:
      enabled: true
...
```

### **Mocking**

The validation objects contain the following properties to configure automatic request validation:

| Name                 | Description                    |
|:---------------------|--------------------------------|
| `mocking.enabled`    | Boolean flag to enable mocking. |

See the [guide on Mocking](../guides/mocking.md) to learn more about this functionality.

Note: Currently `mocking` is incompatible with the `validation` option - the configuration deployment will fail if both are enabled.

**Sample:**

```yaml
...
x-kusk:
  mocking:
    enabled: true
...
```

### **Rate limiting**

The rate_limit object contains the following properties to configure request rate limiting:

| Name                              | Description                                                                                                              |
|:----------------------------------|--------------------------------------------------------------------------------------------------------------------------|
| `rate_limit.requests_per_unit`    | How many requests API can handle per unit of time.                                                                       |
| `rate_limit.unit`                 | Unit of time, can be one of the following: second, minute, hour .                                                        |
| `rate_limit.per_connection`       | Boolean flag, that specifies whether the rate limiting, should be applied per connection or in total. Default: false.    |
| `rate_limit.response_code`        | HTTP response code, which is returned when rate limiting. Default: 429, Too Many Requests.                               |

Note: Currently, rate limiting is applied per Envoy pod - if you have more than a single Envoy pod the total request capacity will be bigger than specified in the rate_limit object. You can check how many Envoy pods you run in the `spec.size` attribute of [EnvoyFleet object](../customresources/envoyfleet.md).

**Sample:**

```yaml
...
x-kusk:
  rate_limit:
    requests_per_limit: 2
    rate_limit.unit: minute
...
```
### **Caching**

The cache object contains the following properties to configure HTTP caching:

| Name                 | Description                    |
|:---------------------|--------------------------------|
| `cache.enabled`      | Boolean flag to enable request validation.|
| `cache.max_age`      | Indicates how long (in seconds) results of a request can be cached.  |

Note: current support for caching is experimental. Check out [https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/cache_filter](Envoy documentation) to learn more about how it works.

**Sample:**

```yaml
...
x-kusk:
  cache:
    enabled: true
    max_age: 60
...
```


### Authentication

The `auth` object contains the following properties to configure HTTP authentication:

| Name                          | Description                                                                                                          |
|:------------------------------|----------------------------------------------------------------------------------------------------------------------|
| `auth.scheme`                      | **Required**. The authentication scheme. Only `basic` authentication is supported at the moment.                     |
| `auth.path_prefix`                 | **Optional**. The prefix to attach to `auth-upstream.host.hostname`                                                  |
| `auth.auth-upstream`               | **Required**. Defines the upstream authentication host.                                                              |
| `auth.auth-upstream.host`          | **Required**. Defines how to reach the authentication server.                                                        |
| `auth.auth-upstream.host.hostname` | **Required**. Defines the `hostname` the authentication server is running on.                                        |
| `auth.auth-upstream.host.port`     | **Required**. Defines the port the authentication server is running on, for the given `auth-upstream.host.hostname`. |

**Sample:**

```yaml
...
x-kusk:
...
  auth:
    scheme: basic
    path_prefix: /login # optional
    auth-upstream:
      host:
        hostname: example.com
        port: 80
...
```

### Exposing OpenAPI defintion

The `openapi-path` field takes the a path name and will expose your OpenAPI defintion in defined path. 

**Sample:**

```yaml
...
x-kusk:
 openapi-path: openapi.json
...
```

This will expose your entire OpenAPI defintion, without the Kusk extensions, on `yourdomain.com/openapi.json`. 

To remove some paths or operations from the exposed OpenAPI, use the [`disabled` option](./#disabled).
