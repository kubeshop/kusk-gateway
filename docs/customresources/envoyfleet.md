# Envoy Fleet

This resource defines the EnvoyFleet, which is the implementation of the gateway in Kubernetes based on Envoy Proxy.

Once the resource manifest is deployed to Kubernetes, it is used by Kusk Gateway Manager to setup K8s Envoy Proxy **Deployment**, **ConfigMap** and **Service**.

The **ConfigMap** config bootstraps Envoy Proxy to connect to the [XDS](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) service of the KGW Manager to retrieve the configuration.
In its initial state there is a minimal configuration, you have to deploy API or StaticRoute resource to setup the routing.

If the Custom Resource is uninstalled, the Manager deletes the created K8s resources.

You can deploy multiple Envoy Fleets and thus have multiple Gateways available.

Once the fleet is deployed, it **status** field shows the success of the process (Deployed, Failed), so it can be shown with ```kubectl describe envoyfleet``` command.

**Alpha Limitations**:

* currently it shows only the success of K8s resources deployment, it doesn't show if the Envoy Proxy pods are alive and if the Service has the External IP Address allocated.

Currently supported parameters:

* metadata.**name** and metadata.**namespace** are used as the EnvoyFleet ID. The Manager will supply the configuration for this specific ID - Envoy will connect to the KGW Manager with it. API/StaticRoute can be deployed to this fleet using their fleet name field.

* spec.**image** is the Envoy Proxy container image tag, usually envoyproxy/envoy-alpine.

* spec.**service** defines the configuration of the K8s Service that exposes Envoy Proxy deployment. It has similar to the K8s Service configuration but with the limited set of fields.
* spec.service.**type** - select the Service type (NodePort, ClusterIP, LoadBalancer).
* spec.service.**ports** - expose TCP ports (80, 443 or any other), routed to the ports names that Deployment exposes (http, https) - ports that Envoy Proxy listener binds to.
* spec.service.**annotations** - add annotations to the Service that will control load balancer configuration in the specific cloud providers implementations (e.g. to setup the internal Google Cloud Platform load balancer in Google Cloud Engine we annotate Service with the related annotation).
* spec.service.**loadBalancerIP** can be used to specify the pre-allocated Load Balancer IP address, so it won't be deleted in case the Service is deleted.
* spec.service.**externalTrafficPolicy** optional parameter denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints. "Local" preserves the client source IP and avoids a second hop for LoadBalancer and Nodeport type services, but risks potentially imbalanced traffic spreading. "Cluster" obscures the client source IP and may cause a second hop to another node, but should have good overall load-spreading. For the preservation of the real client ip in access logs chose "Local"

* spec.**size** optional parameter is the number of Envoy Proxy pods in the K8s deployment, defaults to 1 if not specified.

* spec.**resources** optional parameter configures Envoy Proxy container CPU/Memory requests and limits - read [Resource Management for Pods and Containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/) for the details. By default we don't specify any requests or limits.

* spec.**annotations** optional parameter adds additional annotations to the Envoy Proxy pods, e.g. for the Prometheus scraping.

* spec.**nodeSelector**, spec.**tolerations** and spec.**affinity** optional parameters provide the Envoy Proxy deployment settings for the K8s Pod scheduler. Read [Assigning Pods to Nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) to understand how can you bind Envoy pods to some type of nodes (e.g. non-preemtible node type pool) and how to ensure that Envoy pods are placed onto the different nodes for the High Availability. See the YAML example below too. The structure of these fields are the same as for K8s Deployment. All these options could be used simultaneously influencing each other.

* spec.**accesslog** optional parameter defines Envoy Proxy stdout HTTP requests logging. Each Envoy pod can stream access log to stdout. If not specified - no streaming. If specified, you must chose the **format** and optionally - text or json template to tune the output.
* spec.accesslog.**format** required parameter specifies the format of the output **json** (structured) or **text**. Note that json format doesn't preserve fields order.
* spec.accesslog.**text_template**|**json_template** optional parameters could be used to specify what exactly available Envoy request data to log. See [Envoy's Access Logging](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) for the details. If not specified any - use Kusk Gateway provided defaults.

```yaml EnvoyFleet.yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: EnvoyFleet
metadata:
  name: default
  namespace: default
spec:
  image: "envoyproxy/envoy-alpine:v1.20.0"
  service:
    # NodePort, ClusterIP, LoadBalancer
    type: LoadBalancer
    # Specify annotations to modify service behaviour, e.g. for GCP to create internal load balancer:
    # annotations:
    #   networking.gke.io/load-balancer-type: "Internal"
    # Specify preallocated static load balancer IP address if present
    #loadBalancerIP: 10.10.10.10
    ports:
    - name: http
      port: 80
      targetPort: http
    - name: https
      port: 443
      targetPort: http
  # externalTrafficPolicy: Cluster|Local
  resources:
    # limits:
    #   cpu: 1
    #   memory: 100M
    requests:
      cpu: 10m
      memory: 100M
  # Put annotations to scrape pods.
  # annotations:
  #   prometheus.io/scrape: 'true'
  #   prometheus.io/port: '9102'

  ##### Scheduling directives
  # Read https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/ for the details.

  # Optional - schedule Envoy pods to the node with the label "disktype=ssd".
  # nodeSelector:
  #   disktype: "ssd"

  # Optional - allow to be scheduled on the "tainted" node. Taint with "kubectl taint nodes node1 key1=value1:NoSchedule".
  # Taints will repel the pods from the node unless the pods have the specific toleration.
  # The line below will allow this specific Envoy pod to be scheduled there (but scheduler decideds where to put it anyway).
  # tolerations:
  # - key: "key1"
  #   operator: "Exists"
  #   effect: "NoSchedule"

  # Optional - provide pods affinity and anti-affinity settings.
  # This is more flexible than nodeSelector scheme but they can be specified together.
  # For the scalability and fault tolerance we prefer to put all Envoy pods onto different nodes - in a case one node fails we survive on others.
  # The block below will search for all matching labels of THIS "default" envoy fleet pods and will try to schedule pods onto different nodes.
  # Other fleets (if present) are not taken into consideration. You can specify nodeAffinity and podAffinity as well.
  # affinity:
  #   podAntiAffinity:
  #     requiredDuringSchedulingIgnoredDuringExecution:
  #     - labelSelector:
  #         matchExpressions:
  #         - key: app.kubernetes.io/name
  #           operator: In
  #           values:
  #           - kusk-gateway-envoy-fleet
  #         - key: fleet
  #           operator: In
  #           values:
  #           - default
  #       topologyKey: kubernetes.io/hostname

  # optional, the number of Envoy Proxy pods to start
  size: 1

  # Access logging to stdout
  # Optional, if this is missing no access logging to stdout will be done
  accesslog:
    # json|text
    format: text
    # Depending on format we can specify our own log template or if template is not specified - default Kusk Gateway will be used.
    # See https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings for the details.
    # Below are specified the examples of similar and minimalistic formats for both text and json format types.
    # Text format fields order is preserved.
    # The output example:
    # "[2021-12-15T16:50:50.217Z]" "GET" "/" "200" "1"
    text_template: |
      "[%START_TIME%]" "%REQ(:METHOD)%" "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%" "%RESPONSE_CODE%" "%DURATION%"
    # Json format fields order isn't preserved
    # The output example:
    # {"start_time":"2021-12-15T16:46:52.135Z","path":"/","response_code":200,"method":"GET","duration":1}
    json_template:
      start_time: "%START_TIME%"
      method: "%REQ(:METHOD)%"
      path: "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%"
      response_code: "%RESPONSE_CODE%"
      duration: "%DURATION%"
```
