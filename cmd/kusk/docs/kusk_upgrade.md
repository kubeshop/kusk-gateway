## kusk upgrade

Upgrade kusk-gateway, envoy-fleet, api, and dashboard in a single command

### Synopsis


	Upgrade kusk-gateway, envoy-fleet, api, and dashboard in a single command.

	$ kusk upgrade

	Will upgrade kusk-gateway, a public (for your APIS) and private (for the kusk dashboard and api) 
	envoy-fleet, api, and dashboard in the kusk-system namespace using helm.

	$ kusk upgrade --name=my-release --namespace=my-namespace

	Will upgrade a helm release named with --name in the namespace specified by --namespace.

	$ kusk upgrade --install

	Will upgrade kusk-gateway, the dashboard, api, and envoy-fleets and install them if they are not installed

```
kusk upgrade [flags]
```

### Options

```
  -h, --help               help for upgrade
      --install            install components if not installed
      --name string        name of release to update (default "kusk-gateway")
      --namespace string   namespace to upgrade in (default "kusk-system")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.kusk.yaml)
```

### SEE ALSO

* [kusk](kusk.md)	 - 

