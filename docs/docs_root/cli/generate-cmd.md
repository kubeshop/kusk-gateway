# Generating API CRDs

The `generate` command accepts your OpenAPI definition as input, either as a local file or a URL pointing to your file
and generates a Kusk Gateway compatible API resource that you can apply directly into your cluster. Use this command to automate 
API deployment workflows from an existing OpenAPI definition. 

Configuration of the API resource is done via the `x-kusk` extension.

If the OpenAPI definition doesn't have a top-level `x-kusk` property set, it will add them for you and set
the upstream service, namespace and port to the flag values passed, respectively, and set the rest of the settings to defaults.
This is enough to get you started.

If the `x-kusk` extension is already present, it will override the upstream service, namespace and port to the flag values passed, respectively,
and leave the rest of the settings as they are.

You must specify the name of the EnvoyFleet you wish to use to expose your API. Kusk Gateway could be managing more than one.
In the future, we will add a `default EnvoyFleet` which Kusk Gateway will use when none is specified. i.e., `kusk-gateway-envoy-fleet`.

If you do not specify the EnvoyFleet namespace, it will default to `kusk-system`.

#### **Usage**

_No name specified:_

```sh
kusk api generate \
  -i spec.yaml \
  --envoyfleet.name kusk-gateway-envoy-fleet \
  --envoyfleet.namespace kusk-system
```

In the above example, Kusk will use the OpenAPI definition `info.title` property to generate a manifest name and 
leave the existing `x-kusk` extension settings.

_No api namespace specified:_

```sh
kusk api generate \
  -i spec.yaml \
  --name httpbin-api \
  --upstream.service httpbin \
  --upstream.port 8080 \
  --envoyfleet.name kusk-gateway-envoy-fleet
```

In the above example, as `--namespace` isn't defined, the default namespace will be used.

_Namespace specified:_

```sh
kusk api generate \
  -i spec.yaml \
  --name httpbin-api \
  --upstream.service httpbin \
  --upstream.namespace my-namespace \
  --upstream.port 8080 \
  --envoyfleet.name kusk-gateway-envoy-fleet
```

_OpenAPI definition from URL:_

```sh
kusk api generate \
    -i https://raw.githubusercontent.com/$ORG_OR_USER/$REPO/myspec.yaml \
    --name httpbin-api \
    --upstream.service httpbin \
    --upstream.namespace my-namespace \
    --upstream.port 8080 \
    --envoyfleet.name kusk-gateway-envoy-fleet
```

This will fetch the OpenAPI document from the provided URL and generate a Kusk Gateway API resource.

#### **Example**
Take a look at the [http-bin example spec](/examples/httpbin/httpbin-spec.yaml).

```
kusk api generate -i ./examples/httpbin-spec.yaml --name httpbin-api --upstream.service httpbin --upstream.port 8080 --envoyfleet.name kusk-gateway-envoy-fleet
```

The output should contain the following x-kusk extension at the top level:
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

#### **Arguments**
| Argument                  | Description                                                                                         | Required? |
|:--------------------------|:----------------------------------------------------------------------------------------------------|:---------:|
| `--name`                  | The name to give the API resource e.g. --name my-api. Otherwise, taken from OpenAPI info title field. |     ❌     |
| `--namespace` / `-n`      | The namespace of the API resource e.g. --namespace my-namespace, -n my-namespace (default: default). |     ❌     |
| `--in` / `-i`             | The file path or URL to OpenAPI definition to generate mappings from. e.g. --in apispec.yaml.       |     ✅     |
| `--upstream.service`      | The name of upstream Kubernetes service.                                                                 |     ❌     |
| `--upstream.namespace`    | The namespace of upstream service (default: default).                                                    |     ❌     |
| `--upstream.port`         | The port that upstream service is exposed on (default: 80).                                              |     ❌     |
| `--envoyfleet.name`       | The name of envoyfleet to use for this API.                                                              |     ✅     |
| `envoyfleet.namespace`    | The namespace of envoyfleet to use for this API. Default: kusk-system.                                   |     ❌     |

