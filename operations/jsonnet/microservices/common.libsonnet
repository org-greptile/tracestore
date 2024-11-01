{

  local k = import 'k.libsonnet',
  local kausal = import 'ksonnet-util/kausal.libsonnet',

  local container = k.core.v1.container,
  local volumeMount = k.core.v1.volumeMount,
  local statefulset = k.apps.v1.statefulSet,

  tracestore_chown_container(data_volume, uid='10001', gid='10001', dir='/var/tracestore')::
    container.new('chown-' + data_volume, $._images.tracestore) +
    container.withCommand('chown') +
    container.withArgs([
      '-R',
      '%s:%s' % [uid, gid],
      dir,
    ]) +
    container.withVolumeMounts([
      volumeMount.new(data_volume, dir),
    ]) +
    container.securityContext.withRunAsUser(0) +
    container.securityContext.withRunAsGroup(0) +
    {},

  util+:: {
    local k = import 'ksonnet-util/kausal.libsonnet',
    local container = k.core.v1.container,
    local service = k.core.v1.service,

    withResources(resources)::
      k.util.resourcesRequests(resources.requests.cpu, resources.requests.memory) +
      k.util.resourcesLimits(resources.limits.cpu, resources.limits.memory),

    readinessProbe::
      container.mixin.readinessProbe.httpGet.withPath('/ready') +
      container.mixin.readinessProbe.httpGet.withPort($._config.port) +
      container.mixin.readinessProbe.withInitialDelaySeconds(15) +
      container.mixin.readinessProbe.withTimeoutSeconds(1),

    withInet6():: {
      tracestore_compactor_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      tracestore_distributor_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      tracestore_ingester_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      tracestore_querier_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      tracestore_query_frontend_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      tracestore_query_frontend_discovery_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      tracestore_metrics_generator_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      gossip_ring_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      ingest_service+:
        service.mixin.spec.withIpFamilies(['IPv6']),
      memcached+: {
        service+:
          service.mixin.spec.withIpFamilies(['IPv6']),
      },
      tracestore_config+:: {
        server+: {
          http_listen_address: '::0',
          grpc_listen_address: '::0',
        },
        ingester+: {
          lifecycler+: {
            enable_inet6: true,
          },
        },
        memberlist+: {
          bind_addr: ['::'],
        },
        compactor+: {
          ring+: {
            enable_inet6: true,
          },
        },
        metrics_generator+: {
          ring+: {
            enable_inet6: true,
          },
        },
      },
    },

    withInitChown():: {
      tracestore_metrics_generator_statefulset+:
        statefulset.spec.template.spec.withInitContainers([
          $.tracestore_chown_container('metrics-generator-data'),
        ]),

      tracestore_ingester_statefulset+:
        statefulset.spec.template.spec.withInitContainers([
          $.tracestore_chown_container('ingester-data'),
        ]),
    },
  },
}
