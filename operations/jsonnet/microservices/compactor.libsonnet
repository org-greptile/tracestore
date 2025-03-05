{
  local k = import 'ksonnet-util/kausal.libsonnet',
  local container = k.core.v1.container,
  local containerPort = k.core.v1.containerPort,
  local volumeMount = k.core.v1.volumeMount,
  local deployment = k.apps.v1.deployment,
  local volume = k.core.v1.volume,
  local envVar = k.core.v1.envVar,

  local target_name = 'compactor',
  local tracestore_config_volume = 'tracestore-conf',
  local tracestore_data_volume = 'tracestore-data',
  local tracestore_overrides_config_volume = 'overrides',

  tracestore_compactor_ports:: [containerPort.new('prom-metrics', $._config.port)],
  tracestore_compactor_args:: {
    target: target_name,
    'config.file': '/conf/tracestore.yaml',
    'mem-ballast-size-mbs': $._config.ballast_size_mbs,
  },

  tracestore_compactor_container::
    container.new(target_name, $._images.tracestore) +
    container.withPorts($.tracestore_compactor_ports) +
    container.withArgs($.util.mapToFlags($.tracestore_compactor_args)) +
    (if $._config.variables_expansion then container.withEnvMixin($._config.variables_expansion_env_mixin) else {}) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_config_volume, '/conf'),
      volumeMount.new(tracestore_overrides_config_volume, '/overrides'),
    ]) +
    $.util.withResources($._config.compactor.resources) +
    $.util.readinessProbe +
    (if $._config.variables_expansion then container.withArgsMixin(['-config.expand-env=true']) else {}) +
    (if $._config.compactor.resources.limits.memory != null then container.withEnvMixin([envVar.new('GOMEMLIMIT', $._config.compactor.resources.limits.memory + 'B')]) else {}),

  tracestore_compactor_deployment:
    deployment.new(target_name,
                   $._config.compactor.replicas,
                   [
                     $.tracestore_compactor_container,
                   ],
                   { app: target_name }) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge('50%') +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable('100%') +
    deployment.mixin.spec.template.metadata.withAnnotations({
      config_hash: std.md5(std.toString($.tracestore_compactor_configmap.data['tracestore.yaml'])),
    }) +
    deployment.mixin.spec.template.spec.withVolumes([
      volume.fromConfigMap(tracestore_config_volume, $.tracestore_compactor_configmap.metadata.name),
      volume.fromConfigMap(tracestore_overrides_config_volume, $._config.overrides_configmap_name),
    ]),

  tracestore_compactor_service:
    k.util.serviceFor($.tracestore_compactor_deployment),
}
