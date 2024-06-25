  {{/*
    # uint64 max value is 20 digits. So 19 should be good to get a good big value for a random number with (randNumberic 19)

    # Psuedo code to show what the template logic is doing below
    # if enable_random_minute {
    #     if (random_minute_span + minute) > 59 {
    #         return error("random_minute_span and minute must not add up to more than 59")
    #     }
    #     if !minute.isNumber() {
    #         return error("if enable_random_minute is true, minute must be a single number")
    #     }
    #     if !random_minute_span.isNumber() {
    #         return error("if enable_random_minute is true, random_minute_span must be a single number")
    #     }
    #     if random_minute_span == 0 {
    #         return error("random_minute_span must be greater than 0")
    #     }
    # } else if job.schedule.minute > 59 {
    #     return error("minute must not be greater than 59")
    # }
    # calculated_minute = enable_random_minute ? minute + rand_uint64() % (random_minute_span + 1) : minute
    # return calculated_minute

    # .type: literal string of minute, hour, or day_of_week
    # .var: either minute, hour, or day_of_week
    # .span: either random_minute_span, random_hour_span, or random_day_of_week_span
    # .limit: either 59 (for minutes), 23 (for hours), or 6 (for days_of_week)
    # .enable: either enable_random_minute, enable_random_hour, enable_random_days_of_week
  */}}

{{- define "fleetscheduler.calculate" -}}
  {{- if .enable -}}
    {{- if gt (add .span .var) .limit -}}
      {{- printf "if enabled, random_%s_span and %s must not add up to more than %d (was: %d)" .type .type .limit (add .span .var) | fail -}}
    {{- end -}}
    {{- if not (regexMatch "[0-9]+" .var) -}}
      {{- printf "if enabled, %s must be a single number (was: %s)" .type .var | fail -}}
    {{- end -}}
    {{- if not (regexMatch "[0-9]+" .span) -}}
      {{- printf "if enabled, random_%s_span must be a single number (was: %s)" .type .span | fail -}}
    {{- end -}}
    {{- if eq (atoi .span) 0 -}}
      {{- printf "if enabled, random_%s_span must be greater than 0 (was: %s)" .type .span | fail -}}
    {{- end}}
  {{- else if gt (atoi .var) 59 -}}
      {{- fail "%s must not be greater than %d" -}}
      {{- printf "%s must not be greater than %d (was: %s)" .type .limit .var | fail -}}
  {{- end -}}
  {{- $calculated := ternary
    (add .var (mod (randNumeric 19) (add1 .span)))
    .var
    (eq .enable "true")
  -}}
  {{- $calculated -}}
{{- end -}}