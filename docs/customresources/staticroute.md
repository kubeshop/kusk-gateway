# StaticRoute

This resource defines manually created routing rules. It is useful to setup the routing to a non-API application, e.g. static pages, images or route to some old (possibly external to the cluster) APIs.

It is designed to overcome the shortcomings of OpenAPI based routing, one of which is the inability to configure "catch all prefixes" like **/**.
Its structure is still similar to OpenAPI spec and thus is familiar for the users.

The resource can be deployed additionally to the API resource or completely separately. Routing information from both resources will be merged with the priority given to the **API** resources.

Once the resource manifest is deployed, Kusk Gateway Manager will use it to configure routing for Envoy Fleet.
Multiple resources can exist in different namespaces, all of them will be evaluated and the configuration merged on any action with the separate resource.
Trying to apply a resource that has conflicting routes with the existing resources (i.e. same HTTP method and path) will be rejected with the error.

**Alpha Limitations**:

* currently resource **status** field is not updated by manager when the reconciliation of the configuration finishes.

## Configuration structure description

The main elements of the configuration are in **spec** field.

They specify how the incoming request is matched and what action to take.

Below is the YAML structure of the configuration, please read on further for a full explanation.

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: StaticRoute
metadata:
  name: staticroute-sample
spec:
  hosts: [<string>, <string>, ...]
  paths:
    # Consists of path matchers with HTTP methods (lowercase), which in turn either:
    # 1. "route" (proxying) to the upstream host or
    # 2. "redirect" to other endpoint
    <path_match>:
      <http_method>:
        # "route" defines proxying parameters. Mutually exlusive with "redirect".
        route:
          # upstream is the container for the upstream host specification.
          upstream:
            # host defines the hostname to proxy to. Mutually exlusive with service.
            host:
              # DNS name to proxy to
              hostname: <string>
              # Port to proxy to
              port: <int>
            # service is the convenient way to configure proxying to Kubernetes services. Mutually exlusive with the "host".
            service:
              # service name
              name: <string>
              namespace: <string>
              port: 8080
          # Optional
          qos:
            # the timeout for the upstream connection, seconds. 0 means forever, if unspecified - 15 seconds.
            request_timeout: <int>
            # the timeout for the idle upstream of downstream connection, seconds. 0 means forever, unspecified default 1h.
            idle_timeout: <int>
            # retres define how many retries to upstream with failed (50x code) requests, number. Default 1.
            retries: <int>
          # Optional
          path:
            # Rewrites path with the regex and substitution patterns. 
            rewrite:
              pattern: <string> # path regex pattern
              substitution: <string> # path substitution pattern.
          # Optional
          cors:
            # allowed origins returned in Access-Control-Allow-Origin header
            # the list of domain names
            # Note - regexs other than the wildcard ("*") are not supported right now.
            origins:
            - "*"
            # allowed methods to call this endpoint returned in Access-Control-Allow-Methods header
            # the list of methods
            methods:
            - POST
            - GET
            # allowed headers returned in Access-Control-Allow-Headers header
            # the list of headers
            headers:
            - Content-Type
            # allow browser to send credentials, returned with Access-Control-Allow-Credentials header
            credentials: <true|false>
            # allowed headers that browser can access returned with  Access-Control-Expose-Headers header
            # the list of headers
            expose_headers:
            - X-Custom-Header1
            - X-Custom-Header2
            # how long to cache this CORS information for the browser, returned with Access-Control-Max-Age.
            max_age: <int>
        # "redirect" creates HTTP redirect to other endpoint. Mutually exclusive with "route".
        redirect:
          # redirect to http or https
          scheme_redirect: <http|https>
          # redirect to this hostname
          host_redirect: <string>
          # redirects to different port
          port_redirect: <int>
          # redirect to different URL path. Mutually exlusive with rewrite_regex.
          path_redirect: "<string>"
          # redirect using the rewrite rule. Mutually exlusive with path_redirect.
          rewrite_regex:
            # regex
            pattern: <string>
            # regex parameters substitution pattern
            substitution: <string>
          # response code, by default - Permanent Redirect HTTP 308
          # available HTTP codes: 301, 302, 303, 307, 308
          response_code: <int>
      <http_method>:
        -- skipped --
```

## Request matching

We match the incoming request by HOST header, path and HTTP method.

The following fields specify matching.

**hosts** that define the list of HOST headers this configuration applies to. This will create the Envoy's VirtualHost with the same name and domain matching. Wildcards are possible, e.g. "*" means "any host".
Prefix and suffix wildcards are supported, but not both (i.e. ```example.*, *example.com```, but not ```*example*```).

**paths** is the container of URL paths + HTTP methods collection to match and handle the request during the routing decision.
*paths*.**path_match** is the URL path string, starts with / (e.g. */api*, */robots.txt*). The suffix hints how to match the request:

  * paths ending with `/` will match everything that has that path as a prefix. E.g. ```/api/``` matches ```/api/v1/id```, just ```/``` is a catch all.
  * paths without `/` will match that path exactly. E.g. just ```/resource``` matches exactly this resource with any possible URL query.  **Alpha limitations:** currently regexes are currently not supported.

