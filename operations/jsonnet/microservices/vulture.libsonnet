{
  local k = import 'ksonnet-util/kausal.libsonnet',
  local container = k.core.v1.container,
  local containerPort = k.core.v1.containerPort,
  local deployment = k.apps.v1.deployment,

  local target_name = 'vulture',
  local port = 8080,

  tracestore_vulture_container::
    container.new(target_name, $._images.tracestore_vulture) +
    container.withPorts([
      containerPort.new('prom-metrics', port),
    ]) +
    container.withArgs([
      '-prometheus-listen-address=:' + port,
      '-tracestore-push-url=' + $._config.vulture.tracestorePushUrl,
      '-tracestore-query-url=' + $._config.vulture.tracestoreQueryUrl,
      '-tracestore-org-id=' + $._config.vulture.tracestoreOrgId,
      '-tracestore-retention-duration=' + $._config.vulture.tracestoreRetentionDuration,
      '-tracestore-search-backoff-duration=' + $._config.vulture.tracestoreSearchBackoffDuration,
      '-tracestore-read-backoff-duration=' + $._config.vulture.tracestoreReadBackoffDuration,
      '-tracestore-write-backoff-duration=' + $._config.vulture.tracestoreWriteBackoffDuration,
    ]) +
    k.util.resourcesRequests('50m', '100Mi') +
    k.util.resourcesLimits('100m', '500Mi'),

  tracestore_vulture_deployment:
    deployment.new(target_name,
                   $._config.vulture.replicas,
                   [
                     $.tracestore_vulture_container,
                   ],
                   {
                     app: target_name,
                   }),
}
