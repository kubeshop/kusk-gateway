# OpenAPI Overlays

Kusk supports code-first approaches, i.e. OpenAPI generated from code annotations, by use of OpenAPI Overlays.

[OpenAPI Overlays](https://github.com/OAI/Overlay-Specification) is a new specification that allows you to have an OpenAPI file without any Kusk extensions, and an overlay file containing the extenions you would want to add to your OpenAPI definition. 

After merging the overlay with the OpenAPI file, the resulting file is an OpenAPI definition with all the metadata added to it, ready to be deployed by Kusk.

This way, teams can generate their OpenAPI from code, and then add the gateway deployment metadata later.

## Example

Let's start with an OpenAPI definition that does not have any Kusk extensions added to it.

```yaml file="openapi.yaml"
openapi: 3.0.0
servers:
  - url: http://api.mydomain.com
info:
  title: simple-api
  version: 0.1.0
paths:
  /hello:
    get:
      summary: Returns a Hello world to the user
      responses:
        '200':
          description: A simple hello world!
          content:
            application/json; charset=utf-8:
              schema:
                type: object
                properties:
                  message:
                    type: string
                required:
                  - message
```

Now let's create an overlay that adds Kusk mocking policy to our OpenAPI definition. 

```yaml file="overlay.yaml" 
overlays: 1.0.0
extends: ./openapi.yaml
actions:
  - target: "$"
    update:
      mocking:
        enabled: true
```

To apply the overlay to the OpenAPI definition, run: 

```sh 
kusk deploy --overlay overlay.yaml
```

```sh title="Expected output"
🎉 successfully parsed
✅ initiallizing deployment to fleet kusk-gateway-envoy-fleet
api.gateway.kusk.io/simple-api created
```

If you want to look at the generated OpenAPI file before deploying it to Kusk, you can run: 

```sh 
kusk generate --overlay overlay.yaml
```

```yaml title="Expected output"
apiVersion: gateway.kusk.io/v1alpha1
kind: API
metadata:
  name: simple-api
  namespace: default
spec:
  fleet:
    name: kusk-gateway-envoy-fleet
    namespace: kusk-system
  spec: |
    openapi: 3.0.0
    servers:
    - url: http://api.mydomain.com
    components: {}
    info:
      title: simple-api
      version: 0.1.0
    x-kusk:
      mocking:
        enabled: true
    paths:
      /hello:
        get:
          responses:
            "200":
              content:
                application/json; charset=utf-8:
                  schema:
                    properties:
                      message:
                        type: string
                    required:
                    - message
                    type: object
              description: A simple hello world!
          summary: Returns a Hello world to the user
```

Overlay reference:

* **target** - property is a JSONPath selector (currently proposed by OpenAPI initiative is JMESPath).
* **extends** - defines which API spec to “extend” or which API spec should be overlayed. The value must be in form of full path either as relative or absolute path. For example `extends: overlay.yaml` won't work but `extends: ./overlay.yaml` will
* **update** - property should be a valid YAML that will be placed in the target object
* **remove** - property is a boolean - indicates that the selected target should be removed


