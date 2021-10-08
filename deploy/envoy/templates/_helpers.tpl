{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "envoy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "envoy.config" -}}
node:
    cluster: {{ .Values.node.name }}
    id: {{ .Values.node.id }}

dynamic_resources:
    ads_config:
    api_type: GRPC
    transport_api_version: V3
    grpc_services:
        - envoy_grpc:
            cluster_name: xds_cluster
    cds_config:
    resource_api_version: V3
    ads: {}
    lds_config:
    resource_api_version: V3
    ads: {}

admin:
    access_log_path: /tmp/admin_access.log
    address:
    socket_address:
        protocol: TCP
        address: 0.0.0.0
        port_value: {{ .Values.containerAdminPort }}
static_resources:
    clusters:
    - type: STRICT_DNS
    typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
        explicit_http_config:
            http2_protocol_options: {}
    name: xds_cluster
    load_assignment:
        cluster_name: xds_cluster
        endpoints:
        - lb_endpoints:
        - endpoint:
            address:
                socket_address:
                    address: {{ .Values.xds_cluster.address }}
                    port_value: {{ .Values.xds_cluster.port }}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "envoy.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "envoy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "envoy.labels" -}}
helm.sh/chart: {{ include "envoy.chart" . }}
{{ include "envoy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "envoy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "envoy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "envoy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "envoy.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create the tag for the docker image to use
*/}}
{{- define "envoy.tag" -}}
{{- .Values.image.tag | default .Chart.AppVersion -}}
{{- end -}}
