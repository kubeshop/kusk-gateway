## kusk install

Install kusk-gateway, envoy-fleet, api, and dashboard in a single command

### Synopsis


	Install kusk-gateway, envoy-fleet, api, and dashboard in a single command.

	$ kusk install

	Will install kusk-gateway, a public (for your APIS) and private (for the kusk dashboard and api) 
	envoy-fleet, api, and dashboard in the kusk-system namespace using helm.

	$ kusk install --name=my-release --namespace=my-namespace

	Will create a helm release named with --name in the namespace specified by --namespace.

	$ kusk install --no-dashboard --no-api --no-envoy-fleet

	Will install kusk-gateway, but not the dashboard, api, or envoy-fleet.
	

```
kusk install [flags]
```

### Options

```
  -h, --help               help for install
      --name string        installation name (default "kusk-gateway")
      --namespace string   namespace to install in (default "kusk-system")
      --no-api             don't install the api. Setting this flag implies --no-dashboard
      --no-dashboard       don't the install dashboard
      --no-envoy-fleet     don't install any envoy fleets
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.kusk.yaml)
```

### SEE ALSO

* [kusk](kusk.md)	 - 

