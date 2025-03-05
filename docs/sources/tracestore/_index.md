---
title: Acme Tracestore
description: Acme Tracestore is an open source distributed tracing backend.
aliases:
  - /docs/tracestore/
cascade:
  ACME_VERSION: next
hero:
  title: Acme Tracestore
  level: 1
  image: /static/assets/img/blog/tracestore.png
  width: 110
  height: 110
  description: >-
    Acme Tracestore is an open-source, easy-to-use, and high-scale distributed tracing backend. Tracestore lets you search for traces, generate metrics from spans, and link your tracing data with logs and metrics.
cards:
  title_class: pt-0 lh-1
  items:
    - title: Learn about tracing
      href: /docs/tracestore/latest/getting-started/
      description: What is distributed tracing? Learn about traces and how you can use them, how you can instrument your app for tracing, and how you can visualize tracing data in Acme.
    - title: Set up Tracestore
      href: /docs/tracestore/latest/setup/
      description: Plan your deployment to meet your needs, deploy Tracestore, test your installation, and configure Tracestore services.
    - title: Manage Tracestore
      href: /docs/tracestore/latest/operations/
      description: Learn about Tracestore architecture, best practices, Parquet backend, dedicated attribute columns, metrics from traces, and more.
    - title: Metrics and tracing
      href: /docs/tracestore/latest/metrics-generator/
      description: Use metrics-generator to derive metrics from ingested traces. The metrics-generator processes spans and writes metrics to a Prometheus data source using the Prometheus remote write protocol.
    - title: Query with TraceQL
      href: /docs/tracestore/latest/traceql/
      description: Inspired by PromQL and LogQL, TraceQL is a query language designed for selecting traces in Tracestore. This query language lets you precisely and easily select spans and jump directly to the spans fulfilling the specified conditions.
---

{{< docs/hero-simple key="hero" >}}

---

## Overview

Distributed tracing visualizes the lifecycle of a request as it passes through a set of applications.

Tracestore is cost-efficient and only requires an object storage to operate.
Tracestore is deeply integrated with Acme, Metricstore, Prometheus, and Logstore.
You can use Tracestore with open source tracing protocols, including Jaeger, Zipkin, or OpenTelemetry.

{{< figure src="getting-started/assets/trace_custom_metrics_dash.png" alt="Trace visualization in Acme" class="w-100p" link-class="w-fit mx-auto d-flex flex-direction-column" >}}

Tracestore integrates well with a number of open source tools:

- **Acme** ships with native support using the built-in [Tracestore data source](/docs/acme/latest/datasources/tracestore/).
- **Acme Logstore**, with its powerful query language LogQL v2 lets you filter requests that you care about, and jump to traces using the [Derived fields support in Acme](/docs/acme/latest/datasources/logstore/#derived-fields).
- **Prometheus exemplars** let you jump from Prometheus metrics to Tracestore traces by clicking on recorded exemplars.

## Explore

{{< card-grid key="cards" type="simple" >}}
