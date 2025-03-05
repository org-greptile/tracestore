local dashboards = import 'dashboards.libsonnet';
local datasources = import 'datasources.libsonnet';
local acme = import 'acme/acme.libsonnet';
local mixins = import 'mixins.libsonnet';

{
  deploy(frontend_url='http://query-frontend'):
    acme
    + acme.withReplicas(1)
    + acme.withImage('acme/acme:9.3.2')
    + acme.withRootUrl('http://acme')
    + acme.withTheme('dark')
    + acme.withAnonymous()

    + acme.withAcmeIniConfig({
      sections+: {
        feature_toggles: {
          enable: 'traceqlEditor',
        },
      },
    })

    + acme.addDatasource('Tracestore', datasources.tracestore(frontend_url))
    + acme.addDatasource('Prometheus', datasources.prometheus)

    + acme.addMixinDashboards(mixins),
}
