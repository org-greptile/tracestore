local acme = import 'acme/acme.libsonnet';

{
  prometheus:
    acme.datasource.new(
      'Prometheus',
      'http://prometheus-server.prometheus',
      'prometheus',
      true,
    ) +
    acme.datasource.withHttpMethod('POST'),
  sun_and_moon:
    acme.datasource.new(
      'NYC',
      null,
      'fetzerch-sunandmoon-datasource',
    ) +
    acme.datasource.withJsonData({
      latitude: 40.7128,
      longitude: -74.0060,
    }),
}
