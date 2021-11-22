# OpenAPI extension

Kusk Gateway comes with an [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to accommodate everything within
an OpenAPI spec to make that a real source of truth for configuring the gateway.

`x-kusk` extension has the following structure:

```yaml
x-kusk:
  hosts:
    - example.com
  
  disabled: false

  upstream: # upstream and redirect are mutually exclusive
    host: # host and service are mutually exclusive
      hostname: example.com
      port: 80
    service: # host and service are mutually exclusive
      namespace: default
      service: petstore
      port: 8000
      
  redirect: # upstream and redirect are mutually exclusive
    scheme_redirect: https
    host_redirect: example.org
    port_redirect: 8081
      
    path_redirect: /index.html # path_redirect and rewrite_regex are mutually exclusive
    rewrite_regex: # path_redirect and rewrite_regex are mutually exclusive
      pattern: 'regular_expression'
      substituion: 'substitution'
        
    response_code: 308
    strip_query: true
        
        
  path:
    prefix: /api
    rewrite:
      rewrite_regex:
        pattern: 'regular_expression'
        substituion: 'substitution'
          
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
```

## Properties Overview
`x-kusk` extension can be applied at (not exclusively):
1. Top level of an OpenAPI spec:
```yaml
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
x-kusk:
  hosts: 
  - "example.org"
  disabled: false
  cors:
    ...

```

2. Path level:
```yaml
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
paths:
  /pet:
    x-kusk:
      disabled: true # disables all /pet endpoints
    post:
      ...
```
3. Method (operation) level:
```yaml
  openapi: 3.0.2
  info:
    title: Swagger Petstore - OpenAPI 3.0
  paths:
    /pet:
      post:
        x-kusk:
          upstream: # routes the POST /pet endpoint to a Kubernetes service
            service:
              namespace: default
              service: petstore
              port: 8000
        ...
```

## Property Overriding/inheritance

  `x-kusk` extension at the operation level takes precedence, i.e. overrides, what's specified at the path level, including the `disabled` option.
  Likewise, the path level settings override what's specified at the global level.

  If settings aren't specified at a path or operation level, it will inherit from the layer above. (Operation > Path > Global)

## Available properties

### disabled

This boolean property allows you to disable the corresponding path/operation, allowing you to "hide" internal operations
from being published to end users.

When set to true at the top level all paths will be hidden; you will have to override specific paths/operations with
`disabled: false` to make those operations visible.

### hosts

This string array property configures hosts list the Gateway will listen traffic for.

### cors

The `cors` object sets properties for configuring [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for your API.

| Name | Description |
| :---: | :--- |
| `origins` | list of HTTP origins accepted by the configured operations
| `methods` | list of HTTP methods accepted by the configured operations
| `headers` | list of HTTP headers accepted by the configured operations
| `expose_headers` | list of HTTP headers exposed by the configured operations
| `credentials` | boolean flag for requiring credentials
| `max_age` | indicates how long results of a preflight request can be cached


### qos

Options for configuring QoS settings, such as retries and timeouts.

| Name | Description |
| :---: | :--- |
| `retries` | maximum number of retries (0 by default)
| `request_timeout` | total request timeout (in seconds)
| `idle_timeout` | timeout for idle connections (in seconds)

Please see the documentation for each individual generator to see which of these properties they support and how they apply.

### upstream

This setting configures where the traffic goes. `service` and `host` are available and are mutually exclusive.
The `upstream` settings is mutually exclusive with `redirect` setting.

#### service

The service object sets the target service to receive traffic, it contains the following properties:

| Name | Description |
| :---: | :--- |
| `namespace` | the namespace containing the upstream Service
| `name` | the upstream Service's name
| `port` | the upstream Service's port. Default value is 80

#### host

The host object sets the target host to receive traffic, it contains the following properties:

| Name | Description |
| :---: | :--- |
| `hostname` | the hostname to route traffic to
| `port` | target port to route traffic to

### redirect
Configures where to redirect request to. Redirect and upstream options are mutually exclusive.

| Name                       | Description                                                                 |
|----------------------------|-----------------------------------------------------------------------------|
| scheme_redirect            | redirect scheme (http / https)                                              |
| host_redirect              | host to redirect to                                                         |
| port_redirect              | port to redirect to                                                         |
| path_redirect              | path to redirect to                                                         |
| rewrite_regex.pattern      | regular expression (mutually exclusive with path_redirect)                  |
| rewrite_regex.substitution | regular expression substitution                                             |
| strip_query                | boolean, configures whether to strip the query from the URL (default false) |
| response_code              | redirect response code (301, 302, 303, 307, 308)                            |


### path

The path object contains the following properties to configure service endpoints paths:

| Name                       | Description                                                                                                    |
|----------------------------|----------------------------------------------------------------------------------------------------------------|
| base                       | Base is the preceding prefix for the route  ( i.e. /your-prefix/here/rest/of/the/route ). Default value is "/" |
| rewrite_regex.pattern      | Regular expression to rewrite the URL                                                                          |
| rewrite_regex.substitution | Regular expression's substitution                                                                              |


### qos
The QoS objects contains the following properties to configure quality of service for endpoints:

| Name            | Description                                |
|-----------------|--------------------------------------------|
| retries         | Total number of 5xx retries (default is 0) |
| request_timeout | Total request timeout (seconds)            |
| idle_timeout    | Idle connection timeout (seconds)          |
