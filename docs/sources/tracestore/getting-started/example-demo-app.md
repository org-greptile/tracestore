---
title: Example setups
aliases:
- /docs/tracestore/latest/getting-started/quickstart-tracestore/
- /docs/tracestore/latest/guides/logstore-derived-fields/
weight: 300
---

# Example setups

The following examples show various deployment and configuration options using trace generators so you can get started experimenting with Tracestore without an existing application.

For more information about Tracestore setup and configuration, see:

* [Set up a Tracestore cluster]({{< relref "../setup">}})
* [Tracestore configuration]({{< relref "../configuration" >}})

If you are interested in instrumentation, see [Tracestore instrumentation]({{< relref "instrumentation" >}}).
## Docker Compose

The [docker-compose examples](https://example.com/acme/tracestore/tree/main/example/docker-compose) are simpler and designed to show minimal configuration.

Some of the examples include:

- Trace discovery with Logstore
- Basic Acme Agent/OpenTelemetry Setup
- Various Backends (S3/GCS/Azure)
- [K6 with Traces]({{< relref "docker-example" >}})

This is a great place to get started with Tracestore and learn about various trace discovery flows.

## Tanka

To view an example of a complete microservice-based deployment, this [Jsonnet based example](https://example.com/acme/tracestore/tree/main/example/tk) shows a complete microservice based deployment.
There are monolithic mode and microservices examples.

To learn how to set up a Tracestore cluster, see [Deploy on Kubernetes with Tanka]({{< relref "../setup/tanka" >}}).

## Helm

The Helm [example](https://example.com/acme/tracestore/tree/main/example/helm) shows a complete microservice based deployment.
There are monolithic mode and microservices examples.

To install Tracestore on Kubernetes, use the [Deploy on Kubernetes using Helm](/docs/helm-charts/tracestore-distributed/next/) procedure.

## The New Stack demo

The [New Stack (TNS) demo](https://example.com/acme/tns) demonstrates a fully instrumented three-tier application and the integration of Acme, Prometheus, Logstore, and Tracestore [features](https://example.com/acme/tns#demoable-things), including metrics to traces (exemplars), logs to traces, and traces to logs.

To learn how to set up a TNS app, see [Set up a test application for a Tracestore cluster]({{< relref "../setup/set-up-test-app" >}}).

A good place to start is the [docker-compose setup](https://example.com/acme/tns/tree/main/production/docker-compose) which includes a pre-built dashboard, load generator, and exemplars.

Explanation:
- Metrics To Traces (Exemplars)
  - The weaveworks middleware automatically [records](https://github.com/weaveworks/common/blob/bd288de53d57de300fa286688ce2fc935687213f/middleware/instrument.go#L79) request latency with an exemplar.  Try running the following PromQL query in Acme `Explore` and enabling the exemplars switch. It shows the p50 request latency for the "app" container:  `histogram_quantile(0.5, sum(rate(tns_request_duration_seconds_bucket{job="tns/app"}[$__rate_interval])) by (le))`.  Click the exemplar to see the trace.
- LogqlV2 and Logs to Traces
  - The http client [logs inter-service http requests](https://example.com/acme/tns/blob/main/client/http.go#L70) in `logfmt` format, which enables the ability to perform complex queries over api traffic. Try running the following query which shows all failed api requests from app to db and took longer than 100ms: `{job="tns/app"} | logfmt | level="info" and status>=500 and status <=599 and duration > 100ms`.  Expand the log line and click the Tracestore button near the trace ID to see the trace.
- Traces To Logs
  - When viewing only a trace in the Explore view (i.e. not side-by-side with logs), the Logs icon will appear next to each span.  Click it to view the matching logs.
- Status
  - Exemplar support in Prometheus is still pre-release so a custom image is used, and the feature is enabled with the `--enable-feature=exemplar-storage` command line parameter.
