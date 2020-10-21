{{/*
Expand the name of the chart.
*/}}
{{- define "secret-manager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "secret-manager.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "secret-manager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "secret-manager.appVersion" -}}
{{- printf "%s" (default .Chart.AppVersion .Values.image.tag) }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "secret-manager.labels" -}}
helm.sh/chart: {{ include "secret-manager.chart" . }}
{{ include "secret-manager.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ include "secret-manager.appVersion" . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "secret-manager.selectorLabels" -}}
app.kubernetes.io/name: {{ include "secret-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "secret-manager.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "secret-manager.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/* Force a failure if k8s version is too low */}}
{{- if and (not .Capabilities.APIVersions.Has "apiextensions.k8s.io/v1") .Values.installCRDs -}}
  {{- fail "\n\n ERROR: Kubernetes <1.15 clusters must install the legacy CRDs" -}}
{{- end }}
