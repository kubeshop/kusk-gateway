# Observability

## **Envoy Admin Interface**

Envoy's admin interface is opened on 19000 port and could be used for troubleshooting, configuration dumps, changing logs levels and other administrative tasks.

Refer to the [Troubleshooting](troubleshooting.md) on the usage.

## **Metrics**

Envoy exposes a Stats service on the admin interface.
Currently, we don't configure any stats sinks to publish the metrics, but Prometheus can discover 
Envoy pods and query them for the metrics, if pods have the following annotations:

```yaml
annotations:
  prometheus.io/scrape: 'true'
  prometheus.io/port: '19000'
  prometheus.io/path: /stats/prometheus
```

This can be configured with [EnvoyFleet Custom resource](../customresources/envoyfleet.md) spec.annotations field.

The list of exported HTTP metrics is described in [HTTP Connection Manager Statistics](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/stats). See also
[Listener Metrics](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/stats).
