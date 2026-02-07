{{/*
Expand the name of the chart.
*/}}
{{- define "helm.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "helm.fullname" -}}
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
{{- define "helm.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "helm.labels" -}}
helm.sh/chart: {{ include "helm.chart" . }}
{{ include "helm.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "helm.selectorLabels" -}}
app.kubernetes.io/name: {{ include "helm.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "helm.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "helm.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the proper image name with global registry support
Usage: {{ include "helm.image" (dict "imageConfig" .Values.images.redis "global" .Values.global) }}
*/}}
{{- define "helm.image" -}}
{{- $registry := .global.imageRegistry -}}
{{- $name := .imageConfig.repository | default .imageConfig.name -}}
{{- $tag := .imageConfig.tag | default "latest" | toString -}}
{{- $imageName := "" -}}
{{- if $registry -}}
  {{- $imageName = printf "%s/%s" $registry $name -}}
{{- else -}}
  {{- $imageName = $name -}}
{{- end -}}
{{- if contains "@" $tag -}}
  {{- printf "%s%s" $imageName $tag -}}
{{- else -}}
  {{- printf "%s:%s" $imageName $tag -}}
{{- end -}}
{{- end -}}

{{/*
Get etcd config with user overrides
Usage: {{ include "helm.etcd.mergedConfig" (dict "root" . "configType" "producer") }}
Parameters:
  - root: root context (.)
  - configType: "producer" or "consumer"
Returns: merged JSON configuration
*/}}
{{- define "helm.etcd.mergedConfig" -}}
{{- $dataJsonContent := "" -}}
{{- if .root.Values.initialDataFiles.data_yaml -}}
  {{- $dataJsonContent = .root.Values.initialDataFiles.data_yaml -}}
{{- else -}}
  {{- $dataJsonContent = .root.Files.Get "files/initial_data/data.yaml" -}}
{{- end -}}
{{- $dataJson := $dataJsonContent | fromYaml -}}
{{- $dynamicConfigs := $dataJson.dynamic_configs | default list -}}

{{/* Build default config from dynamic_configs based on scope */}}
{{- $defaultConfig := dict -}}
{{- $targetScope := 1 -}}
{{- if eq .configType "producer" -}}
  {{- $targetScope = 0 -}}
{{- end -}}

{{- range $dynamicConfigs -}}
  {{- if eq (int .scope) $targetScope -}}
    {{- $_ := set $defaultConfig .key .default_value -}}
  {{- end -}}
{{- end -}}

{{/* User overrides from values.yaml */}}
{{- $userConfig := dict -}}
{{- if eq .configType "producer" -}}
  {{- $userConfig = .root.Values.initialConfig.producer.settings | default dict -}}
{{- else if eq .configType "consumer" -}}
  {{- $userConfig = .root.Values.initialConfig.consumer.settings | default dict -}}
{{- end -}}

{{/* Merge: default -> userConfig */}}
{{- $merged := mergeOverwrite (dict) $defaultConfig $userConfig -}}
{{- $merged | toYaml -}}
{{- end -}}

{{/*
Generate system helm config list from data.yaml containers (type=2 only)
Usage: {{ include "helm.systemHelmConfigs" . }}
Returns: JSON array of system helm configurations
*/}}
{{- define "helm.systemHelmConfigs" -}}
{{- $dataJsonContent := "" -}}
{{- if .Values.initialDataFiles.data_yaml -}}
  {{- $dataJsonContent = .Values.initialDataFiles.data_yaml -}}
{{- else -}}
  {{- $dataJsonContent = .Files.Get "files/initial_data/data.yaml" -}}
{{- end -}}
{{- $dataJson := $dataJsonContent | fromYaml -}}
{{- $containers := $dataJson.containers | default list -}}

{{- $systemConfigs := list -}}
{{- range $index, $container := $containers -}}
  {{- if eq (int $container.type) 2 -}}
    {{- $containerVersionId := add $index 1 -}}
    {{- $system := $container.name -}}
    {{- range $container.versions -}}
      {{- if .helm_config -}}
        {{- $config := dict 
          "chart_name" .helm_config.chart_name
          "container_version_id" $containerVersionId
          "system" $system
          "version" .helm_config.version
        -}}
        {{- $systemConfigs = append $systemConfigs $config -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- $systemConfigs | toJson -}}
{{- end -}}
