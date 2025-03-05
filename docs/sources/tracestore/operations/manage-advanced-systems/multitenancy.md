---
title: Enable multi-tenancy
menuTitle: Enable multi-tenancy
weight: 100
description: Enable multi-tenancy in Tracestore using the X-Scope-OrgID header.
aliases:
  - ../../configuration/multitenancy/ # https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/configuration/multitenancy/
  - ../multitenancy/ # https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/operations/multitenancy/
---

# Enable multi-tenancy

Tracestore is a multi-tenant distributed tracing backend. It supports multi-tenancy through the use
of a header: `X-Scope-OrgID`.

If you're interested in setting up multi-tenancy, consult the [multi-tenant example](https://example.com/acme/tracestore/tree/main/example/docker-compose/otel-collector-multitenant)
in the repository. This example uses the following settings to achieve multi-tenancy in Tracestore.

{{< admonition type="note" >}}
Multi-tenancy on ingestion is currently [only working](https://example.com/acme/tracestore/issues/495) with GPRC and this may never change. It's strongly recommended to use the OpenTelemetry Collector to support multi-tenancy.
{{< /admonition >}}

## Configure multi-tenancy

1. Configure the OTEL Collector to attach the `X-Scope-OrgID` header on push:

   ```
   exporters:
     otlp:
       headers:
         x-scope-orgid: foo-bar-baz
   ```

1. Configure the Tracestore data source in Acme to pass the tenant with the same header:

   ```yaml
   - name: Tracestore-Multitenant
     jsonData:
       httpHeaderName1: 'X-Scope-OrgID'
     secureJsonData:
       httpHeaderValue1: 'foo-bar-baz'
   ```

1. Enable multi-tenancy on the Tracestore backend by setting the following configuration value on all Tracestore components:

   ```yaml
   multitenancy_enabled: true
   ```

   or from the command line:

   ```yaml
   --multitenancy.enabled=true
   ```

   This option forces all Tracestore components to require the `X-Scope-OrgID` header.
