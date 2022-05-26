# CORS Settings

CORS (Cross-Origin Resource Sharing) is a standard implemented by browsers for ensuring that only the allowed clients actually access your API,
see [https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).

One of the biggest pain-points when deploying your API that is consumed by a browser application is 
not having the correct CORS configuration on your API Server. Fortunately, Kusk makes configuring CORS 
for your API easy - add the corresponding CORS extension to your
OpenAPI definition at the desired level (usually the root):

```yaml
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
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
..
```

If you want to override CORS settings for a specific operation or path you can do so. For example, to change the allowed origins for a specific operation you could add:

```yaml

paths:
  /hello:
    get:
      operationId: getHello
      x-kusk:
        cors:
          origins:
            - "gethello.com"
      ..
```

See all available CORS configuration options in the [Extension Reference](../../reference/extension/#cors).