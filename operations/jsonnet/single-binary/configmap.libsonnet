{
  local k = import 'ksonnet-util/kausal.libsonnet',
  local configMap = k.core.v1.configMap,

  tracestore_config:: {
    server: {
      http_listen_port: $._config.tracestore.port,
    },
    distributor: {
      receivers: $._config.receivers,
    },
    ingester: {
    },
    compactor: {
      compaction: {
        block_retention: '24h',
      },
    },
    memberlist: {
      abort_if_cluster_join_fails: false,
      bind_port: 7946,
      join_members: [
        '%s:7946' % $._config.tracestore.headless_service_name,
      ],
    },
    storage: {
      trace: {
        backend: 'local',
        wal: {
          path: '/var/tracestore/wal',
        },
        'local': {
          path: '/tmp/tracestore/traces',
        },
      },
    },
    querier: {
      frontend_worker: {
        frontend_address: 'tracestore:9095',
      },
    },
  },

  tracestore_configmap:
    configMap.new('tracestore') +
    configMap.withData({
      'tracestore.yaml': k.util.manifestYaml($.tracestore_config),
    }) +
    configMap.withDataMixin({
      'overrides.yaml': |||
        overrides:
      |||,
    }),

  tracestore_query_configmap:
    configMap.new('tracestore-query') +
    configMap.withData({
      'tracestore-query.yaml': k.util.manifestYaml({
        backend: 'localhost:%d' % $._config.tracestore.port,
      }),
    }),
}
