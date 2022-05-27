# `AUTH`

## Reading List

Official Docs:

* <https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter#config-http-filters-ext-authz>
* <https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto#extensions-filters-http-ext-authz-v3-extauthz>

Introduction or further reading:

* <https://ekhabarov.com/post/envoy-as-an-api-gateway-authentication-and-authorization/>

## Debugging `envoy`

In `internal/controllers/envoyfleet_resources.go`, set flags (`--log-level debug`) to:

```go
		Args: []string{
			"envoy -c /etc/envoy/envoy.yaml --log-level debug --service-node $POD_NAME",
		},
```
