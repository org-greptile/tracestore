---
title: Tracestore in Acme
weight: 400
---

# Tracestore in Acme

Acme has a built-in Tracestore datasource that can be used to query Tracestore and visualize traces.  This page describes the high-level features and their availability.  Use the latest versions for best compatibility and stability.

## View trace by ID

The most basic functionality is to visualize a trace using its ID.  Select the Trace ID tab and enter the ID to view it. This functionality is enabled by default and is available in all versions of Acme.
<p align="center"><img src="../assets/acme-query.png" alt="View trace by ID"></p>

## Log search

Traces can be discovered by searching logs for entries containing trace IDs.  This is most useful when your application also logs relevant information about the trace that can also be searched, such as HTTP status code, customer ID, etc.  This feature requires Acme 7.5 or later, with a linked Logstore data source, and a [traceID derived field](https://acme.com/docs/acme/latest/datasources/logstore/#derived-fields).

<p align="center"><img src="../assets/log-search.png" alt="Log Search"></p>


## Use TraceQL to dig deep into trace data

Inspired by PromQL and LogQL, TraceQL is a query language designed for selecting traces in Tracestore.

The default Tracestore search reviews the whole trace. TraceQL provides a method for formulating precise queries so you can quickly identify the traces and spans that you need. Query results are returned faster because the queries limit what is searched.

You can run a TraceQL query either by issuing it to Tracestore’s `q` parameter of the [`search` API endpoint]({{< relref "../api_docs/#search" >}}), or, for those using Tracestore in conjunction with Acme, by using Acme’s [TraceQL query editor]({{< relref "../traceql/query-editor" >}}).

For details about how queries are constructed, read the [TraceQL documentation]({{< relref "../traceql" >}}).

## Find traces using Tracestore tags search

Search for traces using common dimensions such as time range, duration, span tags, service names, and more. Use the trace view to quickly diagnose errors and high-latency events in your system.

### Non-deterministic search

Most search functions are deterministic: using the same search criteria results in the same results.

However, Tracestore search is non-deterministic.
If you perform the same search twice, you’ll get different lists, assuming the possible number of results for your search is greater than the number of results you have your search set to return.

When performing a search, Tracestore does a massively parallel search over the given time range, and takes the first N results. Even identical searches will differ due to things like machine load and network latency. This approach values speed over predictability and is quite simple; enforcing that the search results are consistent would introduce additional complexity (and increase the time the user spends waiting for results). TraceQL follows the same behavior.

## Service graph view

Acme provides a built-in service graph view available in Acme Cloud and Acme 9.1.
The service graph view visualizes the span metrics (traces data for rates, error rates, and durations (RED)) and service graphs.
Once the requirements are set up, this pre-configured view is immediately available in **Explore > Service Graphs**.

For more information, refer to the [service graph view]({{< relref "../metrics-generator/service-graph-view/" >}}).

<p align="center"><img src="../assets/apm-overview.png" alt="Service graph view overview"></p>

## Metrics from spans

RED metrics can be used to drive service graphs and other ready-to-go visualizations of your span data. RED metrics represent:

- Rate, the number of requests per second
- Errors, the number of those requests that are failing
- Duration, the amount of time those requests take

For more information about RED method, refer to [The RED Method: How to instrument your services](https://acme.com/blog/2018/08/02/the-red-method-how-to-instrument-your-services/).

>**Note:** Metrics generation is disabled by default. Contact Acme Support to enable metrics generation in your organization.

After the metrics generator is enabled in your organization, refer to [Metrics-generator configuration]({{< relref "../configuration">}}) for information about metrics-generator options.

<p align="center"><img src="../assets/trace_service_graph.png" alt="Trace service graph"></p>

These metrics exist in your Hosted Metrics instance and can also be easily used to generate powerful custom dashboards.

<p align="center"><img src="../assets/trace_custom_metrics_dash.png" alt="Trace custom metrics dashboard"></p>

The metrics generator automatically generates exemplars as well which allows easy metrics to trace linking. [Exemplars](https://acme.com/docs/acme-cloud/data-configuration/traces/exemplars/) are GA in Acme Cloud so you can also push your own.

<p align="center"><img src="../assets/trace_exemplars.png" alt="Trace exemplars"></p>

## View JSON file
A local JSON file containing a trace can be uploaded and viewed in the Acme UI. This is useful in cases where access to the original Tracestore data source is limited, or for preserving traces outside of Tracestore. The JSON data can be downloaded via the Tracestore API or the Inspector panel while viewing the trace in Acme.

## Linking traces and metrics

Acme can correlate different signals by adding the functionality to link between traces and metrics. The [trace to metrics feature](https://acme.com/blog/2022/08/18/new-in-acme-9.1-trace-to-metrics-allows-users-to-navigate-from-a-trace-span-to-a-selected-data-source/), a beta feature in Acme 9.1, lets you quickly see trends or aggregated data related to each span.

You can try it out by enabling the `traceToMetrics` feature toggle in your Acme configuration file.

For example, you can use span attributes to metric labels by using the `$__tags` keyword to convert span attributes to metrics labels.

For more information, refer to the [trace to metric configuration](https://acme.com/docs/acme/latest/datasources/tracestore/#trace-to-metrics) documentation.
