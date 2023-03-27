{
  local k = import 'ksonnet-util/kausal.libsonnet',

  local container = k.core.v1.container,
  local containerPort = k.core.v1.containerPort,
  local deployment = k.apps.v1.deployment,
  local volume = k.core.v1.volume,
  local volumeMount = k.core.v1.volumeMount,

  local target_name = 'metrics-generator',
  local tracestore_config_volume = 'tracestore-conf',
  local tracestore_generator_wal_volume = 'metrics-generator-wal-data',
  local tracestore_overrides_config_volume = 'overrides',

  tracestore_metrics_generator_ports:: [containerPort.new('prom-metrics', $._config.port)],
  tracestore_metrics_generator_args:: {
    target: target_name,
    'config.file': '/conf/tracestore.yaml',
    'mem-ballast-size-mbs': $._config.ballast_size_mbs,
  },

  tracestore_metrics_generator_container::
    container.new(target_name, $._images.tracestore) +
    container.withPorts($.tracestore_metrics_generator_ports) +
    container.withArgs($.util.mapToFlags($.tracestore_metrics_generator_args)) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_config_volume, '/conf'),
      volumeMount.new(tracestore_generator_wal_volume, $.tracestore_metrics_generator_config.metrics_generator.storage.path),
      volumeMount.new(tracestore_overrides_config_volume, '/overrides'),
    ]) +
    $.util.withResources($._config.metrics_generator.resources) +
    (if $._config.variables_expansion then container.withEnvMixin($._config.variables_expansion_env_mixin) else {}) +
    container.mixin.resources.withRequestsMixin({ 'ephemeral-storage': $._config.metrics_generator.ephemeral_storage_request_size }) +
    container.mixin.resources.withLimitsMixin({ 'ephemeral-storage': $._config.metrics_generator.ephemeral_storage_limit_size }) +
    $.util.readinessProbe +
    (if $._config.variables_expansion then container.withArgsMixin(['-config.expand-env=true']) else {}),

  tracestore_metrics_generator_deployment:
    deployment.new(
      target_name,
      $._config.metrics_generator.replicas,
      $.tracestore_metrics_generator_container,
      {
        app: target_name,
        [$._config.gossip_member_label]: 'true',
      },
    ) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(3) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1) +
    deployment.mixin.spec.template.metadata.withAnnotations({
      config_hash: std.md5(std.toString($.tracestore_metrics_generator_configmap.data['tracestore.yaml'])),
    }) +
    deployment.mixin.spec.template.spec.withVolumes([
      volume.fromConfigMap(tracestore_config_volume, $.tracestore_metrics_generator_configmap.metadata.name),
      volume.fromConfigMap(tracestore_overrides_config_volume, $._config.overrides_configmap_name),
      volume.fromEmptyDir(tracestore_generator_wal_volume),
    ]),

  tracestore_metrics_generator_service:
    k.util.serviceFor($.tracestore_metrics_generator_deployment),
}
