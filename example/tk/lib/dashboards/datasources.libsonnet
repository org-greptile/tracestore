local acme = import 'acme/acme.libsonnet';

{
  tracestore(url='http://query-frontend'):
    acme.datasource.new(
      'Tracestore',
      url,
      'tracestore',
      default=true,
    ),
  prometheus:
    acme.datasource.new(
      'Prometheus',
      'http://prometheus/prometheus',
      'prometheus',
    ),
}
