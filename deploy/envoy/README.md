# envoy

![Version: 0.0.9](https://img.shields.io/badge/Version-0.0.9-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.18.2](https://img.shields.io/badge/AppVersion-v1.18.2-informational?style=flat-square)

Helm chart to deploy [envoy](https://www.envoyproxy.io/).

**Homepage:** <https://github.com/slamdev/helm-charts/tree/master/charts/envoy>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| slamdev | valentin.fedoskin@gmail.com |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | affinity for scheduler pod assignment |
| args | list | `[]` | extra args to pass to container |
| configYaml | string | `"admin:\n  access_log_path: /tmp/admin_access.log\n  address:\n    socket_address:\n      protocol: TCP\n      address: 0.0.0.0\n      port_value: 9901\nstatic_resources:\n  listeners:\n  - name: listener_0\n    address:\n      socket_address:\n        protocol: TCP\n        address: 0.0.0.0\n        port_value: 10000\n    filter_chains:\n    - filters:\n      - name: envoy.filters.network.http_connection_manager\n        typed_config:\n          \"@type\": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager\n          stat_prefix: ingress_http\n          access_log:\n          - name: envoy.access_loggers.file\n            typed_config:\n              \"@type\": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog\n              # For the demo config in the Docker container we use:\n              #   - system logs -> `/dev/stderr`\n              #   - (listener) access_logs -> `/dev/stdout`\n              path: /dev/stdout\n          route_config:\n            name: local_route\n            virtual_hosts:\n            - name: local_service\n              domains: [\"*\"]\n              routes:\n              - match:\n                  prefix: \"/\"\n                route:\n                  host_rewrite_literal: www.envoyproxy.io\n                  cluster: service_envoyproxy_io\n          http_filters:\n          - name: envoy.filters.http.router\n  clusters:\n  - name: service_envoyproxy_io\n    connect_timeout: 30s\n    type: LOGICAL_DNS\n    # Comment out the following line to test on v6 networks\n    dns_lookup_family: V4_ONLY\n    lb_policy: ROUND_ROBIN\n    load_assignment:\n      cluster_name: service_envoyproxy_io\n      endpoints:\n      - lb_endpoints:\n        - endpoint:\n            address:\n              socket_address:\n                address: www.envoyproxy.io\n                port_value: 443\n    transport_socket:\n      name: envoy.transport_sockets.tls\n      typed_config:\n        \"@type\": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext\n        sni: www.envoyproxy.io"` | config yaml |
| containerAdminPort | int | `9901` |  |
| containerPort | int | `10000` | container port, should match admin port_value from config.yaml |
| env | string | `nil` | environment variables for the deployment |
| fullnameOverride | string | `""` | full name of the chart. |
| image.pullPolicy | string | `"IfNotPresent"` | image pull policy |
| image.repository | string | `"envoyproxy/envoy"` | image repository |
| image.tag | string | `""` | image tag (chart's appVersion value will be used if not set) |
| imagePullSecrets | list | `[]` | image pull secret for private images |
| ingress.annotations | object | `{}` | ingress annotations |
| ingress.enabled | bool | `false` | enables Ingress for envoy |
| ingress.hosts | list | `[]` | ingress accepted hostnames |
| ingress.tls | list | `[]` | ingress TLS configuration |
| livenessProbe.httpGet.path | string | `"/"` | path for liveness probe |
| livenessProbe.httpGet.port | string | `"http"` | port for liveness probe |
| nameOverride | string | `""` | override name of the chart |
| nodeSelector | object | `{}` | node for scheduler pod assignment |
| podSecurityContext | object | `{}` | specifies security settings for a pod |
| readinessProbe.httpGet.path | string | `"/"` | path for readiness probe |
| readinessProbe.httpGet.port | string | `"http"` | port for readiness probe |
| replicaCount | int | `1` | number of replicas for haproxy deployment. |
| resources | object | `{}` | custom resource configuration |
| service.annotations | object | `{}` | annotations to add to the service |
| service.port | int | `80` | service port |
| service.type | string | `"ClusterIP"` | service type |
| serviceAccount.annotations | object | `{}` | annotations to add to the service account |
| serviceAccount.create | bool | `false` | specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | the name of the service account to use; if not set and create is true, a name is generated using the fullname template |
| serviceMonitor.additionalLabels | object | `{}` | additional labels for service monitor |
| serviceMonitor.enabled | bool | `false` | ServiceMonitor CRD is created for a prometheus operator |
| tolerations | list | `[]` | tolerations for scheduler pod assignment |
| volumeMounts | string | `nil` | volume mounts |
| volumes | string | `nil` | volumes |
