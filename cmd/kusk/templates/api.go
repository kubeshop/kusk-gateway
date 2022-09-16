/*
The MIT License (MIT)

# Copyright © 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package templates

type APITemplateArgs struct {
	Name                string
	Namespace           string
	EnvoyfleetName      string
	EnvoyfleetNamespace string
	Spec                []string
}

var APITemplate = `
---
apiVersion: gateway.kusk.io/v1alpha1
kind: API
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  fleet:
    name: {{ .EnvoyfleetName }}
    namespace: {{ .EnvoyfleetNamespace }}
  spec: |
  {{- range $line := .Spec }}
    {{ $line -}}
  {{- end }}
`

var ConfigMapTemplate = `apiVersion: v1
data:
  AGENT_MANAGER_BIND_ADDR: :18010
  ANALYTICS_ENABLED: {{ ".AnalyticsEnabled" }}
  ENABLE_LEADER_ELECTION: "false"
  ENVOY_CONTROL_PLANE_BIND_ADDR: :18000
  HEALTH_PROBE_BIND_ADDR: :8081
  LOG_LEVEL: INFO
  METRICS_BIND_ADDR: 127.0.0.1:8080
  WEBHOOK_CERTS_DIR: /tmp/k8s-webhook-server/serving-certs
kind: ConfigMap
metadata:
  labels:
    app.kubernetes.io/component: kusk-gateway-manager
    app.kubernetes.io/instance: kusk-gateway
    app.kubernetes.io/name: kusk-gateway
    app.kubernetes.io/version: {{ .Version }}
  name: kusk-gateway-manager
  namespace: kusk-system
`
