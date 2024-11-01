{
  local k = import 'ksonnet-util/kausal.libsonnet',
  local container = k.core.v1.container,
  local containerPort = k.core.v1.containerPort,
  local volumeMount = k.core.v1.volumeMount,
  local deployment = k.apps.v1.deployment,
  local volume = k.core.v1.volume,
  local service = k.core.v1.service,
  local servicePort = k.core.v1.servicePort,

  local target_name = 'query-frontend',
  local tracestore_config_volume = 'tracestore-conf',
  local tracestore_query_config_volume = 'tracestore-query-conf',
  local tracestore_data_volume = 'tracestore-data',
  local tracestore_overrides_config_volume = 'overrides',

  tracestore_query_frontend_ports:: [containerPort.new('prom-metrics', $._config.port)],
  tracestore_query_frontend_args:: {
    target: target_name,
    'config.file': '/conf/tracestore.yaml',
    'mem-ballast-size-mbs': $._config.ballast_size_mbs,
  },

  tracestore_query_frontend_container::
    container.new(target_name, $._images.tracestore) +
    container.withPorts($.tracestore_query_frontend_ports) +
    container.withArgs($.util.mapToFlags($.tracestore_query_frontend_args)) +
    (if $._config.variables_expansion then container.withEnvMixin($._config.variables_expansion_env_mixin) else {}) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_config_volume, '/conf'),
      volumeMount.new(tracestore_overrides_config_volume, '/overrides'),
    ]) +
    $.util.withResources($._config.query_frontend.resources) +
    $.util.readinessProbe +
    (if $._config.variables_expansion then container.withArgsMixin(['-config.expand-env=true']) else {}),

  tracestore_query_container::
    container.new('tracestore-query', $._images.tracestore_query) +
    container.withPorts([
      containerPort.new('jaeger-ui', 16686),
      containerPort.new('jaeger-metrics', 16687),
    ]) +
    container.withArgs([
      '--query.base-path=' + $._config.jaeger_ui.base_path,
      '--grpc-storage-plugin.configuration-file=/conf/tracestore-query.yaml',
      '--query.bearer-token-propagation=true',
    ]) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_query_config_volume, '/conf'),
    ]),

  tracestore_query_frontend_deployment:
    deployment.new(
      target_name,
      $._config.query_frontend.replicas,
      std.prune([
        $.tracestore_query_frontend_container,
        if $._config.tracestore_query.enabled then $.tracestore_query_container,
      ]),
      {
        app: target_name,
      }
    ) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(0) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1) +
    deployment.mixin.spec.template.metadata.withAnnotations({
      config_hash: std.md5(std.toString($.tracestore_query_frontend_configmap.data['tracestore.yaml'])),
    }) +
    deployment.mixin.spec.template.spec.withVolumes(std.prune([
      if $._config.tracestore_query.enabled then volume.fromConfigMap(tracestore_query_config_volume, $.tracestore_query_configmap.metadata.name),
      volume.fromConfigMap(tracestore_config_volume, $.tracestore_query_frontend_configmap.metadata.name),
      volume.fromConfigMap(tracestore_overrides_config_volume, $._config.overrides_configmap_name),
    ])),

  tracestore_query_frontend_service:
    k.util.serviceFor($.tracestore_query_frontend_deployment)
    + service.mixin.spec.withPortsMixin([
      servicePort.withName('http')
      + servicePort.withPort(80)
      + servicePort.withTargetPort($._config.port),
    ]),

  tracestore_query_frontend_discovery_service:
    k.util.serviceFor($.tracestore_query_frontend_deployment)
    + service.mixin.spec.withPortsMixin([
      servicePort.withName('grpc')
      + servicePort.withPort(9095)
      + servicePort.withTargetPort(9095),
    ])
    + service.mixin.spec.withPublishNotReadyAddresses(true)
    + service.mixin.spec.withClusterIp('None')
    + service.mixin.metadata.withName('query-frontend-discovery'),
}
