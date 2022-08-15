# Envoy Fleet

This resource defines the Envoy Fleet, which is the implementation of the gateway in Kubernetes based on Envoy Proxy.

Once the resource manifest is deployed to Kubernetes, it is used by Kusk Gateway Manager to set up K8s Envoy Proxy **Deployment**, 
**ConfigMap** and **Service**.

The **ConfigMap** config bootstraps Envoy Proxy to connect to the [XDS](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) service of the KGW Manager to retrieve the configuration.
In its initial state there is a minimal configuration, you have to deploy API or StaticRoute resource to set up the routing.

If the Custom Resource is uninstalled, the Manager deletes the created K8s resources.

You can deploy multiple Envoy Fleets and have multiple Gateways available.

Once the Fleet is deployed, its **status** field shows the success of the process (Deployed, Failed), so it can be shown with ```kubectl describe envoyfleet``` command.

## **Limitations**

Currently, only the success of K8s resources deployment is shown, not if the Envoy Proxy pods are alive or if the Service has the External IP Address allocated.

**Supported parameters:**

* metadata.**name** and metadata.**namespace** - Used as the Envoy Fleet ID. The Manager will supply the configuration for this specific ID - Envoy will connect to the KGW Manager with it. API/Static Route can be deployed to this fleet using their fleet name field.

* spec.**image** - The Envoy Proxy container image tag, usually envoyproxy/envoy-alpine.

* spec.**service** - Defines the configuration of the K8s Service that exposes the Envoy Proxy deployment. It is similar to the K8s Service configuration but with a limited set of fields.

* spec.service.**type** - Select the Service Type (NodePort, ClusterIP, LoadBalancer).

* spec.service.**ports** - Expose TCP ports (80, 443 or any other) routed to the ports names that Deployment exposes (http, https); ports to which the Envoy Proxy listener binds.

* spec.service.**annotations** - Add annotations to the Service that will control load balancer configuration in the specific cloud providers implementations (e.g. to set up the internal Google Cloud Platform load balancer in the Google Cloud Engine, we annotate Service with the related annotation).

* spec.service.**loadBalancerIP** - Used to specify the pre-allocated Load Balancer IP address so it won't be deleted in case the Service is deleted.

* spec.service.**externalTrafficPolicy** - Optional parameter that denotes if this Service routes external traffic to node-local or cluster-wide endpoints. **Local** preserves the client source IP and avoids a second hop for LoadBalancer and Nodeport type services, but risks potentially imbalanced traffic spreading. **Cluster** obscures the client source IP and may cause a second hop to another node, but should have good overall load-spreading. For the preservation of the real client IP in access logs, choose "Local"

* spec.**size** - Optional parameter to specify the number of Envoy Proxy pods in the K8s deployment. If not specified, defaults to 1.

* spec.**resources** - Optional parameter that configures the Envoy Proxy container CPU/Memory requests and limits. Read [Resource Management for Pods and Containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/) for the details. By default, no requests or limits are specified.

* spec.**annotations** - Optional parameter that adds additional annotations to the Envoy Proxy pods, e.g. for Prometheus scraping.

* spec.**nodeSelector**, spec.**tolerations** and spec.**affinity** - Optional parameters that provide the Envoy Proxy deployment settings for the K8s Pod scheduler. Read [Assigning Pods to Nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) to understand how can you bind Envoy pods to some types of nodes (e.g. non-preemtible node type pool) and how to ensure that Envoy pods are placed onto the different nodes for High Availability. See the YAML example below, too. The structure of these fields are the same as for K8s Deployment. These options can be used simultaneously, influencing each other.

* spec.**accesslog** - Optional parameter that defines Envoy Proxy stdout HTTP requests logging. Each Envoy pod can stream the access log to stdout. If not specified,  no streaming occurs. If specified, you must chose the **format** and, optionally, the text or JSON template to tune the output.

* spec.accesslog.**format** - Required parameter that specifies the format of the output: **JSON** (structured) or **text**. Note that JSON format doesn't preserve fields order.

* spec.accesslog.**text_template**|**json_template** - Optional parameters that could be used to specify the exact Envoy request data to log. See [Envoy's Access Logging](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) for the details. If not specified, Kusk Gateway provided defaults.

* spec.**tls** - Optional parameter that defines TLS settings for the Envoy Fleet. If not specified, the Envoy Fleet will accept only HTTP traffic.

* spec.tls.**cipherSuites** - An optional field that, when specified, the TLS listener will only support the specified cipher list when negotiating TLS 1.0 or 1.2 (this setting has no effect when negotiating TLS 1.3). If not specified, a default list will be used. Defaults are different for server (downstream) and client (upstream) TLS configurations. For more information see: [Envoy's Common TLS Configuration](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/common.proto).

* spec.tls.**tlsMinimumProtocolVersion** - An optional field specifying the minimum TLS protocol version. By default, TLSv1_2 for clients and TLSv1_0 for servers.

* spec.tls.**tlsMaximumProtocolVersion** - An optional field specifying the maximum TLS protocol version. By default, TLSv1_2 for clients and TLSv1_3 for servers.

* spec.tls.**https_redirect_hosts** - An optional field specifying the domain names to force use of HTTPS with. Non-secure HTTP requests with the matched Host header will be automatically redirected to secure HTTPS with the "301 Moved Permanently" code.

* spec.tls.**tlsSecrets** - Secret name and namespace combinations for locating TLS secrets containing TLS certificates. More than one may be specified.
Kusk Gateway Manager pulls the certificates from the secrets, extracts the matching hostnames from the SubjectAlternativeNames (SAN) certificate field and configures Envoy to use that certificate for those hostnames.

* spec.tls.tlsSecrets.**secretRef** - The name of the Kubernetes secret containing the TLS certificate.

* spec.tls.tlsSecrets.**namespace** - The namespace where the Kubernetes secret resides.

```yaml
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
  # Put any additional annotations to the Enovy pod.
  # Here we add the annotations for the Prometheus service discovery to scrape Envoy pods for the Prometheus metrics.
  # annotations:
  #   prometheus.io/scrape: 'true'
  #   prometheus.io/port: '19000'
  #   prometheus.io/path: /stats/prometheus

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
  # This is more flexible than nodeSelector scheme, but they can be specified together.
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

  # TLS configuration
  # tls:
    # cipherSuites:
    #   - ECDHE-ECDSA-AES128-SHA
    #   - ECDHE-RSA-AES128-SHA
    #   - AES128-GCM-SHA256
    # tlsMinimumProtocolVersion: TLSv1_2
    # tlsMaximumProtocolVersion: TLSv1_3
    # https_redirect_hosts:
    #     - "example.com"
    #     - "my-other-example.com"
    # tlsSecrets:
    #   - secretRef: my-cert
    #     namespace: default
```
