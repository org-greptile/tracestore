local tanka = import 'example.com/acme/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);
local kustomize = tanka.kustomize.new(std.thisFile);

{
  // render the Acme Chart, set namespace to "test"
  acme: helm.template('acme', './charts/acme', {
    values: {
      persistence: { enabled: true },
      plugins: ['acme-clock-panel'],
    },
    namespace: 'test',
  }),

  // render the Prometheus Kustomize
  // then entrypoint for `kustomize build` will be ./base/prometheus/kustomization.yaml
  prometheus: kustomize.build('./base/prometheus'),
}
