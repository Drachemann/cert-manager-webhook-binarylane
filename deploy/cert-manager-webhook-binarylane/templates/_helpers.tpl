{{- define "cert-manager-webhook-binarylane.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- define "cert-manager-webhook-binarylane.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "cert-manager-webhook-binarylane.name" .) | trunc 63 | trimSuffix "-" }}
{{- end }}
