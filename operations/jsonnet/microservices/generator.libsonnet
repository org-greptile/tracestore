{
  local k = import 'k.libsonnet',
  local kausal = import 'ksonnet-util/kausal.libsonnet',

  local container = k.core.v1.container,
  local containerPort = kausal.core.v1.containerPort,
  local deployment = k.apps.v1.deployment,
  local statefulset = k.apps.v1.statefulSet,
  local volume = k.core.v1.volume,
  local pvc = k.core.v1.persistentVolumeClaim,
  local volumeMount = k.core.v1.volumeMount,

  local target_name = 'metrics-generator',
  local tracestore_config_volume = 'tracestore-conf',
  local tracestore_data_volume = 'metrics-generator-data',
  local tracestore_overrides_config_volume = 'overrides',

  tracestore_metrics_generator_ports:: [containerPort.new('prom-metrics', $._config.port)],
  tracestore_metrics_generator_args:: {
    target: target_name,
    'config.file': '/conf/tracestore.yaml',
    'mem-ballast-size-mbs': $._config.ballast_size_mbs,
  },

  tracestore_metrics_generator_pvc::
    pvc.new(tracestore_data_volume)
    + pvc.mixin.spec.resources.withRequests({ storage: $._config.metrics_generator.pvc_size })
    + pvc.mixin.spec.withAccessModes(['ReadWriteOnce'])
    + pvc.mixin.spec.withStorageClassName($._config.metrics_generator.pvc_storage_class)
    + pvc.mixin.metadata.withLabels({ app: target_name })
    + pvc.mixin.metadata.withNamespace($._config.namespace),

  tracestore_metrics_generator_container::
    container.new(target_name, $._images.tracestore) +
    container.withPorts($.tracestore_metrics_generator_ports) +
    container.withArgs($.util.mapToFlags($.tracestore_metrics_generator_args)) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_config_volume, '/conf'),
      volumeMount.new(tracestore_data_volume, '/var/tracestore'),
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
      0,
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
      volume.fromEmptyDir(tracestore_data_volume),
    ])
  ,

  newGeneratorStatefulSet(name, container, with_anti_affinity=true)::
    statefulset.new(
      name,
      $._config.metrics_generator.replicas,
      $.tracestore_metrics_generator_container,
      self.tracestore_metrics_generator_pvc,
      {
        app: target_name,
        [$._config.gossip_member_label]: 'true',
      },
    )
    + kausal.util.antiAffinityStatefulSet
    + statefulset.mixin.spec.withServiceName(target_name)
    + statefulset.mixin.spec.template.metadata.withAnnotations({
      config_hash: std.md5(std.toString($.tracestore_metrics_generator_configmap.data['tracestore.yaml'])),
    })
    + statefulset.mixin.spec.template.spec.withVolumes([
      volume.fromConfigMap(tracestore_config_volume, $.tracestore_metrics_generator_configmap.metadata.name),
      volume.fromConfigMap(tracestore_overrides_config_volume, $._config.overrides_configmap_name),
    ]) +
    statefulset.mixin.spec.withPodManagementPolicy('Parallel') +
    $.util.podPriority('high') +
    (if with_anti_affinity then $.util.antiAffinity else {}),

  tracestore_metrics_generator_statefulset:
    $.newGeneratorStatefulSet(target_name, self.tracestore_metrics_generator_container)
    + statefulset.spec.template.spec.securityContext.withFsGroup(10001)  // 10001 is the UID of the tracestore user
    + statefulset.mixin.spec.withReplicas($._config.metrics_generator.replicas),

  tracestore_metrics_generator_service:
    kausal.util.serviceFor($.tracestore_metrics_generator_deployment),
}
