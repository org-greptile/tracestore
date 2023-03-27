{
  local k = import 'ksonnet-util/kausal.libsonnet',
  local container = k.core.v1.container,
  local containerPort = k.core.v1.containerPort,
  local volumeMount = k.core.v1.volumeMount,
  local deployment = k.apps.v1.deployment,
  local volume = k.core.v1.volume,
  local service = k.core.v1.service,

  local target_name = 'distributor',
  local tracestore_config_volume = 'tracestore-conf',
  local tracestore_overrides_config_volume = 'overrides',

  tracestore_distributor_ports:: [containerPort.new('prom-metrics', $._config.port)],
  tracestore_distributor_args:: {
    target: target_name,
    'config.file': '/conf/tracestore.yaml',
    'mem-ballast-size-mbs': $._config.ballast_size_mbs,
  },

  tracestore_distributor_container::
    container.new(target_name, $._images.tracestore) +
    container.withPorts($.tracestore_distributor_ports) +
    container.withArgs($.util.mapToFlags($.tracestore_distributor_args)) +
    (if $._config.variables_expansion then container.withEnvMixin($._config.variables_expansion_env_mixin) else {}) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_config_volume, '/conf'),
      volumeMount.new(tracestore_overrides_config_volume, '/overrides'),
    ]) +
    $.util.withResources($._config.distributor.resources) +
    $.util.readinessProbe +
    (if $._config.variables_expansion then container.withArgsMixin(['-config.expand-env=true']) else {}),

  tracestore_distributor_deployment:
    deployment.new(target_name,
                   $._config.distributor.replicas,
                   [
                     $.tracestore_distributor_container,
                   ],
                   {
                     app: target_name,
                     [$._config.gossip_member_label]: 'true',
                   }) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(3) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1) +
    deployment.mixin.spec.template.spec.withTerminationGracePeriodSeconds(60) +
    deployment.mixin.spec.template.metadata.withAnnotations({
      config_hash: std.md5(std.toString($.tracestore_distributor_configmap.data['tracestore.yaml'])),
    }) +
    deployment.mixin.spec.template.spec.withVolumes([
      volume.fromConfigMap(tracestore_config_volume, $.tracestore_distributor_configmap.metadata.name),
      volume.fromConfigMap(tracestore_overrides_config_volume, $._config.overrides_configmap_name),
    ]),

  tracestore_distributor_service:
    k.util.serviceFor($.tracestore_distributor_deployment),

  ingest_service:
    k.util.serviceFor($.tracestore_distributor_deployment)
    + service.mixin.metadata.withName('ingest')
    + service.mixin.spec.withClusterIp('None'),
}
