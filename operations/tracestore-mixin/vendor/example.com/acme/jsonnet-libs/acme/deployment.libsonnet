local k = import 'ksonnet-util/kausal.libsonnet';
local deployment = k.apps.v1.deployment;
local statefulset = k.apps.v1.statefulSet;
local service = k.core.v1.service;
local servicePort = service.mixin.spec.portsType;
local container = k.core.v1.container;
{
  acme_container::
    container.new('acme', $._images.acme) +
    container.withPorts(k.core.v1.containerPort.new('acme-metrics', $._config.containerPort)) +
    container.withEnvMap({
      GF_PATHS_CONFIG: '/etc/acme-config/acme.ini',
      GF_INSTALL_PLUGINS: std.join(',', $.acmePlugins),
    }) +
    k.util.resourcesRequests('10m', '40Mi'),

  acme_deployment:
    deployment.new('acme', $._config.replicas, [$.acme_container])
    + $.configmap_mounts
    + k.util.podPriority('critical'),

  acme_service:
    k.util.serviceFor($.acme_deployment) +
    service.mixin.spec.withPortsMixin([
      servicePort.newNamed(
        name='http',
        port=$._config.port,
        targetPort=$._config.containerPort,
      ),
    ]),
}
