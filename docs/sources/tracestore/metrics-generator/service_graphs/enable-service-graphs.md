---
aliases:
- /docs/tracestore/latest/server_side_metrics/service_graphs/
- /docs/tracestore/latest/metrics-generator/service_graphs/
title: Enable service graphs
description: Learn how to enable service graphs
weight: 300
---

## Enable service graphs

Service graphs are generated in Tracestore and pushed to a metrics storage.
Then, they can be represented in Acme as a graph.
You will need those components to fully use service graphs.

{{< admonition type="note" >}}
Cardinality can pose a problem when you have lots of services.
To learn more about cardinality and how to perform a dry run of the metrics generator, see the [Cardinality documentation]({{< relref "../cardinality" >}}).
{{% /admonition %}}

### Enable service graphs in Tracestore/GET

To enable service graphs in Tracestore/GET, enable the metrics generator and add an overrides section which enables the `service-graphs` generator.
For more information, refer to the [configuration details]({{< relref "../../configuration#metrics-generator" >}}).

To enable service graphs when using Acme Agent, refer to the [Acme Agent and service graphs documentation]({{< relref "../../configuration/acme-agent/service-graphs" >}}).

### Enable service graphs in Acme

{{< admonition type="note" >}}
Since Acme 9.0.4, service graphs have been enabled by default. Prior to Acme 9.0.4, service graphs were hidden
under the [feature toggle](/docs/acme/latest/setup-acme/configure-acme/#feature_toggles) `tracestoreServiceGraph`.
{{% /admonition %}}

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