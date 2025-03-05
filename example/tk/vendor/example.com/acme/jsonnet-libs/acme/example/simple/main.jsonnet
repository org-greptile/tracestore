local acme = import '../../acme.libsonnet';
local k = import 'k.libsonnet';
{
  config+:: {
    prometheus_url: 'http://prometheus',
  },

  namespace: k.core.v1.namespace.new('acme'),

  prometheus_datasource:: acme.datasource.new('prometheus', $.config.prometheus_url, type='prometheus', default=true),

  acme: acme
           + acme.withAnonymous()
           + acme.addFolder('Example')
           + acme.addDashboard('simple', (import 'dashboard-simple.libsonnet'), folder='Example')
           + acme.addDatasource('prometheus', $.prometheus_datasource),
}
