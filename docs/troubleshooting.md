# Troubleshooting

## To see what routes are set up on the gateway
### Get the name of your envoy fleet deployment

```
❯ kubectl get deployment -n kusk-system
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
kusk-controller-manager   1/1     1            1           15m
kusk-envoy-default        1/1     1            1           2m33s
```

For this example, it's `kusk-envoy-default`. Be sure to query the correct namespace for your installation.

### Port forward to the envoy deployment on port 19000
The admin console is configured to listen on port 19000 so we will port forward to it

```
❯ kubectl port-forward deployment/kusk-envoy-default -n kusk-system 19000
Forwarding from 127.0.0.1:19000 -> 19000
Forwarding from [::1]:19000 -> 19000
Handling connection for 19000
```

### Hit localhost:19000/config_dump in your browser


### Alternatively, use curl and jq to query routes

```
curl http://localhost:19000/config_dump | jq '.configs[] | select(.["@type"] == "type.googleapis.com/envoy.admin.v3.RoutesConfigDump") | .dynamic_route_configs[].route_config.virtual_hosts[].routes[]'
```

If the command hangs at all, cancel it and run it again
