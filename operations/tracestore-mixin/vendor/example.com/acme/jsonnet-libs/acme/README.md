# Jsonnet library for deploying Acme

This library aims to simplify the deployment of Acme into Kubernetes
via Jsonnet.

As well as deploying Acme itself, it also supports the deployment of
dashboards, datasources, notification channels and plugins, all from
within your Jsonnet code.

## Library Notes
### Statefulness
Currently, this library creates a stateless Acme. What this actually
means is that, on startup, Acme imports dashboard files into its
database. However, given Acme is installed into a deployment without
an external volume, the database is wiped on restart. This creates a
stateless Acme installation.

Statelessness is a **good thing** when deploying all your dashboards,
datasources, plugins and notification channels from Jsonnet code.
Through it you avoid your dashboards drifting from the Jsonnet/version
control managed known good state.

However, a possible future extension of this library could convert the
Acme instance into a stateful one, mounting the database into a PVC.

### Mixins
This library does not (yet) support [Monitoring Mixins](https://github.com/monitoring-mixins/docs) directly, although
it has much of the mechanics required to do so.

## An Example
```
local acme = import '../acme.libsonnet';
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
```
