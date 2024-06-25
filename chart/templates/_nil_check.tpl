{{- define "fleetscheduler.scheduleValueExists" -}}
  {{- if .job.schedule -}}
    {{- if index .job.schedule (printf "%s" .var) -}}
      {{- index .job.schedule (printf "%s" .var) -}}
    {{- else -}}
      {{- .default -}}
    {{- end -}}
  {{- else -}}
    {{- .default -}}
  {{- end -}}
{{- end -}}