(import 'configmap.libsonnet') +
(import 'config.libsonnet') +
{
  local k = import 'ksonnet-util/kausal.libsonnet',
  local container = k.core.v1.container,
  local containerPort = k.core.v1.containerPort,
  local volumeMount = k.core.v1.volumeMount,
  local pvc = k.core.v1.persistentVolumeClaim,
  local statefulset = k.apps.v1.statefulSet,
  local volume = k.core.v1.volume,
  local service = k.core.v1.service,
  local servicePort = service.mixin.spec.portsType,

  local tracestore_config_volume = 'tracestore-conf',
  local tracestore_query_config_volume = 'tracestore-query-conf',
  local tracestore_data_volume = 'tracestore-data',

  namespace:
    k.core.v1.namespace.new($._config.namespace),

  tracestore_pvc::
    pvc.new() +
    pvc.mixin.spec.resources
    .withRequests({ storage: $._config.pvc_size }) +
    pvc.mixin.spec
    .withAccessModes(['ReadWriteOnce']) +
    pvc.mixin.spec
    .withStorageClassName($._config.pvc_storage_class) +
    pvc.mixin.metadata
    .withLabels({ app: 'tracestore' }) +
    pvc.mixin.metadata
    .withNamespace($._config.namespace) +
    pvc.mixin.metadata
    .withName(tracestore_data_volume) +
    { kind: 'PersistentVolumeClaim', apiVersion: 'v1' },

  tracestore_container::
    container.new('tracestore', $._images.tracestore) +
    container.withPorts([
      containerPort.new('prom-metrics', $._config.tracestore.port),
      containerPort.new('memberlist', 9095),
      containerPort.new('otlp', 4317),
    ]) +
    container.withArgs([
      '-target=scalable-single-binary',
      '-config.file=/conf/tracestore.yaml',
      '-mem-ballast-size-mbs=' + $._config.ballast_size_mbs,
    ]) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_config_volume, '/conf'),
      volumeMount.new(tracestore_data_volume, '/var/tracestore'),
    ]) +
    k.util.resourcesRequests('3', '3Gi') +
    k.util.resourcesLimits('5', '5Gi'),

  tracestore_query_container::
    container.new('tracestore-query', $._images.tracestore_query) +
    container.withPorts([
      containerPort.new('jaeger-ui', 16686),
      containerPort.new('jaeger-metrics', 16687),
    ]) +
    container.withArgs([
      '--query.base-path=' + $._config.jaeger_ui.base_path,
      '--grpc-storage-plugin.configuration-file=/conf/tracestore-query.yaml',
    ]) +
    container.withVolumeMounts([
      volumeMount.new(tracestore_query_config_volume, '/conf'),
    ]),

  tracestore_statefulset:
    statefulset.new('tracestore',
                    $._config.tracestore.replicas,
                    [
                      $.tracestore_container,
                      $.tracestore_query_container,
                    ],
                    self.tracestore_pvc,
                    { app: 'tracestore' }) +
    statefulset.mixin.spec.withServiceName('tracestore') +
    statefulset.mixin.spec.template.metadata.withAnnotations({
      config_hash: std.md5(std.toString($.tracestore_configmap.data['tracestore.yaml'])),
    }) +
    statefulset.mixin.metadata.withLabels({ app: $._config.tracestore.headless_service_name, name: 'tracestore' }) +
    statefulset.mixin.spec.selector.withMatchLabels({ name: 'tracestore' }) +
    statefulset.mixin.spec.template.metadata.withLabels({ name: 'tracestore', app: $._config.tracestore.headless_service_name }) +
    statefulset.mixin.spec.template.spec.withVolumes([
      volume.fromConfigMap(tracestore_query_config_volume, $.tracestore_query_configmap.metadata.name),
      volume.fromConfigMap(tracestore_config_volume, $.tracestore_configmap.metadata.name),
    ]),

  tracestore_service:
    k.util.serviceFor($.tracestore_statefulset),

  tracestore_headless_service:
    service.new(
      $._config.tracestore.headless_service_name,
      { app: $._config.tracestore.headless_service_name },
      []
    ) +
    service.mixin.spec.withClusterIP('None') +
    service.mixin.spec.withPublishNotReadyAddresses(true),
}
