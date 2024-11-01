(import 'dashboards/tracestore-operational.libsonnet') +
(import 'dashboards/tracestore-reads.libsonnet') +
(import 'dashboards/tracestore-resources.libsonnet') +
(import 'dashboards/tracestore-tenants.libsonnet') +
(import 'dashboards/tracestore-writes.libsonnet') +
{
  acmeDashboards+:
    (import 'dashboards/rollout-progress.libsonnet') +

    { _config:: $._config },
}
