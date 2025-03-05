{
  local k = import 'ksonnet-util/kausal.libsonnet',

  local container = k.core.v1.container,
  local containerPort = k.core.v1.containerPort,
  local volumeMount = k.core.v1.volumeMount,
  local statefulset = k.apps.v1.statefulSet,
  local volume = k.core.v1.volume,
  local envVar = k.core.v1.envVar,
  local configMap = k.core.v1.configMap,

  local target_name = 'block-builder',
  local tracestore_config_volume = 'tracestore-conf',
  local tracestore_overrides_config_volume = 'overrides',

  // Statefulset

  tracestore_block_builder_ports:: [containerPort.new('prom-metrics', $._config.port)],

  tracestore_block_builder_args:: {
    target: target_name,
    'config.file': '/conf/tracestore.yaml',
    'mem-ballast-size-mbs': $._config.ballast_size_mbs,
  },

  tracestore_block_builder_container::
    container.new(target_name, $._images.tracestore) +
    container.withPorts($.tracestore_block_builder_ports) +
    container.withArgs($.util.mapToFlags($.tracestore_block_builder_args)) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_config_volume, '/conf'),
      volumeMount.new(tracestore_overrides_config_volume, '/overrides'),
    ]) +
    $.util.withResources($._config.block_builder.resources) +
    (if $._config.variables_expansion then container.withEnvMixin($._config.variables_expansion_env_mixin) else {}) +
    $.util.readinessProbe +
    (if $._config.variables_expansion then container.withArgsMixin(['-config.expand-env=true']) else {}),

  newBlockBuilderStatefulSet(concurrent_rollout_enabled=false, max_unavailable=1)::
    statefulset.new(target_name, $._config.block_builder.replicas, $.tracestore_block_builder_container, [], { app: target_name }) +
    statefulset.mixin.spec.withServiceName(target_name) +
    statefulset.spec.template.spec.securityContext.withFsGroup(10001) +  // 10001 is the UID of the tracestore user
    statefulset.mixin.spec.template.metadata.withAnnotations({
      config_hash: std.md5(std.toString($.tracestore_block_builder_configmap.data['tracestore.yaml'])),
    }) +
    statefulset.mixin.spec.template.spec.withVolumes([
      volume.fromConfigMap(tracestore_config_volume, $.tracestore_block_builder_configmap.metadata.name),
      volume.fromConfigMap(tracestore_overrides_config_volume, $._config.overrides_configmap_name),
    ]) +
    statefulset.mixin.spec.withPodManagementPolicy('Parallel') +
    (
      if !concurrent_rollout_enabled then {} else
        statefulset.mixin.spec.selector.withMatchLabels({ name: 'block-builder', 'rollout-group': 'block-builder' }) +
        statefulset.mixin.spec.updateStrategy.withType('OnDelete') +
        statefulset.mixin.metadata.withLabelsMixin({ 'rollout-group': 'block-builder' }) +
        statefulset.mixin.metadata.withAnnotationsMixin({ 'rollout-max-unavailable': std.toString(max_unavailable) }) +
        statefulset.mixin.spec.template.metadata.withLabelsMixin({ 'rollout-group': 'block-builder' })
    ),

  tracestore_block_builder_statefulset:
    $.newBlockBuilderStatefulSet($._config.block_builder_concurrent_rollout_enabled, $._config.block_builder_max_unavailable),

  // Configmap

  tracestore_block_builder_configmap:
    configMap.new('tracestore-block-builder') +
    configMap.withData({
      'tracestore.yaml': k.util.manifestYaml($.tracestore_block_builder_config),
    }),

  // Service

  tracestore_block_builder_service:
    k.util.serviceFor($.tracestore_block_builder_statefulset),

}
