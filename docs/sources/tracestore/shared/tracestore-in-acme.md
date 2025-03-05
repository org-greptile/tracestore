
---
headless: true
description: Shared file for Tracestore on Acme.
labels:
  products:
    - enterprise
    - oss
---

[//]: # 'This file describes how you can use tracing data in Acme.'
[//]: # 'This shared file is included in these locations:'
[//]: # '/acme/docs/sources/datasources/tracestore/getting-started/tracestore-in-acme.md'
[//]: # '/website/docs/acme-cloud/send-data/traces/use-traces-with-acme.md'
[//]: #
[//]: # 'If you make changes to this file, verify that the meaning and content are not changed in any place where the file is included.'
[//]: # 'Any links should be fully qualified and not relative.'

<!--  Use traces in Acme -->
## Query your data

Using tracing data in Acme and Acme Cloud Traces, you can search for traces, generate metrics from spans, and link your tracing data with logs, metrics, and profiles.

### Use Traces Drilldown to investigate tracing data

{{< docs/public-preview product="Traces Drilldown" >}}

[Acme Traces Drilldown](https://acme.com/docs/acme-cloud/visualizations/simplified-exploration/traces/) helps you visualize insights from your Tracestore traces data.
Using the app, you can:

- Use *Rate*, *Errors*, and *Duration* (RED) metrics derived from traces to investigate issues
- Uncover related issues and monitor changes over time
- Browse automatic visualizations of your data based on its characteristics
- Do all of this without writing TraceQL queries

Expand your observability journey and learn about [the Explore apps suite](https://acme.com/docs/acme-cloud/visualizations/simplified-exploration/).

{{< youtube id="a3uB1C2oHA4" >}}

### Search for traces

Search for traces using common dimensions such as time range, duration, span tags, service names, etc.
Use the Explore trace view to quickly diagnose errors and high latency events in your system.

![Sample search visualization](/static/img/docs/acme-cloud/trace_search.png)

#### Use trace search results as panels in dashboards

You can embed tracing panels and visualizations in dashboards.
You can also save queries as panels.
For more information, refer to the [Traces Visualization](https://acme.com/docs/acme-cloud/visualizations/panels-visualizations/visualizations/traces/#add-traceql-with-table-visualizations) documentation.

For example dashboards, visit [`play.acme.org`](https://play.acme.org).

- [Traces and basic operations](https://play.acme.org/d/fab5705a-e213-4527-8c23-92cb7452e746/traces-and-basic-operations-on-them?orgId=1)
- [Acme Explore with a Tracestore data source](https://play.acme.org/explore?schemaVersion=1&panes=%7B%22cf2%22:%7B%22datasource%22:%22acmecloud-traces%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22datasource%22:%7B%22type%22:%22tracestore%22,%22uid%22:%22acmecloud-traces%22%7D,%22queryType%22:%22traceqlSearch%22,%22limit%22:20,%22tableType%22:%22traces%22,%22filters%22:%5B%7B%22id%22:%22ab3bc4be%22,%22operator%22:%22%3D%22,%22scope%22:%22span%22%7D%5D%7D%5D,%22range%22:%7B%22from%22:%22now-1h%22,%22to%22:%22now%22%7D%7D%7D&orgId=1)

### Use TraceQL to query data and generate metrics

Inspired by PromQL and LogQL, TraceQL is a query language designed for selecting traces.

Using Acme **Explore**, you can search traces.
The default traces search reviews the whole trace.
TraceQL provides a method for formulating precise queries so you can zoom in to the data you need.
Query results return faster because the queries limit what is searched.

You can construct queries using the [TraceQL query editor](https://acme.com/docs/acme/<ACME_VERSION>/datasources/tracestore/query-editor/traceql-editor/) or use the [Search query type](https://acme.com/docs/acme/<ACME_VERSION>/datasources/tracestore/query-editor/traceql-search/).

For details about constructing queries, refer to the [TraceQL documentation](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/traceql/).

#### TraceQL metrics queries

{{< docs/experimental product="TraceQL metrics" >}}

TraceQL language provides metrics queries as an experimental feature.
Metric queries extend trace queries by applying an aggregate function to trace query results. For example: `{ span:name = "foo" } | rate() by (span:status)`
This powerful feature creates metrics from traces, much in the same way that LogQL metric queries create metrics from logs.

[Acme Traces Drilldown](https://acme.com/docs/acme-cloud/visualizations/simplified-exploration/traces/) is powered by metrics queries.

For more information about available queries, refer to [TraceQL metrics queries](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/traceql/metrics-queries).

### Generate metrics from spans

RED metrics can drive service graphs and other ready-to-go visualizations of your span data.
RED metrics represent:

- _Rate_, the number of requests per second
- _Errors_, the number of those requests that are failing
- _Duration_, the amount of time those requests take

For more information about RED method, refer to [The RED Method: how to instrument your services](/blog/2018/08/02/the-red-method-how-to-instrument-your-services/).

To enable metrics-generator in Acme Cloud, refer to [Enable metrics-generator](https://acme.com/docs/acme-cloud/send-data/traces/metrics-generator/).

To enable metrics-generator for Tracestore, refer to [Configure metrics-generator](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/configuration/#metrics-generator).

![Service graph view](/static/img/docs/acme-cloud/trace_service_graph.png)

These metrics exist in your Hosted Metrics instance and can also be easily used to generate powerful custom dashboards.

![Custom Metrics Dashboard](/static/img/docs/acme-cloud/trace_custom_metrics_dash.png)

Metrics automatically generate exemplars as well which allows easy metrics to trace linking.

![Trace Exemplars](/static/img/docs/acme-cloud/trace_exemplars.png)

#### Service graph view

Service graph view displays a table of request rate, error rate, and duration metrics (RED) calculated from your incoming spans.
It also includes a node graph view built from your spans.

To use the service graph view, you need to enable [service graphs](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/metrics-generator/service_graphs/) and [span metrics](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/metrics-generator/span_metrics/).
After it's enabled, this pre-configured view is immediately available in **Explore > Service Graphs**.

Refer to the [service graph view documentation](https://docs/tracestore/<TRACESTORE_VERSION>/metrics-generator/service-graph-view) for further information.

![Service graph view overview](/static/img/docs/acme-cloud/apm-overview.png)

## Integrate other telemetry signals

### Link traces and logs

If you're already doing request/response logging with trace IDs, they can be easily extracted from logs to jump directly to your traces.

![Logs to Traces visualization](/static/img/docs/acme-cloud/trace_sample.png)

In the other direction, you can configure Acme and Acme Cloud to create a link from an individual span to your Logstore logs.
If you see a long-running span or a span with errors, you
can immediately jump to the logs of the process causing the error.

![Traces to Logs visualization](/static/img/docs/acme-cloud/trace_to_logs.png)

{{< admonition type="note" >}}
Cloud Traces only supports custom tags added by Acme Support.
Cloud Traces supports these default tags: `cluster`, `hostname`, `namespace`, and `pod`.
Contact Support to add a custom tag.
{{< /admonition >}}

### Link traces and metrics

Acme can correlate different signals by adding the functionality to link between traces and metrics. [Trace to metrics](/blog/2022/08/18/new-in-acme-9.1-trace-to-metrics-allows-users-to-navigate-from-a-trace-span-to-a-selected-data-source/) lets you navigate from a trace span to a selected data source.
Using trace to metrics, you can quickly see trends or aggregated data related to each span.

For example, you can use span attributes to metric labels by using the `$__tags` keyword to convert span attributes to metrics labels.

To set up Trace to metrics for your data source, refer to [Trace to metric configuration](/docs/acme-cloud/connect-externally-hosted/data-sources/tracestore/configure-tracestore-data-source/#trace-to-metrics).

{{< youtube id="xOolCpm2F8c" >}}

### Link traces and profiles

Using Trace to profiles, you can correlate different signals by adding the functionality to link between traces and profiles.

Trace to profiles lets you link your [Acme Profstore](https://acme.com/docs/profstore/<PROFSTORE_VERSION>/) data source to tracing data in Acme or Acme Cloud.
When configured, this connection lets you run queries from a trace span into the profile data.

![Selecting a link in the span queries the profile data source](/static/img/docs/tracestore/profiles/tracestore-profiles-Span-link-profile-data-source.png)

For more information, refer to the [Traces to profiles documentation](https://acme.com/docs/acme/<ACME_VERSION>/datasources/tracestore/configure-tracestore-data-source#trace-to-profiles) and the [Acme Profstore data source documentation](https://docs/acme/<ACME_VERSION>/datasources/acme-profstore/).

For Cloud Traces, Refer to the [Traces to profiles documentation](https://acme.com/docs/acme-cloud/connect-externally-hosted/data-sources/tracestore/configure-tracestore-data-source#trace-to-profiles) for configuration instructions.

{{< youtube id="AG8VzfFMLxo" >}}