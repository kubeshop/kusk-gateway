## kusk api generate

Generate a Kusk Gateway API resource from your OpenAPI spec file

### Synopsis


	Generate accepts your OpenAPI spec file as input either as a local file or a URL pointing to your file
	and generates a Kusk Gateway compatible API resource that you can apply directly into your cluster.

	Configuration of the API resource is done via the x-kusk extension.

	If the OpenAPI spec doesn't have a top-level x-kusk annotation set, it will add them for you and set
	the upstream service, namespace and port to the flag values passed in respectively and set the rest of the settings to defaults.
	This is enough to get you started

	If the x-kusk extension is already present, it will override the the upstream service, namespace and port to the flag values passed in respectively
	and leave the rest of the settings as they are.

	You must specify the name of the envoyfleet you wish to use to expose your API. This is because Kusk Gateway could be managing more than one.
	In the future, we will add the notion of a default envoyfleet which kusk gateway will use when none is specified.

	If you do not specify the envoyfleet namespace, it will default to kusk-system.

	Sample usage

	No name specified
	kusk api generate \
		-i spec.yaml \
		--envoyfleet.name kusk-gateway-envoy-fleet \
		--envoyfleet.namespace kusk-system

	In the above example, kusk will use the openapi spec info.title to generate a manifest name and leave the existing
	x-kusk extension settings

	No api namespace specified
	kusk api generate \
		-i spec.yaml \
		--name httpbin-api \
		--upstream.service httpbin \
		--upstream.port 8080 \
		--envoyfleet.name kusk-gateway-envoy-fleet

	In the above example, as --namespace isn't defined, it will assume the default namespace.

	Namespace specified
	kusk api generate \
		-i spec.yaml \
		--name httpbin-api \
		--upstream.service httpbin \
		--upstream.namespace my-namespace \
		--upstream.port 8080 \
		--envoyfleet.name kusk-gateway-envoy-fleet

	OpenAPI spec at URL
	kusk api generate \
			-i https://raw.githubusercontent.com/$ORG_OR_USER/$REPO/myspec.yaml \
			 --name httpbin-api \
			 --upstream.service httpbin \
			 --upstream.namespace my-namespace \
			 --upstream.port 8080 \
			 --envoyfleet.name kusk-gateway-envoy-fleet

	This will fetch the OpenAPI document from the provided URL and generate a Kusk Gateway API resource
	

```
kusk api generate [flags]
```

### Options

```
  -a, --apply                         to automatically apply the manifest to the cluster. Default: false
      --envoyfleet.name string        name of envoyfleet to use for this API
      --envoyfleet.namespace string   namespace of envoyfleet to use for this API. Default: kusk-system (default "kusk-system")
  -h, --help                          help for generate
  -i, --in string                     file path or URL to OpenAPI spec file to generate mappings from. e.g. --in apispec.yaml
      --name string                   the name to give the API resource e.g. --name my-api
  -n, --namespace string              the namespace of the API resource e.g. --namespace my-namespace, -n my-namespace (default "default")
  -o, --output string                 path to the location where to save the output of the command
      --upstream.namespace string     namespace of upstream service (default "default")
      --upstream.port uint32          port of upstream service (default 80)
      --upstream.service string       name of upstream service
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.kusk.yaml)
```

### SEE ALSO

* [kusk api](kusk_api.md)	 - parent command for api related functions

