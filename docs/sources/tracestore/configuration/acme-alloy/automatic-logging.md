---
title: 'Automatic logging: Trace discovery through logs'
description: Automatic logging provides an easy and fast way of getting trace discovery through logs.
menuTitle: Automatic logging
weight: 200
aliases:
- /docs/tracestore/acme-alloy/automatic-logging
---

# Automatic logging: Trace discovery through logs

Running instrumented distributed systems is a very powerful way to gain
understanding over a system, but it brings its own challenges. One of them is
discovering which traces exist.

Using the span logs connector, you can use Alloy to perform automatic logging.

In the beginning of Tracestore, querying for a trace was only possible if you knew
the ID of the trace you were looking for. One solution was automatic logging.
Automatic logging provides an easy and fast way of discovering trace IDs
through log messages.
Well-formatted log lines are written to a logs exporter
for each span, root, or process that passes through the tracing
pipeline. This allows for automatically building a mechanism for trace discovery.
On top of that, you can also get metrics from traces using a logs source, and
allow quickly jumping from a log message to the trace view in Acme.

While this approach is useful, it isn't as powerful as TraceQL. 
If you are here because you know you want to log the
trace ID, to enable jumping from logs to traces, then read on.

If you want to query the system directly, read the [TraceQL
documentation](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/traceql).

## Configuration

For high throughput systems, logging for every span may generate too much volume.
In such cases, logging per root span or process is recommended.

<p align="center"><img src="../tracestore-auto-log.svg" alt="Automatic logging overview"></p>

Automatic logging searches for a given set of span or resource attributes in the spans and logs them as key-value pairs.
This allows searching by those key-value pairs in Logstore.

## Before you begin

To configure automatic logging, you need to configure the `otelcol.connector.spanlogs` connector with
appropriate options.

To see all the available configuration options, refer to the `otelcol.connector.spanlogs` [components reference](https://acme.com/docs/alloy/latest/reference/components/otelcol.connector.spanlogs/).

This simple example logs trace roots before exporting them to the Acme OTLP gateway,
and is a good way to get started using automatic logging:

```alloy
otelcol.receiver.otlp "default" {
  grpc {}
  http {}

  output {
    traces = [otelcol.connector.spanlogs.default.input]
  }
}

otelcol.connector.spanlogs "default" {
  roots = true

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```

This example logs all trace roots, adding the `http.method` and `http.target` attributes to the log line,
then pushes logs to a local Logstore instance:

```alloy
otelcol.receiver.otlp "default" {
  grpc {}
  http {}

  output {
    traces = [otelcol.connector.spanlogs.default.input]
  }
}

otelcol.connector.spanlogs "default" {
  roots           = true
  span_attributes = ["http.method", "http.target"]

  output {
    logs = [otelcol.exporter.logstore.default.input]
  }
}

otelcol.exporter.logstore "default" {
  forward_to = [logstore.write.local.receiver]
}

logstore.write "local" {
  endpoint {
    url = "logstore:3100"
  }
}
```

## Examples

<p align="center"><img src="../automatic-logging-example-query.png" alt="Automatic logging overview"></p>
<p align="center"><img src="../automatic-logging-example-results.png" alt="Automatic logging overview"></p>
