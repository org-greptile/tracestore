// The jsonnet file used to generate the Kubernetes manifests.

local tracestore = import 'microservices/tracestore.libsonnet';

tracestore {
  _images+:: {
    tracestore: 'acme/tracestore:latest',
    tracestore_vulture: 'acme/tracestore-vulture:latest',
    tracestore_query: 'acme/tracestore-query:latest',
  },

  // generate with `tracestore_query.enabled: true` to include tracestore-query manifests
  _config+:: {
    namespace: 'tracing',
    compactor+: {
      replicas: 5,
    },
    query_frontend+: {
      replicas: 2,
    },
    querier+: {
      replicas: 5,
    },
    ingester+: {
      replicas: 10,
      pvc_size: '10Gi',
      pvc_storage_class: 'fast',
    },
    distributor+: {
      replicas: 5,
      receivers: {
        jaeger: {
          protocols: {
            grpc: {
              endpoint: '0.0.0.0:14250',
            },
          },
        },
        otlp: {
          protocols: {
            grpc: {
              endpoint: '0.0.0.0:4317',
            },
          },
        },
      },
    },
    metrics_generator+: {
      pvc_size: '10Gi',
      pvc_storage_class: 'fast',
      ephemeral_storage_request_size: '10Gi',
      ephemeral_storage_limit_size: '11Gi',
    },
    memcached+: {
      replicas: 5,
    },
    vulture+: {
      replicas: 1,
      tracestoreOrgId: '1',
      tracestorePushUrl: 'http://distributor',
      tracestoreQueryUrl: 'http://query-frontend:3200/tracestore',
    },
    jaeger_ui: {
      base_path: '/tracestore',
    },
    backend: 'gcs',
    bucket: 'tracestore',
  },

  local statefulSet = $.apps.v1.statefulSet,
  tracestore_ingester_statefulset:
    if !$._config.multi_zone_ingester_enabled then super.tracestore_ingester_statefulset + statefulSet.mixin.spec.withPodManagementPolicy('Parallel') else null,

}
