# Generating API CRDs

Generate accepts your OpenAPI spec file as input either as a local file or a URL pointing to your file
and generates a Kusk Gateway compatible API resource that you can apply directly into your cluster. Use this command to automate 
API deployment workflows from an existing OpenAPI defintion. 

Configuration of the API resource is done via the `x-kusk` extension.

If the OpenAPI spec doesn't have a top-level `x-kusk` property set, it will add them for you and set
the upstream service, namespace and port to the flag values passed in respectively and set the rest of the settings to defaults.
This is enough to get you started.

If the `x-kusk` extension is already present, it will override the the upstream service, namespace and port to the flag values passed in respectively
and leave the rest of the settings as they are.

You must specify the name of the envoyfleet you wish to use to expose your API. This is because Kusk Gateway could be managing more than one.
In the future, we will add the notion of a default envoyfleet which kusk gateway will use when none is specified. i.e. kusk-gateway-envoy-fleet

If you do not specify the envoyfleet namespace, it will default to `kusk-system`.

#### Sample usage

_No name specified_

```sh
kusk api generate \
  -i spec.yaml \
  --envoyfleet.name kusk-gateway-envoy-fleet \
  --envoyfleet.namespace kusk-system
```

In the above example, kusk will use the openapi spec info.title to generate a manifest name and leave the existing `x-kusk` extension settings.

_No api namespace specified_

```sh
kusk api generate \
  -i spec.yaml \
  --name httpbin-api \
  --upstream.service httpbin \
  --upstream.port 8080 \
  --envoyfleet.name kusk-gateway-envoy-fleet
```

In the above example, as `--namespace` isn't defined, it will assume the default namespace.

_Namespace specified_

```sh
kusk api generate \
  -i spec.yaml \
  --name httpbin-api \
  --upstream.service httpbin \
  --upstream.namespace my-namespace \
  --upstream.port 8080 \
  --envoyfleet.name kusk-gateway-envoy-fleet
```

_OpenAPI spec from URL_

```sh
kusk api generate \
    -i https://raw.githubusercontent.com/$ORG_OR_USER/$REPO/myspec.yaml \
    --name httpbin-api \
    --upstream.service httpbin \
    --upstream.namespace my-namespace \
    --upstream.port 8080 \
    --envoyfleet.name kusk-gateway-envoy-fleet
```

This will fetch the OpenAPI document from the provided URL and generate a Kusk Gateway API resource

#### Flags
| Flag                   | Description                                                                                         | Required? |
|:-----------------------|:----------------------------------------------------------------------------------------------------|:---------:|
| `--name`               | the name to give the API resource e.g. --name my-api. Otherwise taken from OpenAPI info title field |     ❌     |
| `--namespace` / `-n`   | the namespace of the API resource e.g. --namespace my-namespace, -n my-namespace (default: default) |     ❌     |
| `--in` / `-i`          | file path or URL to OpenAPI spec file to generate mappings from. e.g. --in apispec.yaml             |     ✅     |
| `--upstream.service`   | name of upstream Kubernetes service                                                                 |     ❌     |
| `--upstream.namespace` | namespace of upstream service (default: default)                                                    |     ❌     |
| `--upstream.port`      | port that upstream service is exposed on (default: 80)                                              |     ❌     |
| `--envoyfleet.name`    | name of envoyfleet to use for this API                                                              |     ✅     |
| `envoyfleet.namespace` | namespace of envoyfleet to use for this API. Default: kusk-system                                   |     ❌     |

#### Example
Take a look at the [http-bin example spec](./examples/httpbin-spec.yaml)

```
kusk api generate -i ./examples/httpbin-spec.yaml --name httpbin-api --upstream.service httpbin --upstream.port 8080 --envoyfleet.name kusk-gateway-envoy-fleet
```

The output should contain the following x-kusk extension at the top level
```
...
x-kusk:
  cors: {}
  path:
    rewrite:
      pattern: ""
      substitution: ""
  upstream:
    service:
	name: httpbin
	namespace: default
	port: 8080
```
