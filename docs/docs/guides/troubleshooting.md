# Troubleshooting

## **To see what routes are set up on the gateway:**
### **1. Port forward to the EnvoyFleet deployment on port 19000.**

The admin console is configured to listen on port `19000`, so you will need to port-forward to it:

```
$ kubectl port-forward -n kusk-system deployments/kusk-gateway-envoy-fleet 19000:1900
```

### **2 a. Check the config dump from the browser.**

Open `http://localhost:19000` in your browser.


### **2 b. Alternatively, use curl and jq to query routes.**

```
curl http://localhost:19000/config_dump | jq '.configs[] | select(.["@type"] == "type.googleapis.com/envoy.admin.v3.RoutesConfigDump") | .dynamic_route_configs[].route_config.virtual_hosts[].routes[]'
```

If the command hangs at all, cancel it and run it again.

## **Webhooks Timeouts During Deployment**

You may encounter an error during the resources' deployment with kubectl like:

```shell
Error from server (InternalError): error when creating "examples/todomvc/kusk-backend-api.yaml": Internal error occurred: failed calling webhook "mapi.kb.io": failed to call webhook: Post "https://kusk-gateway-webhooks-service.kusk-system.svc:443/mutate-gateway-kusk-io-v1alpha1-api?timeout=10s": context deadline exceeded
```

This means that K8s masters control plane can't call the webhooks service residing on Kusk Gateway Manager on TCP port 9443. This problem is not specific to Kusk Gateway Manager itself and is related to the configuration of your cluster and the firewall rules.

To resolve this, in your firewall settings, add port 9443 to the rule containing the list of ports allowed to be accessed by K8s masters control plane.
