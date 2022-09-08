# `kusk deploy`

The `deploy` accepts your OpenAPI definition as input, either as a local file or a URL pointing to your file
and deploys the API to Kusk Gateway.

Configuration of the API resource is done via the `x-kusk` extension.

#### **Usage**

```sh
kusk deploy -i spec.yaml 
```

_Using the file watcher:_

```sh
kusk deploy -i spec.yaml --watch
```

This will watch your file for changes and deploy them automatically for you.

_OpenAPI definition from URL:_

```sh
kusk deploy -i https://raw.githubusercontent.com/$ORG_OR_USER/$REPO/myspec.yaml
```

This will fetch the OpenAPI document from the provided URL and apply the API. 

#### **Arguments**
| Argument                  | Description                                                                                         | Required? |
|:--------------------------|:----------------------------------------------------------------------------------------------------|:---------:|
| `--in` / `-i`             | The file path or URL to OpenAPI definition to generate mappings from. e.g. --in apispec.yaml.       |     ✅     |
| `--watch / -w`      | Watches the file for changes.                                                                 |     ❌     |