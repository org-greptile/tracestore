{
  local k = import 'ksonnet-util/kausal.libsonnet',
  local configMap = k.core.v1.configMap,

  tracestore_config:: {
    http_api_prefix: $._config.http_api_prefix,

    server: {
      http_listen_port: $._config.port,
    },
    distributor: {},
    ingester: {
      lifecycler: {
        ring: {
          replication_factor: 3,
        },
      },
    },
    compactor: {},
    storage: {
      trace: {
        blocklist_poll: '0',
        backend: $._config.backend,
        wal: {
          path: '/var/tracestore/wal',
        },
        gcs: {
          bucket_name: $._config.bucket,
          chunk_buffer_size: 10485760,  // 1024 * 1024 * 10
        },
        s3: {
          bucket: $._config.bucket,
        },
        azure: {
          container_name: $._config.bucket,
        },
        pool: {
          queue_depth: 2000,
        },
        cache: 'memcached',
        memcached: {
          consistent_hash: true,
          timeout: '200ms',
          host: 'memcached',
          service: 'memcached-client',
        },
      },
    },
    overrides: {
      per_tenant_override_config: '/overrides/overrides.yaml',
    },
    memberlist: {
      abort_if_cluster_join_fails: false,
      bind_port: $._config.gossip_ring_port,
      join_members: ['gossip-ring.%s.svc.cluster.local.:%d' % [$._config.namespace, $._config.gossip_ring_port]],
    },
  },

  tracestore_distributor_config:: $.tracestore_config {
    distributor+: {
      receivers+: $._config.distributor.receivers,
    },
  },

  tracestore_ingester_config:: $.tracestore_config {},

  tracestore_metrics_generator_config:: $.tracestore_config {
    metrics_generator+: {
      storage+: {
        path: '/var/tracestore/generator_wal',
      },
    },
  },

  tracestore_compactor_config:: $.tracestore_config {
    compactor+: {
      compaction+: {
        v2_in_buffer_bytes: 10485760,
        block_retention: '144h',
      },
      ring+: {
        kvstore+: {
          store: 'memberlist',
        },
      },
    },
    storage+: {
      trace+: {
        blocklist_poll: '5m',
      },
    },
  },

  tracestore_querier_config:: $.tracestore_config {
    server+: {
      log_level: 'debug',
    },
    storage+: {
      trace+: {
        blocklist_poll: '5m',
        pool+: {
          max_workers: 200,
        },
      },
    },
    querier+: {
      frontend_worker+: {
        frontend_address: 'query-frontend-discovery.%s.svc.cluster.local.:9095' % [$._config.namespace],
      },
    },
  },

  tracestore_query_frontend_config:: $.tracestore_config {},

  // This will be the single configmap that stores `overrides.yaml`.
  overrides_config:
    configMap.new($._config.overrides_configmap_name) +
    configMap.withData({
      'overrides.yaml': k.util.manifestYaml({
        overrides: $._config.overrides,
      }),
    }),

  tracestore_distributor_configmap:
    configMap.new('tracestore-distributor') +
    configMap.withData({
      'tracestore.yaml': k.util.manifestYaml($.tracestore_distributor_config),
    }),

  tracestore_ingester_configmap:
    configMap.new('tracestore-ingester') +
    configMap.withData({
      'tracestore.yaml': k.util.manifestYaml($.tracestore_ingester_config),
    }),

  tracestore_metrics_generator_configmap:
    configMap.new('tracestore-metrics-generator') +
    configMap.withData({
      'tracestore.yaml': $.util.manifestYaml($.tracestore_metrics_generator_config),
    }),

  tracestore_compactor_configmap:
    configMap.new('tracestore-compactor') +
    configMap.withData({
      'tracestore.yaml': k.util.manifestYaml($.tracestore_compactor_config),
    }),

  tracestore_querier_configmap:
    configMap.new('tracestore-querier') +
    configMap.withData({
      'tracestore.yaml': $.util.manifestYaml($.tracestore_querier_config),
    }),

  tracestore_query_frontend_configmap:
    configMap.new('tracestore-query-frontend') +
    configMap.withData({
      'tracestore.yaml': $.util.manifestYaml($.tracestore_query_frontend_config),
    }),

  tracestore_query_configmap:
    configMap.new('tracestore-query') +
    configMap.withData({
      'tracestore-query.yaml': $.util.manifestYaml({
        backend: 'localhost:%d%s' % [$._config.port, $._config.http_api_prefix],
      }),
    }),
}
