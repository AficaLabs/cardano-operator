{{/*
Expand the name of the chart.
*/}}
{{- define "cardano-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cardano-operator.fullname" -}}
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
{{- define "cardano-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cardano-operator.labels" -}}
helm.sh/chart: {{ include "cardano-operator.chart" . }}
{{ include "cardano-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: cardano-operator
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cardano-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cardano-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
control-plane: controller-manager
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "cardano-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "cardano-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get the image reference
*/}}
{{- define "cardano-operator.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}

{{/*
Webhook certificate secret name
*/}}
{{- define "cardano-operator.webhookCertSecret" -}}
{{- printf "%s-webhook-server-cert" (include "cardano-operator.fullname" .) }}
{{- end }}

{{/*
Metrics certificate secret name
*/}}
{{- define "cardano-operator.metricsCertSecret" -}}
{{- printf "%s-metrics-server-cert" (include "cardano-operator.fullname" .) }}
{{- end }}

{{/*
Leader election ID
*/}}
{{- define "cardano-operator.leaderElectionId" -}}
{{- printf "%s.cardano.org" (include "cardano-operator.fullname" .) }}
{{- end }}
