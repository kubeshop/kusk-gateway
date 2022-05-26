# Request Timeouts 

Kusk Gateway makes it easy to specify timeouts for your API operations, both globally and individually, for each path or operation. 

For example the API below defines a global 60-second request timeout, which is overridden for the getHello operation and set to 10 seconds:

```yaml
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  qos:
    request_timeout: 60
paths:
  /hello:
    get:
      operationId: getHello
      x-kusk:
        qos:
          request_timeout: 10
   ..
```

See all available timeout configuration options in the [Extension Reference](/reference/extension/#qos).