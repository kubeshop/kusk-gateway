# OpenAPI Extension Reference

Kusk Gateway comes with an [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to accommodate everything within
an OpenAPI spec to create a source of truth for the operational behaviour of your API.

Check out the [OpenAPI Extension Guide](./guides/working-with-extension.md) to configure the operational aspects of your API.

## **Available Properties**

### **Hosts**

This string array property configures the hosts (i.e. `Host` HTTP header) list the Gateway will listen traffic for. Wildcard hosts are supported in the suffix or prefix form, exclusively, i.e.:

- *.example.org
- example.*

```yaml title="openapi.yaml"
...
x-kusk:
  hosts:
    - onehost.com
...
```

Read more in the [guide on Routing](./guides/routing.md#using-hosts-for-multi-hosting-scenarios).

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

```yaml title="openapi.yaml"
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

Read more in the [guide on CORS](./guides/cors.md).

### **QoS**

Options for configuring QoS settings, such as retries and timeouts.

| Name              | Description                               |
|:------------------|:------------------------------------------|
| `qos.retries`         | Maximum number of retries (0 by default).  |
| `qos.request_timeout` | Total request timeout (in seconds).       |
| `qos.idle_timeout`    | Timeout for idle connections (in seconds). |

**Sample:**

```yaml title="openapi.yaml"
...
x-kusk:
  qos:
    request_timeout: 60
...
```

Read more in the [guide on Timeouts](./guides/timeouts.md).

### **Websocket**

An optional boolean field that defines whether to enable handling of "Upgrade: websocket" and other actions related to Websocket HTTP headers in the request to create a Websocket tunnel to the backend. The default value is false.

**Sample:**

```yaml title="openapi.yaml"
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

See the [guide on Routing](./guides/routing.md) to learn more about this functionality.

#### **Rewrite**

Additionally, `upstream` has an optional object `rewrite`. This allows modification of the URL of the request before forwarding
it to the upstream service.

| Name                   | Description                     |
|:-----------------------|---------------------------------|
| `upstream.rewrite.pattern`      | Regular expression.              |
| `upstream.rewrite.substitution` | Regular expression substitution. |

**Sample:**

```yaml title="openapi.yaml"
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
| `upstream.service.namespace` | The namespace containing the upstream Service.    |
| `upstream.service.name`      | The upstream Service's name.                      |
| `upstream.service.port`      | The upstream Service's port. Default value is 80. |

**Sample:**

```yaml title="openapi.yaml"
...
x-kusk:
  upstream:
    service:
      namespace: default
      name: svc-name
      port: 8080
...
```

#### **Host**

The host object sets the target host to receive traffic. It contains the following properties:

| Name       | Description                      |
|:-----------|:---------------------------------|
| `upstream.host.hostname` | The hostname to route traffic to. |
| `upstream.host.port`     | The target port to route traffic to.  |

Note: `service` and `host` are mutually exclusive since they define the same thing (the upstream host to route to).

**Sample:**

```yaml title="openapi.yaml"
...
x-kusk:
  upstream:
    host:
      hostname: example.org
      port: 80
...
```

### **Path**

The path object contains the following properties to configure service endpoints paths:

| Name     | Description                                                                              |
|:---------|------------------------------------------------------------------------------------------|
| `path.prefix` | Prefix for the route  ( i.e. /your-prefix/here/rest/of/the/route ). Default value is "/". |

If the `upstream.rewrite` option is not specified, the upstream service will receive the request "as is" with this prefix
still appended to the URL. If the upstream application doesn't know about this path, usually `404` is returned.

**Sample:**

```yaml title="openapi.yaml"
...
x-kusk:
  path:
    prefix: /v1
...
```

See the [guide on Routing](./guides/routing.md) to learn more about this functionality.

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

```yaml title="openapi.yaml"
...
x-kusk:
  redirect:
    scheme_redirect: https
    host_redirect: thenewhost.com
    response_code: 302
...
```

See the [guide on Routing](./guides/routing.md) to learn more about this functionality.

### **Validation**

The validation objects contain the following properties to configure automatic request validation:

| Name                         | Description                               |
|:-----------------------------|-------------------------------------------|
| `validation.request.enabled` | Boolean flag to enable request validation. |

See the [guide on Validation](./guides/validation.md) to learn more about this functionality.

**Sample:**

```yaml title="openapi.yaml"
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

See the [guide on Mocking](./guides/mocking.md) to learn more about this functionality.

**Sample:**

```yaml title="openapi.yaml"
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

Note: Currently, rate limiting is applied per Envoy pod - if you have more than a single Envoy pod the total request capacity will be bigger than specified in the rate_limit object. You can check how many Envoy pods you run in the `spec.size` attribute of [EnvoyFleet object](./reference/customresources/envoyfleet.md).

**Sample:**

```yaml title="openapi.yaml"
...
x-kusk:
  rate_limit:
    requests_per_unit: 2
    unit: minute
...
```
### **Caching**

The cache object contains the following properties to configure HTTP caching:

| Name                 | Description                    |
|:---------------------|--------------------------------|
| `cache.enabled`      | Boolean flag to enable request validation.|
| `cache.max_age`      | Indicates how long (in seconds) results of a request can be cached.  |

**Sample:**

```yaml title="openapi.yaml"
...
x-kusk:
  cache:
    enabled: true
    max_age: 60
...
```

### **Authentication**

The `auth` object contains the following properties to configure HTTP authentication:

| Name                               | Description                                                                                                                                                                                                      |
|:-----------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `auth.scheme`                      | **Required**. The authentication scheme. `basic`, `cloudentity` and `oauth2` are valid values.                                                                                                                   |
| `auth.path_prefix`                 | **Optional**. The prefix to attach to `auth-upstream.host.hostname`                                                                                                                                              |
| `auth.auth-upstream`                    | **Required, if `scheme` is `basic`**. Defines the upstream authentication host.                                                                                                                                  |
| `auth.auth-upstream.host`               | **Required, if `scheme` is `basic`**. Defines how to reach the authentication server.                                                                                                                            |
| `auth.auth-upstream.host.hostname`      | **Required, if `scheme` is `basic`**. Defines the `hostname` the authentication server is running on.                                                                                                            |
| `auth.auth-upstream.host.port`          | **Required, if `scheme` is `basic`**. Defines the port the authentication server is running on, for the given `auth-upstream.host.hostname`.                                                                     |
| `auth.oauth2`                           | **Required, if `scheme` is `oauth2`**. Defines the OAuth2 configuration.                                                                                                                                         |
| `auth.oauth2.token_endpoint`            | **Required, if `scheme` is `oauth2`**. Defines the `token_endpoint`, e.g., the field `token_endpoint` from <https://kubeshop-kusk-gateway-oauth2.eu.auth0.com/.well-known/openid-configuration>.                 |
| `auth.oauth2.authorization_endpoint`    | **Required, if `scheme` is `oauth2`**. Defines the `authorization_endpoint`, e.g., the field `authorization_endpoint` from <https://kubeshop-kusk-gateway-oauth2.eu.auth0.com/.well-known/openid-configuration>. |
| `auth.oauth2.credentials.client_id`     | **Required, if `scheme` is `oauth2`**. Defines the Client ID. |
| `auth.oauth2.credentials.client_secret` | **Required, if `scheme` is `oauth2`**. Defines the Client Secret. |
| `auth.oauth2.redirect_uri`              | **Required, if `scheme` is `oauth2`**. The redirect URI passed to the authorization endpoint. |
| `auth.oauth2.signout_path`              | **Required, if `scheme` is `oauth2`**. The path to sign a user out, clearing their credential cookies. |
| `auth.oauth2.redirect_path_matcher`     | **Required, if `scheme` is `oauth2`**. After a redirecting the user back to the `redirect_uri`, using this new grant and the `token_secret`, the `kusk-gateway` then attempts to retrieve an access token from the `token_endpoint`. The `kusk-gateway` knows it has to do this instead of reinitiating another login because the incoming request has a path that matches the `redirect_path_matcher` criteria. |
| `auth.oauth2.forward_bearer_token`      | **Required, if `scheme` is `oauth2`**. If the Bearer Token should be forwarded, you generally want this to be `true`. When the authn server validates the client and returns an authorization token back to `kusk-gateway`, no matter what format that token is, if `forward_bearer_token` is set to true `kusk-gateway` will send over a cookie named `BearerToken` to the upstream. Additionally, the `Authorization` header will be populated with the same value, i.e., Forward the OAuth token as a Bearer to upstream web service. |
| `auth.oauth2.auth_scopes`               | **Optional, if `scheme` is `oauth2`**. Optional list of OAuth scopes to be claimed in the authorization request. If not specified, defaults to `user` scope. OAuth RFC <https://tools.ietf.org/html/rfc6749#section-3.3>. |
| `auth.oauth2.resources`                 | **Optional, if `scheme` is `oauth2`**. Optional list of resource parameters for authorization request RFC: <https://tools.ietf.org/html/rfc8707>. |

`auth.auth-upstream` (named as `basic` and `cloudentity`) are mutually exclusive with `oauth2`: The two cannot be specified together.

#### References

* More information about the `basic` configuration can be found at [`External Authorization — envoy`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter).
* More information about the `oauth2` configuration can be found at [`OAuth2 — envoy`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/oauth2_filter).

#### `basic` Sample

```yaml title="openapi.yaml"
...
x-kusk:
...
  auth:
    custom:
      path_prefix: /login # optional
      host:
        hostname: example.com
        port: 80
...
```

Check the [Custom Upstream auth guide](./guides/authentication/custom-auth-upstream.md) for more details on how to secure your APIs with your own custom auth service. 

#### `cloudentity` Sample

```yaml title="openapi.yaml"
...
x-kusk:
...
  auth:
    cloudentity:
      host:
        hostname: cloudentity-authorizer-standalone-authorizer.kusk-system
        port: 9004
...
```

For more details on this authorization flow, check [the guide on how to use Cloudentity with Kusk](./guides/authentication/cloudentity.md). 

#### `oauth2` Sample

The example below ensures the whole API is protected via OAuth2.

```yaml title="api.yml"
...
x-kusk:
  auth:
    scheme: oauth2
    oauth2:
      token_endpoint: *TOKEN_ENDPOINT* # <- for example https://yourdomain.eu.auth0.com/oauth/token
      authorization_endpoint: *AUTHORIZATION_ENDPOINT* # <- for example https://yourdomain.eu.auth0.com/authorize
      credentials:
        client_id: *CLIENT_ID*
        client_secret: *CLIENT_SECRET*
      redirect_uri: /oauth2/callback
      redirect_path_matcher: /oauth2/callback
      signout_path: /oauth2/signout
      forward_bearer_token: true
      auth_scopes:
        - openid
...
```

[Check the OAuth guide](./guides/authentication/oauth2.md) for more details on how to set up OAuth for your APIs.

### **Exposing OpenAPI defintion**

The `public_api_path` field takes a path name and will expose your OpenAPI definition at the defined path.

**Sample:**

```yaml title="openapi.yaml"
...
x-kusk:
 public_api_path: openapi.json
...
```

This will expose your entire OpenAPI definition, without the Kusk extensions, on `yourdomain.com/openapi.json`.

To remove some paths or operations from the exposed OpenAPI, use the [`disabled` option](./#disabled).

### **Disabled**

This boolean property allows you to hide the corresponding path/operation from being exposed and managed by Kusk.

When set to true at the top level, all paths will be disabled; you will have to override specific paths/operations with `disabled: false` to make those operations visible.

```yaml title="openapi.yaml"
...
  /path:
    x-kusk:
      disabled: true
...
```
