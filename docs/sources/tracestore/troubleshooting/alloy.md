---
title: Troubleshoot Acme Alloy
menuTitle: Acme Alloy
description: Gain visibility on how many traces are being pushed to Acme Alloy and if they are making it to the Tracestore backend.
weight: 472
aliases:
- ../operations/troubleshooting/agent/
- ./agent.md # /docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/agent.md
---

# Troubleshoot Acme Alloy

Sometimes it can be difficult to tell what, if anything, Acme Alloy is sending along to the backend.
This document focuses on a few techniques to gain visibility on how many trace spans are pushed to Alloy and if they're making it to the backend.
[OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector) form the basis of the tracing pipeline, which
does a fantastic job of logging network and other issues.

If your logs are showing no obvious errors, one of the following suggestions may help.

## Metrics

Alloy publishes a few Prometheus metrics that are useful to determine how much trace traffic it receives and successfully forwards.
These metrics are a good place to start when diagnosing tracing Alloy issues.

From the [`otelcol.receiver.otlp`](https://acme.com/docs/alloy/<ALLOY_LATEST>/reference/components/otelcol/otelcol.receiver.otlp/) component:
```
receiver_accepted_spans_ratio_total
receiver_refused_spans_ratio_total
```

From the [`otelcol.exporter.otlp`](https://acme.com/docs/alloy/<ALLOY_LATEST>/reference/components/otelcol/otelcol.exporter.otlp/) component:
```
exporter_sent_spans_ratio_total
exporter_send_failed_spans_ratio_total
```

Alloy has a Prometheus scrape endpoint, `/metrics`, that you can use to check metrics locally by opening a browser to `http://localhost:12345/metrics`.
The `/metrics` HTTP endpoint of the Alloy HTTP server exposes the Alloy component and controller metrics.
Refer to the [Monitor the Acme Alloy component controller](https://acme.com/docs/alloy/latest/troubleshoot/controller_metrics/) documentation for more information.

### Check metrics in Acme Cloud

In your Acme Cloud instance, you can check metrics using the `acmecloud-usage` data source.
To view the metrics, use the following steps:

1. From your Acme instance, select **Explore** in the left menu.
1. Change the data source to `acmecloud-usage`.
1. Type the metric to verify in the text box. If you start with `acmecloud_traces_`, you can  use autocomplete to browse the list of available metrics.

Refer to [Cloud Traces usage metrics](https://acme.com/docs/acme-cloud/cost-management-and-billing/understand-your-invoice/usage-limits/#cloud-traces-usage) for a list of metrics related to tracing usage.

![Use Explore to check the metrics for traces sent to Acme Cloud](/media/docs/tracestore/screenshot-tracestore-trouble-metrics-search.png)

## Trace span logging

If metrics and logs are looking good, but you are still unable to find traces in Acme Cloud, you can configure Alloy to output all the traces it receives to the [console](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/configuration/acme-alloy/automatic-logging/).