*paths*.*path_match*.**http_method** adds an additional request matcher which is the lowercased HTTP method (get, post, ...). Calls to the paths with a method type that is not set here will return "404 Not Found".

## Final action on the matched request

Once the request is matched, we can decide what to do with it.

*paths*.*path_match*.http_method_match.**route|redirect** specifes the routing decision. The request can be either proxied to the upstream host (backend) or returned to the user as a redirect. Either [**redirect**](#redirect) or [**route**](#route) must be specified, but not both.

 **Alpha Limitations:** currently additional request handling (e.g. direct request response like returning 404 Not Found) is not implemented.

### Redirect

**redirect** provides HTTP redirect options with the following fields. All of them are optional but once specified enable a part of redirection behaviour.

**redirect** structure:

```yaml

redirect:
  scheme_redirect: <http|https> # redirect to http or https.
  host_redirect: <string> # redirect to this hostname.
  port_redirect: <string> # redirect to this port.
  path_redirect: <string> # redirect to this path, old path is removed. Mutually exclusive with rewrite_regex.
  rewrite_regex: # redirect to this rewritten with regex path. Mutually exclusive with path_redirect.
   pattern: <string> # path regex pattern
   substitution: <string> # path substitution pattern.
  response_code: # redirect HTTP response code to return to the user. Available HTTP codes: 301, 302, 303, 307, 308
  strip_query: <bool> # strip path query during redirect, default false.
```

**rewrite_regex** pattern matching and substitution provides a powerful mechanism to rewrite redirect path based on incoming requests.
Copy from Envoy's documentation:
> Indicates that during redirect, portions of the path that match the pattern should be rewritten, even allowing the substitution of capture groups from the pattern into the new path as specified by the rewrite substitution string. This is useful to allow application paths to be rewritten in a way that is aware of segments with variable content like identifiers.

>Examples using Google’s RE2 engine:

>    The path pattern ^/service/([^/]+)(/.*)$ paired with a substitution string of \2/instance/\1 would transform /service/foo/v1/api into /v1/api/instance/foo.
>
>    The pattern one paired with a substitution string of two would transform /xxx/one/yyy/one/zzz into /xxx/two/yyy/two/zzz.
>
>    The pattern ^(.*?)one(.*)$ paired with a substitution string of \1two\2 would replace only the first occurrence of one, transforming path /xxx/one/yyy/one/zzz into /xxx/two/yyy/one/zzz.
>
>    The pattern (?i)/xxx/ paired with a substitution string of /yyy/ would do a case-insensitive match and transform path /aaa/XxX/bbb to /aaa/yyy/bbb.

### Route

**route** specifies how the request will be proxied to the upstream with the following fields.

**route** structure:

```yaml
route:
  # upstream is the container for the upstream host specification. Either upstream.host or upstream.service must be specified.
  upstream:
    # host defines the hostname to proxy to. Mutually exlusive with service.
    host:
      # DNS hostname to proxy to
      hostname: <string>
      # host port
      port: <int>
    # service is the convenient way to configure proxying to Kubernetes services. Mutually exlusive with the "host".
    service:
      # K8s service name to proxy to
      name: <string>
      # service namespace
      namespace: <string>
      # service port
      port: <int>
  # Quality of Service for the request
  # Optional
  qos:
    # the timeout for the upstream connection, seconds. 0 means forever, if unspecified - 15 seconds.
    request_timeout: <int>
    # the timeout for the idle upstream of downstream connection, seconds. 0 means forever, unspecified default 1h.
    idle_timeout: <int>
    # retres define how many retries to upstream with failed (50x code) requests, number. Default 1.
    retries: <int>
  # What to do with the path when proxying to the upstream.
  # Optional
  path:
    # Rewrites path with the regex and substitution patterns. 
    rewrite:
      pattern: <string> # path regex pattern
      substitution: <string> # path substitution pattern.
  # Optional
  cors:
    # allowed origins returned in Access-Control-Allow-Origin header
    # the list of domain names
    # Note - regexs other than the wildcard ("*") are not supported right now.
    # WARNING - this is just the example, write your own CORS settings.
    origins:
    - "*"
    # allowed methods to call this endpoint returned in Access-Control-Allow-Methods header
    # the list of methods
    methods:
    - POST
    - GET
    # allowed headers returned in Access-Control-Allow-Headers header
    # the list of headers
    headers:
    - Content-Type
    # allow browser to send credentials, returned with Access-Control-Allow-Credentials header
    credentials: <true|false>
    # allowed headers that browser can access returned with  Access-Control-Expose-Headers header
    # the list of headers
    expose_headers:
    - X-Custom-Header1
    - X-Custom-Header2
    # how long to cache this CORS information for the browser in seconds, returned with Access-Control-Max-Age header
    max_age: <int>

```


*route*.**upstream** is a required field that defines the upstream host parameters.
We proxy using DNS hostname or local cluster K8s service parameters, which are further resolved to DNS hostname. Either *upstream*.**host** or *upstream*.**service** must be specified inside.

*route*.**path** is an optional field that specifies what to do with the URL path when proxying to the upstream - possible values right now is to rewrite it. See the rewrite_regex section in redirect action above for the explanation.

*route*.**qos** optional field is the container for request Quality of Service parameters, i.e. timeouts, failure retry policy.

*route*.**cors** optional field is the container for [Cross-Origin Resource Sharing](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) headers parameters. If this field is specified, route will be augmented with CORS preflight OPTIONS HTTP method matching. This will allow Envoy to return the response to OPTIONS request with the specified here CORS headers to the user without proxying to upstream. It is advised to read [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) before trying to configure this.

Note: the structure for CORS specified above is an example, i.e. one should write its own set of methods and headers.


## Example

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: StaticRoute
metadata:
  name: sample
spec:
  # should work with localhost, example.org, any host
  hosts: [ "localhost", "*"]
  paths:
    # Catch all prefix /
    /: 
      # HTTP method GET
      get:
        route: &root_route # here we're using YAML anchors to decrease the boilerplate for all HTTP methods - the configuration is the same.
          upstream:
            host:
              # DNS name to proxy forward to
              hostname: front.somehostname.com
              # Port to proxy to
              port: 80
      post: *root_route
      put: *root_route
      head: *root_route
      patch: *root_route
    # robots.txt is served by the new frontend. Here we use "host" to show that it can replace "service" safely.
    /robots.txt: 
      get:
        route:
          upstream:
            host:
              hostname: front.frontapps.svc.cluster.local.
              port: 80
    # GET to /oldstatic resource redirects to /static
    /oldstatic/: 
      get:
        redirect:
          # redirect to different path /oldstatic/blabla -> /static/blabla
          rewrite_regex:
            pattern: '/oldstatic/(.*)'
            substitution: '/static/\1'
          response_code: 308
    /static/:
      get:
        route:
          upstream:
            service:
              name: "front"
              namespace: "frontapps"
              port: 80
    # GET to /images/ is proxied to K8s service images in images namespace
    /images/: 
      get:
        route: 
          upstream:
            service:
              name: images
              namespace: images
              port: 8080
    # old API is served on other path with the rewrite of path to upstream
    /api/v0/:
      get:
        route: &old_api_route
          upstream:
            service:
              name: api0
              namespace: legacy
              port: 80
        path:
          # removes /api/v0 from the path when proxying to upstream
          rewrite:
            pattern: "^/api/v0"
            substitution: ""
        # Old API is slow and unreliable
        qos:
          request_timeout: 30
          idle_timeout: 60
          retries: 5
        cors:
          origins:
          - "*"
          methods:
          - GET
          - POST
          - PUT
          - PATCH
          - HEAD
          headers:
          - Content-Type
          - Content-Encoding
          credentials: false
          expose_headers:
          - X-API-VERSION
          max_age: 8600
      post: *old_api_route
      put: *old_api_route
      patch: *old_api_route
      head: *old_api_route
```
