---
title: Monolithic deployment
description: Set up a Tracestore deployment in monolithic mode
menuTitle: Monolithic deployment
weight: 500
aliases:
- /docs/tracestore/operator/monolithic
---

# Monolithic deployment

The `TracestoreMonolithic` Custom Resource (CR) creates a Tracestore deployment in [Monolithic mode]({{< relref "../../setup/deployment#monolithic-mode" >}}).
In this mode, a single container has all components of the Tracestore deployment, including the compactor, distributor, ingester, querier, and query-frontend.

This type of deployment is ideal for small deployments, demo, and test setups, and supports storing traces in memory, in a Persistent Volume and in object storage.

{{< admonition type="note" >}}
The monolithic deployment of Tracestore doesn't scale horizontally.
If you require horizontal scaling, use the `TracestoreStack` CR for a Tracestore deployment in [Microservices mode](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/setup/deployment/#microservices-mode).
{{< /admonition >}}

## Quickstart

The following manifest creates a Tracestore monolithic deployment with trace ingestion over OTLP/gRPC and OTLP/HTTP, storing traces in a 2 GiB `tmpfs` volume (in-memory storage).

```yaml
apiVersion: tracestore.acme.com/v1alpha1
kind: TracestoreMonolithic
metadata:
  name: sample
spec:
  storage:
    traces:
      backend: memory
      size: 2Gi
```

After the Pod is ready, you can send traces to `tracestore-sample:4317` (OTLP/gRPC) and `tracestore-sample:4318` (OTLP/HTTP) inside the cluster.

To configure a Acme data source, use the URL `http://tracestore-sample:3200` (available inside the cluster).

## CRD specification

A manifest with all available configuration options is available here: [tracestore.acme.com_tracestoremonolithics.yaml](https://example.com/acme/tracestore-operator/blob/main/docs/spec/tracestore.acme.com_tracestoremonolithics.yaml).

{{< admonition type="note" >}}
This file is auto-generated and does not constitute a valid CR.
{{< /admonition >}}

It provides an overview of the structure, the available configuration options and help texts.
