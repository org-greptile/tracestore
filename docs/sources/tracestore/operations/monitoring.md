---
title: Monitoring Tracestore
weight: 40
---

# Monitoring Tracestore

Tracestore is instrumented to expose metrics, logs and traces.
Additionally, the Tracestore repository has a [mixin](https://example.com/acme/tracestore/tree/main/operations/tracestore-mixin) that includes a
set of dashboards, rules and alerts.
Together, these can be used to monitor Tracestore in production.

## Instrumentation

Tracestore is already instrumented with metrics, logs and traces.
These can be collected to observe Tracestore.

### Metrics

Tracestore is instrumented with [Prometheus metrics](https://prometheus.io/).
It emits RED metrics for most services and backends.
The [Tracestore mixin](#dashboards) provides several dashboards using these metrics.

### Logs

Tracestore emits logs in the `key=value` ([logfmt](https://brandur.org/logfmt)) format.

### Traces

Tracestore uses the [Jaeger Golang SDK](https://github.com/jaegertracing/jaeger-client-go) for tracing instrumentation.
As of this writing, the complete read path and some parts of the write of Tracestore are instrumented for tracing.

The tracer can be configured [using environment variables](https://github.com/jaegertracing/jaeger-client-go#environment-variables).
To enable tracing, set one of the following: `JAEGER_AGENT_HOST` and `JAEGER_AGENT_PORT`, or `JAEGER_ENDPOINT`.

The Jaeger client uses remote sampling by default, if the management server is not available no traces will be sent.
To always send traces (no sampling), set the following environment variables:

```
JAEGER_SAMPLER_TYPE=const
JAEGER_SAMPLER_PARAM=1
```

## Dashboards

The [Tracestore mixin](https://example.com/acme/tracestore/tree/main/operations/tracestore-mixin) has four Acme dashboards in the `yamls` folder that you can download and import into your Acme UI.
At the moment, these work well when Tracestore is run in a Kubernetes (k8s) environment and metrics scraped have the
`cluster` and `namespace` labels.

### Tracestore Reads dashboard

> This is available as `tracestore-reads.json`.

The Reads dashboard gives information information on Requests, Errors and Duration (R.E.D) on the Query Path of Tracestore.
Each query touches the Gateway, Tracestore-Query, Query-Frontend, Queriers, Ingesters, Cache (if present) and the backend.

Use this dashboard to monitor the performance of each of the above mentioned components and to decide the number of
replicas in each deployment.

### Tracestore Writes dashboard

> This is available as `tracestore-writes.json`.

The Writes dashboard gives information information on Requests, Errors and Duration (R.E.D) on the write/ingest Path of Tracestore.
A write query touches the Gateway, Distributors, Ingesters and eventually the backend. This dashboard also gives information
on the number of operations performed by the Compactor to the backend.

Use this dashboard to monitor the performance of each of the above mentioned components and to decide the number of
replicas in each deployment.

### Tracestore Resources dashboard

> This is available as `tracestore-resources.json`.

The Resources dashboard provides information on `CPU`, `Container Memory` and `Go Heap Inuse`, and is useful for resource
provisioning for the different Tracestore components.

Use this dashboard to see if any components are running close to their assigned limits!

### Tracestore Operational dashboard

> This is available as `tracestore-operational.json`.

The Tracestore Operational dashboard deserves special mention b/c it probably a stack of dashboard anti-patterns.
It's big and complex, doesn't use jsonnet and displays far too many metrics in one place.  And I love it.
For just getting started the Reads, Write and Resources dashboards are great places to learn how to monitor Tracestore in an opaque way.

This dashboard is included in this repo for two reasons:

- It provides a stack of metrics for other operators to consider monitoring while running Tracestore.
- We want it in our internal infrastructure and we vendor the tracestore-mixin to do this.


## Rules and alerts

The Rules and Alerts are available as [yaml files in the compiled mixin](https://example.com/acme/tracestore/tree/main/operations/tracestore-mixin-compiled) on the repository.

To set up alerting, download the provided json files and configure them for use on your Prometheus monitoring server.

Check the [runbook](https://example.com/acme/tracestore/blob/main/operations/tracestore-mixin/runbook.md) to understand the
various steps that can be taken to fix firing alerts!
