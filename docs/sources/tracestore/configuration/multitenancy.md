---
title: Multi-tenancy
weight: 60
---
# Multi-tenancy

Tracestore is a multi-tenant distributed tracing backend. It supports multi-tenancy through the use
of a header: `X-Scope-OrgID`. 
This guide details how to setup multi-tenancy.

If you're interested in setting up multi-tenancy, consult the [multi-tenant example](https://example.com/acme/tracestore/tree/main/example/docker-compose/otel-collector-multitenant)
in the repo. This example uses the following settings to achieve multi-tenancy in Tracestore.

Configure the OTEL Collector to attach the X-Scope-OrgID header on push:

```
exporters:
  otlp:
    headers:
      x-scope-orgid: foo-bar-baz
```

### Acme 7.5.x and higher

Configure the Tracestore data source in Acme to pass the tenant with the same header:

```
- name: Tracestore-Multitenant
  jsonData:
    httpHeaderName1: 'X-Scope-OrgID'
  secureJsonData:
    httpHeaderValue1: 'foo-bar-baz'
```

### Acme 7.4.x

Acme 7.4.x has the following configuration requirements:

- Configure the Tracestore data source in Acme to pass the tenant as a bearer token. This is necessary because it is the only header that Jaeger can be configured to pass to its GRPC plugin.

```
- name: Tracestore-Multitenant
  jsonData:
    httpHeaderName1: 'Authorization'
  secureJsonData:
    httpHeaderValue1: 'Bearer foo-bar-baz'
```

- Configure Jaeger Query to pass the bearer token to its backend.

```
--query.bearer-token-propagation=true
```

## Important notes

Multi-tenancy on ingestion is currently [only working](https://example.com/acme/tracestore/issues/495) with GPRC and this may never change. 
It is strongly recommended to use the OpenTelemetry Collector to support multi-tenancy as described above.

## Enabling multi-tenancy

To enable multi-tenancy on Tracestore backend, set the following configuration value on all Tracestore components:

```
multitenancy_enabled: true
```

or from the command line:

```
--multitenancy.enabled=true
```

This option will force all Tracestore components to require the `X-Scope-OrgID` header.
