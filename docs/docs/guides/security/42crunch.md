# `42Crunch`

Automatically run security and vulnerabilities scan on your OpenAPI definition. 



## 42Crunch Reference

| Name                 | Description                                                               | Type    | Required |
| :------------------- | ------------------------------------------------------------------------- |---------|----------|
| `security.42crunch` | Enables 42Crunch scan assesment | object    | true     |
| `security.42crunch.token` | Object holding API Token for 42Crunch | object    | true     |
| `security.42crunch.token.name` | Name of the kubernetes secret  | string    | true     |
| `security.42crunch.token.namespace` | Namespace of the kubernetes secret  | string    | true     |

## Example Configuration

A minimal example of the configuration for this filter is:

```yaml title=crunch42.yaml"
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  cors:
    origins:
      - "*"
    methods:
      - GET
      - POST
  upstream:
    service:
      name: hello-world-svc
      namespace: default
      port: 8080
  security:
    42crunch:
      token:
        name: demo-secret
        namespace: default
paths:
  /hello:
    get:
      responses:
        "200":
          description: A simple hello world!
          content:
            application/json:
              schema:
                type: object
                properties:
                  message:
                    type: string
              example:
                message: Hello from a mocked response!

```

1. Create a 42Crunch account https://platform.42crunch.com/login.
2. Get 42Crunch API token https://platform.42crunch.com/settings/tokens
3. Store the token in a Kubernetes secret

```bash
$ export 42CRUNCH_TOKEN=[your_api_token] 
$ echo $42CRUNCH_TOKEN | base64 
bXk0MmNydW5jaHRva2VuCg==
```
Copy the output and paste it under `CRUNCH42_TOKEN` property of the secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: demo-secret
  namespace: default
type: Opaque
data:
  CRUNCH42_TOKEN: bXk0MmNydW5jaHRva2VuCg==
```

4. Run `kubectl apply -f openapi.yaml` (use the example above)
5. Log in into [42Crunch](https://platform.42crunch.com/login) and look for the API collection named simple-api and within it there will be your API with all security scans ran.

