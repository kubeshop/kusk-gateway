# Troubleshooting

## **To see what routes are set up on the gateway:**
### **1. Get the name of your envoy fleet deployment.**

```
❯ kubectl get deployment -n kusk-system
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
kusk-gateway-manager   1/1     1            1           15m
kusk-gateway-envoy-default    1/1     1            1           2m33s
```

For this example, it's `kusk-envoy-default`. Be sure to query the correct namespace for your installation.

### **2. Port forward to the envoy deployment on port 19000.**
The admin console is configured to listen on port 19000, so we will port forward to it:

```
❯ kubectl port-forward deployment/kusk-envoy-default -n kusk-system 19000
Forwarding from 127.0.0.1:19000 -> 19000
Forwarding from [::1]:19000 -> 19000
Handling connection for 19000
```

### **3. Hit localhost:19000/config_dump in your browser.**


### **4. Alternatively, use curl and jq to query routes.**

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
