{{/*
Expand the name of the chart.
*/}}
{{- define "pocket.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "pocket.fullname" -}}
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
{{- define "pocket.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "pocket.labels" -}}
helm.sh/chart: {{ include "pocket.chart" . }}
{{ include "pocket.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "pocket.selectorLabels" -}}
app.kubernetes.io/name: {{ include "pocket.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
pokt.network/purpose: {{ .Values.nodeType }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "pocket.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "pocket.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Determine the PostgreSQL hostname based on whether the subchart is enabled.
*/}}
{{- define "pocket.postgresqlHost" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "%s-%s" .Release.Name "postgresql" -}}
{{- else -}}
{{- .Values.externalPostgresql.host -}}
{{- end -}}
{{- end -}}

{{/*
Determine the PostgreSQL port based on whether the subchart is enabled.
*/}}
{{- define "pocket.postgresqlPort" -}}
{{- if .Values.postgresql.enabled -}}
{{- .Values.global.postgresql.service.ports.postgresql | toString -}}
{{- else -}}
{{- .Values.externalPostgresql.port | toString -}}
{{- end -}}
{{- end -}}

{{/*
Determine the PostgreSQL schema based on whether the subchart is enabled.
*/}}
{{- define "pocket.postgresqlDatabase" -}}
{{- if .Values.postgresql.enabled -}}
{{- "postgres" -}}
{{- else -}}
{{- .Values.externalPostgresql.database -}}
{{- end -}}
{{- end -}}

{{/*
Determine the PostgreSQL schema based on whether the subchart is enabled.
*/}}
{{- define "pocket.postgresqlSchema" -}}
{{- .Values.config.persistence.node_schema -}}
{{- end -}}

{{/*
Determine the PostgreSQL user SecretKeyRef based on whether the subchart is enabled.
*/}}
{{- define "pocket.postgresqlUserValueOrSecretRef" -}}
{{- if .Values.postgresql.enabled -}}
value: postgres
{{- else -}}
valueFrom:
  secretKeyRef:
    name: {{ .Values.externalPostgresql.userSecretKeyRef.name }}
    key: {{ .Values.externalPostgresql.userSecretKeyRef.key }}
{{- end -}}
{{- end -}}

{{/*
Determine the PostgreSQL password SecretKeyRef based on whether the subchart is enabled.
*/}}
{{- define "pocket.postgresqlPasswordSecretKeyRef" -}}
{{- if .Values.postgresql.enabled -}}
valueFrom:
  secretKeyRef:
    name: {{ printf "%s-%s" .Release.Name "postgresql" }}
    key: postgres-password
{{- else -}}
valueFrom:
  secretKeyRef:
    name: {{ .Values.externalPostgresql.passwordSecretKeyRef.name }}
    key: {{ .Values.externalPostgresql.passwordSecretKeyRef.key }}
{{- end -}}
{{- end -}}

{{/*
Determine the genesis ConfigMap based on whether pre-provisioned genesis is enabled.
*/}}
{{- define "pocket.genesisConfigMap" -}}
{{- if .Values.genesis.preProvisionedGenesis.enabled -}}
{{- printf "%s-%s" (include "pocket.fullname" .) "genesis" -}}
{{- else -}}
{{- .Values.genesis.externalConfigMap.name -}}
{{- end -}}
{{- end -}}
