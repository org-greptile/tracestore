local clusters = import 'clusters.libsonnet';
local tanka = import 'example.com/acme/jsonnet-libs/tanka-util/main.libsonnet';

{
  environment(cluster)::
    tanka.environment.new(
      name='acme/' + cluster.name,
      namespace=cluster.namespace,
      apiserver=cluster.apiServer,
    )
    + tanka.environment.withLabels({ cluster: cluster.name })
    + tanka.environment.withData( cluster.data {

      _config+:: {
        namespace: cluster.namespace,
      },

    } + cluster.dataOverride)
    + {
      spec+: {
        injectLabels: true,
      },
    },

  envs: {
    [cluster.name]: $.environment(cluster)
    for cluster in clusters
  },
}
