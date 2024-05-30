local utils = import 'mixin-utils/utils.libsonnet';

{
  prometheusRules+:: {
    groups+: [{
      name: 'tracestore_rules',
      rules:
        utils.histogramRules('tracestore_request_duration_seconds', $._config.job_selectors + ['route']),
    }],
  },
}
