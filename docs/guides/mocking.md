# Response Mocking 

## **How can Mocking help?**

Mocking is an integral part of the API lifecycle. Providing mocks can help both API consumers (for example, FE/Mobile app developers) and 
testers to bootstrap their efforts without being dependent on the actual implementation of an API being available. It can also help prototype
API integrations and validate API designs with end-users before actually implementing them.

## **How does Mocking work?**

Kusk Gateway makes mocking of your APIs extremely easy; simply add the `x-kusk.mocking` property to your API (globally or at any other level) 
to enable the mocking of responses using the examples provided in an OpenAPI definition - [examples object](https://swagger.io/specification/#example-object).

Either `example:` (singular) or `examples:` (plural) are supported, however with multiple objects in `examples`, the response will include only one from that object map.
If both are specified, singular (`example`) has the priority over plural.

Examples are defined as a part of the response HTTP code and its contents' media-type (e.g. `application/json`) and could be different for the different media types.

As part of the mocked response, the related HTTP code is returned, but only success codes are supported for the mocking. For example, 200 and 201, but not 400, 404 or 500.
Though the OpenAPI standard allows the response code to be a range or a wildcard (e.g. "2xx"), we need to know exactly what code to return, so this should be specified exactly as integer compatible ("200").
In case the response doesn't have the response schema, only the single http code is used to mock the response, the body is not returned.
This is useful to test, for example, DELETE or PATCH operations that don't produce the body.

`mocking` is inheritable - if it is specified on the path or root level it will include every operation below it.
In case there are responses without the response schema but with the examples, these must be explicitly disabled with `mocking.enabled: false`, otherwise the configuration submission will fail.

Note: Currently `mocking` is incompatible with the `validation` option, the configuration deployment will fail if both are enabled.

## **Mocking Example**

Consider the following operation in an OpenAPI definition:

```yaml
      /mocked/{id}:
        # Enable mocking with x-kusk
        # Will enable mocking for all HTTP operations below
        x-kusk:
          mocking:
            enabled: true
        get:
          responses:
            # This HTTP code will be returned.
            '200':
              description: 'Mocked ToDos'
              content:
                application/json:
                  schema:
                    type: object
                    properties:
                      title:
                        type: string
                        description: Description of what to do
                      completed:
                        type: boolean
                      order:
                        type: integer
                        format: int32
                      url:
                        type: string
                    required:
                      - title
                      - completed
                      - order
                      - url
                  # Singular example has the priority over other examples.
                  example:
                    title: "Mocked JSON title"
                    completed: true
                    order: 13
                    url: "http://mockedURL.com"
                  examples:
                    first:
                      title: "Mocked JSON title #1"
                      completed: true
                      order: 12
                      url: "http://mockedURL.com"
                    second:
                      title: "Mocked JSON title #2"
                      completed: true
                      order: 13
                      url: "http://mockedURL.com"
                application/xml:
                  example:
                    title: "Mocked XML title"
                    completed: true
                    order: 13
                    url: "http://mockedURL.com"
                text/plain:
                  example: |
                    title: "Mocked Text title"
                    completed: true
                    order: 13
                    url: "http://mockedURL.com"
        patch:
          # Disable for patch
          x-kusk:
            mocking:
              enabled: true
        ...
```

With the example above, the response to the GET request will be different depending on the client's preferred media type when using the `Accept` header.

Below, we're using the example.com setup from the development/testing directory.

1. Curl call without specifying the **Accept** header:

    From the list of specified media types (application/json, plain/text, application/xml), this example uses our default Mocking media type - application/json:

    ```shell
    curl -v -H "Host: example.com" http://192.168.49.3/testing/mocked/multiple/1

    < HTTP/1.1 200 OK
    < content-type: application/json
    < x-kusk-mocked: true
    < date: Mon, 21 Feb 2022 14:36:52 GMT
    < content-length: 81
    < x-envoy-upstream-service-time: 0
    < server: envoy
    < 
    {"completed":true,"order":13,"title":"Mocked JSON title","url":"http://mockedURL.com"}
    ```

   The response includes the `x-kusk-mocked: true` header indicating mocking.

2. With the **Accept** header, that has application/xml as the preferred type:

    ```shell
    curl -v -H "Host: example.com" -H "Accept: application/xml"  http://192.168.49.3/testing/mocked/1
    < HTTP/1.1 200 OK
    < content-type: application/xml
    < x-kusk-mocked: true
    < date: Mon, 28 Feb 2022 08:56:46 GMT
    < content-length: 117
    < x-envoy-upstream-service-time: 0
    < server: envoy

    <doc><completed>true</completed><order>13</order><title>Mocked XML title</title><url>http://mockedURL.com</url></doc>
    ```

3. With the **Accept** header specifying multiple weighted preferred media types, text/plain with more weight.

    ```shell
    curl -v -H "Host: example.com" -H "Accept: application/json;q=0.8,text/plain;q=0.9"  http://192.168.49.3/testing/mocked/1
    < content-type: text/plain
    < x-kusk-mocked: true
    < date: Mon, 28 Feb 2022 08:56:00 GMT
    < content-length: 81
    < x-envoy-upstream-service-time: 0
    < server: envoy
    < 
    title: "Mocked Text title"
    completed: true
    order: 13
    url: "http://mockedURL.com"

    ```
