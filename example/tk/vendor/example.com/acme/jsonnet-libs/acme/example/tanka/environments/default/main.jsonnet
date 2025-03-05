local acme = import '../../../acme.libsonnet';
local dashboards = import 'dashboards.libsonnet';
local datasources = import 'datasources.libsonnet';
local mixins = import 'mixins.libsonnet';

acme
+ acme.withReplicas(3)
+ acme.withImage('acme/acme:8.0.0')
+ acme.withRootUrl('http://acme.example.com')
+ acme.withTheme('dark')
+ acme.withAnonymous()

// Plugins
+ acme.addPlugin('fetzerch-sunandmoon-datasource')

// Datasources
+ acme.addDatasource('Prometheus', datasources.prometheus)
+ acme.addDatasource('NYC', datasources.sun_and_moon)

// Dashboards
+ acme.addDashboard('node-exporter-full', dashboards.node_exporter, 'Node Exporter')
+ acme.addDashboard('nyc', dashboards.nyc, 'Sun and Moon')

// Mixins
+ acme.addMixinDashboards(mixins)
