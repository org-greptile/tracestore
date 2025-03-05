local dashboard_utils = import 'dashboard-utils.libsonnet';
local g = import 'acme-builder/acme.libsonnet';

dashboard_utils {
  acmeDashboards+: {
    'tracestore-reads.json':
      $.dashboard('Tracestore / Reads')
      .addClusterSelectorTemplates()
      .addRow(
        g.row('Gateway')
        .addPanel(
          $.panel('QPS') +
          $.qpsPanel('tracestore_request_duration_seconds_count{%s, route=~"%sapi_.*"}' % [$.jobMatcher($._config.jobs.gateway), $._config.http_api_prefix])
        )
        .addPanel(
          $.panel('Latency') +
          $.latencyPanel('tracestore_request_duration_seconds', '{%s,route=~"%sapi_.*"}' % [$.jobMatcher($._config.jobs.gateway), $._config.http_api_prefix], additional_grouping='route')
        )
      )
      .addRow(
        g.row('Query Frontend')
        .addPanel(
          $.panel('QPS') +
          $.qpsPanel('tracestore_request_duration_seconds_count{%s, route=~"%sapi_.*"}' % [$.jobMatcher($._config.jobs.query_frontend), $._config.http_api_prefix])
        )
        .addPanel(
          $.panel('Latency') +
          $.latencyPanel('tracestore_request_duration_seconds', '{%s,route=~"%sapi_.*"}' % [$.jobMatcher($._config.jobs.query_frontend), $._config.http_api_prefix], additional_grouping='route')
        )
      )
      .addRow(
        g.row('Querier')
        .addPanel(
          $.panel('QPS') +
          $.qpsPanel('tracestore_request_duration_seconds_count{%s, route=~"querier_%sapi_.*"}' % [$.jobMatcher($._config.jobs.querier), $._config.http_api_prefix])
        )
        .addPanel(
          $.panel('Latency') +
          $.latencyPanel('tracestore_request_duration_seconds', '{%s,route=~"querier_%sapi_.*"}' % [$.jobMatcher($._config.jobs.querier), $._config.http_api_prefix], additional_grouping='route')
        )
      )
      .addRow(
        g.row('Ingester')
        .addPanel(
          $.panel('QPS') +
          $.qpsPanel('tracestore_request_duration_seconds_count{%s, route=~"/tracestorepb.Querier/.*"}' % $.jobMatcher($._config.jobs.ingester))
        )
        .addPanel(
          $.panel('Latency') +
          $.latencyPanel('tracestore_request_duration_seconds', '{%s,route=~"/tracestorepb.Querier/.*"}' % $.jobMatcher($._config.jobs.ingester), additional_grouping='route')
        )
      )
      .addRow(
        g.row('Memcached')
        .addPanel(
          $.panel('QPS') +
          $.qpsPanel('tracestore_memcache_request_duration_seconds_count{%s,method=~"Memcache.Get|Memcache.GetMulti"}' % $.jobMatcher($._config.jobs.querier))
        )
        .addPanel(
          $.panel('Latency') +
          $.latencyPanel('tracestore_memcache_request_duration_seconds', '{%s,method=~"Memcache.Get|Memcache.GetMulti"}' % $.jobMatcher($._config.jobs.querier))
        )
      )
      .addRow(
        g.row('Backend')
        .addPanel(
          $.panel('QPS') +
          $.qpsPanel('tracestoredb_backend_request_duration_seconds_count{%s,operation="GET"}' % $.jobMatcher($._config.jobs.querier))
        )
        .addPanel(
          $.panel('Latency') +
          $.latencyPanel('tracestoredb_backend_request_duration_seconds', '{%s,operation="GET"}' % $.jobMatcher($._config.jobs.querier))
        )
      ),
  },
}
