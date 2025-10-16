{{/*
Expand the name of the chart.
*/}}
{{- define "belimang.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "belimang.fullname" -}}
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
{{- define "belimang.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "belimang.labels" -}}
helm.sh/chart: {{ include "belimang.chart" . }}
{{ include "belimang.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "belimang.selectorLabels" -}}
app.kubernetes.io/name: {{ include "belimang.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
App-specific selector labels (for main application pods only)
*/}}
{{- define "belimang.appSelectorLabels" -}}
app.kubernetes.io/name: {{ include "belimang.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: app
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "belimang.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "belimang.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database URL
*/}}
{{- define "belimang.databaseUrl" -}}
{{- if .Values.postgresql.enabled -}}
postgres://{{ .Values.postgresql.auth.username }}:{{ .Values.postgresql.auth.password }}@{{ include "belimang.fullname" . }}-postgresql:5432/{{ .Values.postgresql.auth.database }}?sslmode=disable
{{- else -}}
{{ .Values.secrets.database.url }}
{{- end -}}
{{- end }}

{{/*
Redis Address
*/}}
{{- define "belimang.redisAddr" -}}
{{- if .Values.redis.enabled -}}
{{ include "belimang.fullname" . }}-redis:6379
{{- else -}}
{{ .Values.configMap.redis.addr }}
{{- end -}}
{{- end }}