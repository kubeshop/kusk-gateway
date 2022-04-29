# Validation with Kusk Gateway



The validation objects contains the following properties to configure automatic request validation:

| Name                       | Description                               |
|----------------------------|-------------------------------------------|
| validation.request.enabled | boolean flag to enable request validation |

#### strict validation of request bodies

Strict validation means that the request body must conform exactly to the schema specified in your openapi spec.
Any fields not in schema will cause the validation to fail the request/response.
To enable this, please add the following field to your schema block if the request body is of type `object`

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

Note: currently `mocking` is incompatible with the `validation` option, the configuration deployment will fail if both are enabled.
