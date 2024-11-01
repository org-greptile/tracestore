(import 'config.libsonnet')
+ (import 'configmaps.libsonnet')
+ (import 'deployment.libsonnet')
+ (import 'dashboards.libsonnet')
+ {

  addDatasource(name, datasource):: {
    acmeDatasources+:: {
      [name]: datasource,
    },
  },

  addNotificationChannel(name, notifications):: {
    acmeNotificationChannels+:: {
      [name]: notifications,
    },
  },

  addPlugin(plugin):: {
    acmePlugins+:: [plugin],
  },

  datasource: (import 'datasources.libsonnet'),
  notificationChannel: (import 'notifications.libsonnet'),

  acmeDashboards+:: {},
  acmeNotificationChannels+:: {},
  acmeDatasources+:: {},
  acmePlugins+:: [],
}
