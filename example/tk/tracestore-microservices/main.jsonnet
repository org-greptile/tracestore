local tracestore = import '../../../operations/jsonnet/microservices/tracestore.libsonnet';
local dashboards = import 'dashboards/acme.libsonnet';
local metrics = import 'metrics/prometheus.libsonnet';
local minio = import 'minio/minio.libsonnet';
local load = import 'synthetic-load-generator/main.libsonnet';

minio + metrics + load + tracestore {

  dashboards:
    dashboards.deploy(),

  _images+:: {
    // images can be overridden here if desired
  },

  _config+:: {
    cluster: 'k3d',
    namespace: 'default',
    compactor+: {
    },
    querier+: {
    },
    ingester+: {
      pvc_size: '5Gi',
      pvc_storage_class: 'local-path',
    },
    distributor+: {
      receivers: {
        opencensus: null,
        jaeger: {
          protocols: {
            thrift_http: null,
          },
        },
      },
    },
    metrics_generator+: {
      ephemeral_storage_limit_size: '2Gi',
      ephemeral_storage_request_size: '1Gi',
    },
    memcached+: {
      replicas: 1,
    },
    vulture+: {
      replicas: 0,
      // Disable search until release
      tracestoreSearchBackoffDuration: '0s',
    },
    backend: 's3',
    bucket: 'tracestore',
    tracestore_query_url: 'http://query-frontend:3200',
  },

  // manually overriding to get tracestore to talk to minio
  tracestore_config+:: {
    storage+: {
      trace+: {
        s3+: {
          endpoint: 'minio:9000',
          access_key: 'tracestore',
          secret_key: 'supersecret',
          insecure: true,
        },
      },
    },
  },

  local k = import 'ksonnet-util/kausal.libsonnet',
  local service = k.core.v1.service,
  tracestore_service:
    k.util.serviceFor($.tracestore_distributor_deployment)
    + service.mixin.metadata.withName('tracestore'),

  local container = k.core.v1.container,
  local containerPort = k.core.v1.containerPort,
  tracestore_compactor_container+::
    k.util.resourcesRequests('500m', '500Mi'),

  tracestore_distributor_container+::
    k.util.resourcesRequests('500m', '500Mi') +
    container.withPortsMixin([
      containerPort.new('opencensus', 55678),
      containerPort.new('jaeger-http', 14268),
    ]),

  tracestore_ingester_container+::
    k.util.resourcesRequests('500m', '500Mi'),

  // clear affinity so we can run multiple ingesters on a single node
  tracestore_ingester_statefulset+: {
    spec+: {
      template+: {
        spec+: {
          affinity: {},
        },
      },
    },
  },

  tracestore_querier_container+::
    k.util.resourcesRequests('500m', '500Mi'),

  tracestore_query_frontend_container+::
    k.util.resourcesRequests('300m', '500Mi'),

  // clear affinity so we can run multiple instances of memcached on a single node
  memcached_all+: {
    statefulSet+: {
      spec+: {
        template+: {
          spec+: {
            affinity: {},
          },
        },
      },
    },
  },

  local ingress = k.networking.v1.ingress,
  local rule = k.networking.v1.ingressRule,
  local path = k.networking.v1.httpIngressPath,
  ingress:
    ingress.new('ingress') +
    ingress.mixin.metadata
    .withAnnotationsMixin({
      'ingress.kubernetes.io/ssl-redirect': 'false',
    }) +
    ingress.mixin.spec.withRules(
      rule.http.withPaths([
        path.withPath('/')
        + path.withPathType('ImplementationSpecific')
        + path.backend.service.withName('acme')
        + path.backend.service.port.withNumber(3000),
      ]),
    ),
}
