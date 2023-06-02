---
title: Deploy Tracestore with Helm
menuTitle: Deploy with Helm
weight: 350
---

# Deploy Tracestore with Helm

The Helm charts for Acme Tracestore and Acme Enterprise Traces allows you to configure, install, and upgrade Acme Tracestore or Acme Enterprise Traces within a Kubernetes cluster.

The Tracestore repository has an [example Helm chart](https://example.com/acme/tracestore/tree/main/example/helm) that shows a complete microservice-based deployment.

Tracestore has two primary charts used for deployment:

* [`tracestore-distributed` Helm chart](https://example.com/acme/helm-charts/tree/main/charts/tracestore-distributed) deploys Tracestore in microservices mode
* [`tracestore` Helm chart](https://example.com/acme/helm-charts/tree/main/charts/tracestore) deploys Tracestore in monolithic (single binary) mode

To deploy Tracestore using the `tracestore-distributed` Helm chart, read the [Get started with Acme Tracestore using Helm](/docs/helm-charts/tracestore-distributed/next/get-started-helm-charts).
