# Request Validation 

Validating request payloads and providing meaningful error messages to consumers of an API can go a long way to provide
a better developer experience. Instead of just receiving a 40X response without further details, it can help immensely 
to know what was actually wrong in the payload - missing properties? incorrect formatting? etc. Writing the code on the server is tedious and often overlooked, making it harder for both consumers and testers to resolve issues they might have 
when working with your API.

Kusk Gateway provides end-user friendly validation against the provided OpenAPI definition automatically, without 
requiring the implementer of the API to write any code, saving development time for both BE and API consumers.

Enabling validation is straight-forward - simply add the corresponding x-kusk property to your OpenAPI definition:

```yaml
x-kusk:
  validation:
    request:
      enabled:
```

Adding this at the global level will ensure all incoming requests are validated against the corresponding OpenAPI definition
in regard to request parameters and payload. If the request does not match the specified metadata, a meaningful error is returned to the user without any request being forward to your actual API implementation.

Another positive side effect of this functionality is that it provides a "security-gate" for your API; malicious requests
that are outside your defined operations will not reach the target service where they could do potential harm.

See all available validation configuration options in the [Extension Reference](/reference/extension/#validation).

## **Strict Validation of Request Bodies**

Strict validation means that the request body must conform exactly to the schema specified in your OpenAPI spec.
Any fields not in the schema will cause the validation to fail the request/response.
To enable this, please add the following field to your schema block if the request body is of type `object`:

```yaml
paths:
  /todos/{id}:
    ...
    patch:
      ...
      requestBody:
        content:
          application/json:
            schema:
              type: object
              # if you want strict validation of request bodies, please enable this option in your OpenAPI file
              additionalProperties: false
              properties:
                title:
                  type: string
                completed:
                  type: boolean
                order:
                  type: integer
                  format: int32l
```

Note: Currently, `mocking` is incompatible with the `validation` option, the configuration deployment will fail if both are enabled.
