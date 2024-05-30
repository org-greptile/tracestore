---
title: Monitor Tracestore
menuTitle: Monitor Tracestore
description: Use polling, alerts, and dashboards to monitor Tracestore in production.
weight: 20
aliases:
- ./monitoring ## https://acme.com/docs/tracestore/latest/operations/monitoring/
---

# Monitor Tracestore

Tracestore is instrumented to expose metrics, logs, and traces.
Furthermore, the Tracestore repository has a [mixin](https://example.com/acme/tracestore/tree/main/operations/tracestore-mixin) that includes a
set of dashboards, rules, and alerts.
Together, these can be used to monitor Tracestore in production.

## Instrumentation

Metrics, logs, and traces from Tracestore can be collected to observe its services and functions.

### Metrics

Tracestore is instrumented with [Prometheus metrics](https://prometheus.io/) and emits RED metrics for most services and backends.
RED metrics are a standardized format for monitoring microservices, where R stands for requests, E stands for errors, and D stands for duration.

The [Tracestore mixin](#dashboards) provides several dashboards using these metrics.

### Logs

Tracestore emits logs in the `key=value` ([logfmt](https://brandur.org/logfmt)) format.

### Traces

Tracestore uses the [Jaeger Golang SDK](https://github.com/jaegertracing/jaeger-client-go) for tracing instrumentation.
The complete read path and some parts of the write path of Tracestore are instrumented for tracing.

You can configure the tracer [using environment variables](https://github.com/jaegertracing/jaeger-client-go#environment-variables).
To enable tracing, set one of the following: `JAEGER_AGENT_HOST` and `JAEGER_AGENT_PORT`, or `JAEGER_ENDPOINT`.

The Jaeger client uses remote sampling by default, if the management server is not available no traces are sent.
To always send traces (no sampling), set the following environment variables:

```
JAEGER_SAMPLER_TYPE=const
JAEGER_SAMPLER_PARAM=1
```

## Polling

Tracestore maintains knowledge of the state of the backend by polling it on regular intervals. There are currently only two components that need this knowledge and, consequently, only two that poll the backend: compactors and queriers.

Refer to [Use polling to monitor the backend status]({{< relref "./polling" >}}) for Tracestore.

## Dashboards

The [Tracestore mixin](https://example.com/acme/tracestore/tree/main/operations/tracestore-mixin) has four Acme dashboards in the `yamls` folder that you can download and import into your Acme UI.
These dashboards work well when you run Tracestore in a Kubernetes (k8s) environment and metrics scraped have the
`cluster` and `namespace` labels.

### Tracestore Reads dashboard

> This is available as `tracestore-reads.json`.

The Reads dashboard gives information on Requests, Errors, and Duration (RED) on the query path of Tracestore.
Each query touches the Gateway, Tracestore-Query, Query-Frontend, Queriers, Ingesters, the backend, and Cache, if present.

Use this dashboard to monitor the performance of each of the mentioned components and to decide the number of
replicas in each deployment.

### Tracestore Writes dashboard

> This is available as `tracestore-writes.json`.

The Writes dashboard gives information on RED on the write/ingest path of Tracestore.
A write query touches the Gateway, Distributors, Ingesters, and the backend.
This dashboard also gives information
on the number of operations performed by the Compactor to the backend.

Use this dashboard to monitor the performance of each of the mentioned components and to decide the number of
replicas in each deployment.

### Tracestore Resources dashboard

> This is available as `tracestore-resources.json`.

The Resources dashboard provides information on `CPU`, `Container Memory`, and `Go Heap Inuse`.
This dashboard is useful for resource provisioning for the different Tracestore components.

Use this dashboard to see if any components are running close to their assigned limits.

### Tracestore Operational dashboard

> This is available as `tracestore-operational.json`.

The Tracestore Operational dashboard deserves special mention because it is probably a stack of dashboard anti-patterns.
It's big and complex, doesn't use `jsonnet`, and displays far too many metrics in one place.
For just getting started, the RED dashboards are great places to learn how to monitor Tracestore in an opaque way.

This dashboard is included in the Tracestore repository for two reasons:

- The dashboard provides a stack of metrics for other operators to consider monitoring while running Tracestore.
- We want the dashboard in our internal infrastructure and we vendor the `tracestore-mixin` to do this.

## Rules and alerts

The Rules and Alerts are available as [YAML files in the compiled mixin](https://example.com/acme/tracestore/tree/main/operations/tracestore-mixin-compiled) on the repository.

To set up alerting, download the provided JSON files and configure them for use on your Prometheus monitoring server.

Check the [runbook](https://example.com/acme/tracestore/blob/main/operations/tracestore-mixin/runbook.md) to understand the
various steps that can be taken to fix firing alerts.