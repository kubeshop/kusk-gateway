## kusk dashboard

Access the kusk dashboard

### Synopsis

Access the kusk dashboard. kusk dashboard will start a port-forward session on port 8080 to the envoyfleet
serving the dashboard and will open the dashboard in the browser. By default this is kusk-gateway-private-envoy-fleet.kusk-system.
The flags --envoyfleet.namespace and --envoyfleet.name can be used to change the envoyfleet.
	

```
kusk dashboard [flags]
```

### Examples

```

	$ kusk dashboard

	Opens the kusk gateway dashboard in the browser by exposing the default private envoy fleet on port 8080

	$ kusk dashboard --envoyfleet.namespace=other-namespace --envoyfleet.name=other-envoy-fleet

	Specify other envoyfleet and namespace that is serving the dashboard

	$ kusk dashboard --external-port=9090

	Expose dashboard on port 9090
	
```

### Options

```
      --envoyfleet.name string        kusk gateway dashboard envoy fleet service name (default "kusk-gateway-private-envoy-fleet")
      --envoyfleet.namespace string   kusk gateway dashboard envoy fleet namespace (default "kusk-system")
      --external-port int             external port to access dashboard at (default 8080)
  -h, --help                          help for dashboard
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.kusk.yaml)
```

### SEE ALSO

* [kusk](kusk.md)	 - 

