# Accessing the Kusk Gateway Dashboard

Kusk provides a kusk dashboard command to expose and navigate easily to the Kusk Gateway Dashboard in the browser. `kusk dashboard` will start a port-forward session on port 8080, by default, to the envoyfleet serving the dashboard and will open the dashboard in the browser. By default this is kusk-gateway-private-envoy-fleet.kusk-system.

If you installed all the components using `kusk install` without changing any of the default values, running `kusk dashboard` will be sufficient to open the dashboard.

The flags --envoyfleet.namespace and --envoyfleet.name can be used to change the envoyfleet.

### Flags
|           Flag           |                                          Description                                         | Required? |
|:------------------------:|:--------------------------------------------------------------------------------------------:|:---------:|
|      `--kubeconfig`      |                                 absolute path to kube config                                 |     ❌     |
|    `--envoyfleet.name`   | kusk gateway dashboard envoy fleet service name. (default: kusk-gateway-private-envoy-fleet) |     ❌     |
| `--envoyfleet.namespace` |         kusk gateway dashboard envoy fleet service namespace. (default: kusk-system)         |     ❌     |
|     `--external-port`    |                     external port to access dashboard at. (default: 8080)                    |     ❌     |

### Examples
```
$ kusk dashboard
```

Opens the kusk gateway dashboard in the browser by exposing the default private envoy fleet on port 8080

```
$ kusk dashboard --envoyfleet.namespace=other-namespace --envoyfleet.name=other-envoy-fleet
```

Specify other envoyfleet and namespace that is serving the dashboard

```
$ kusk dashboard --external-port=9090
```

Expose dashboard on port 9090

```
$ kusk dashboard --kubeconfig=/path/to/kube/config
```
Specify path to kube config. $HOME/.kube/config is used by default.