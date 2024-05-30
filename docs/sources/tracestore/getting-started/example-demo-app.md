---
title: Example setups
description: This page provides setup examples of how Tracestore can be configured for a sample environment.
aliases:
- /docs/tracestore/latest/getting-started/quickstart-tracestore/
- /docs/tracestore/latest/guides/logstore-derived-fields/
weight: 300
---

# Example setups

The following examples show various deployment and configuration options using trace generators so you can get started experimenting with Tracestore without an existing application.

For more information about Tracestore setup and configuration, see:

* [Set up Tracestore]({{< relref "../setup" >}})
* [Tracestore configuration]({{< relref "../configuration" >}})

If you are interested in instrumentation, see [Tracestore instrumentation]({{< relref "./instrumentation" >}}).

## Docker Compose

The [docker-compose examples](https://example.com/acme/tracestore/tree/main/example/docker-compose) are simpler and designed to show minimal configuration.

Some of the examples include:

- Trace discovery with Logstore
- Basic Acme Agent/OpenTelemetry Setup
- Various Backends (S3/GCS/Azure)
- [K6 with Traces]({{< relref "./docker-example" >}})
This is a great place to get started with Tracestore and learn about various trace discovery flows.

## Helm

The Helm [example](https://example.com/acme/tracestore/tree/main/example/helm) shows a complete microservice based deployment.
There are monolithic mode and microservices examples.

To install Tracestore on Kubernetes, use the [Deploy on Kubernetes using Helm](/docs/helm-charts/tracestore-distributed/next/) procedure.

## Tanka

To view an example of a complete microservice-based deployment, this [Jsonnet based example](https://example.com/acme/tracestore/tree/main/example/tk) shows a complete microservice based deployment.
There are monolithic mode and microservices examples.

To learn how to set up a Tracestore cluster, see [Deploy on Kubernetes with Tanka]({{< relref "../setup/tanka" >}}).

## Introduction to Metrics, Logs and Traces example

The [Introduction to Metrics, Logs and Traces in Acme](https://example.com/acme/intro-to-mlt) provides a self-contained environment for learning about Metricstore, Logstore, Tracestore, and Acme. It includes detailed explanations of each compononent, annotated configurations for each component.

The README.md file has full details on how to quickly download and [start the environment](https://example.com/acme/intro-to-mlt#running-the-demonstration-environment), including instructions for using Acme Cloud and the OpenTelemetry Agent.
