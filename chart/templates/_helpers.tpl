{{- define "utils.maps.merge" -}}
{{- $workMap := dict -}}
{{- if kindIs "slice" . -}}
{{- range . -}}
{{- $workMap = merge $workMap . -}}
{{- end -}}
{{- end -}}
{{- if gt (len $workMap) 0 -}}
{{ toYaml $workMap }}
{{- end -}}
{{- end -}}
