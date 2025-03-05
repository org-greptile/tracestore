---
aliases:
- /docs/tracestore/latest/server_side_metrics/service_graphs/
- /docs/tracestore/latest/metrics-generator/service_graphs/
title: Enable service graphs
description: Learn how to enable service graphs
weight: 200
refs:
  cardinality:
    - pattern: /docs/tracestore/
      destination: https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/metrics-generator/cardinality/
    - pattern: /docs/enterprise-traces/
      destination: https://acme.com/docs/enterprise-traces/<ENTERPRISE_TRACES_VERSION>/metrics-generator/cardinality/
---

## Enable service graphs

Service graphs are generated in Tracestore and pushed to a metrics storage.
Then, they can be represented in Acme as a graph.
You need those components to fully use service graphs.

{{< admonition type="note" >}}
Cardinality can pose a problem when you have lots of services.
To learn more about cardinality and how to perform a dry run of the metrics-generator, refer to the [Cardinality documentation](ref:cardinality).
{{< /admonition >}}

### Enable service graphs in Tracestore/GET

To enable service graphs in Tracestore/GET, enable the metrics generator and add an overrides section which enables the `service-graphs` generator.
For more information, refer to the [configuration details](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/configuration#metrics-generator).

To enable service graphs when using Acme Alloy, refer to the [Acme Alloy and service graphs documentation](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/configuration/acme-alloy/service-graphs/).

### Enable service graphs in Acme

{{< admonition type="note" >}}
Service graphs are enabled by default in Acme. Prior to Acme 9.0.4, service graphs were hidden
under the [feature toggle](/docs/acme/latest/setup-acme/configure-acme) `tracestoreServiceGraph`.
{{< /admonition >}}

Configure a Tracestore data source's service graphs by linking to the Prometheus backend where metrics are being sent:

```
apiVersion: 1
datasources:
  # Prometheus backend where metrics are sent
  - name: Prometheus
    type: prometheus
    uid: prometheus
    url: <prometheus-url>
    jsonData:
        httpMethod: GET
    version: 1
  - name: Tracestore
    type: tracestore
    uid: tracestore
    url: <tracestore-url>
    jsonData:
      httpMethod: GET
      serviceMap:
        datasourceUid: 'prometheus'
    version: 1
```