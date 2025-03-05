---
title: Deploy Tracestore with Tracestore Operator
menuTitle: Deploy with operator
description: Learn how to deploy Tracestore with Tracestore Operator
weight: 375
aliases:
- /docs/tracestore/operator/operator
---

# Deploy Tracestore with Tracestore Operator

The Tracestore Operator allows you to configure, install, upgrade, and operate Acme Tracestore on Kubernetes and OpenShift clusters.

Some of the operator features are:

* **Resource Limits** - Specify overall resource requests and limits in the `TracestoreStack` CR; the operator assigns fractions of it to each component
* **AuthN and AuthZ** - Supports OpenID Control (OIDC) and role-based access control (RBAC)
* **Managed upgrades** - Updating the operator will automatically update all managed Tracestore clusters
* **Multitenancy** - Multiple tenants can send traces to the same Tracestore cluster
* **mTLS** - Communication between the Tracestore components can be secured via mTLS
* **Jaeger UI** - Traces can be visualized in Jaeger UI and exposed via Ingress or OpenShift Route
* **Observability** - The operator and `TracestoreStack` operands expose telemetry (metrics, traces) and integrate with Prometheus `ServiceMonitor` and `PrometheusRule`

The source of the Tracestore Operator can be found at [acme/tracestore-operator](https://example.com/acme/tracestore-operator).

## Installation

The operator can be installed from:
* [Kubernetes manifest](https://example.com/acme/tracestore-operator/releases/latest/download/tracestore-operator.yaml) file on a Kubernetes cluster
* [operatorhub.io](https://operatorhub.io/operator/tracestore-operator) on a Kubernetes cluster
* OperatorHub on an OpenShift cluster

## Compatibility

### Tracestore

The supported Tracestore version by the operator can be found in the [changelog](https://example.com/acme/tracestore-operator/blob/main/CHANGELOG.md) or on the [release page](https://example.com/acme/tracestore-operator/releases).

### Kubernetes

The Tracestore Operator is supported on Kubernetes versions v1.25 to v1.29.

### cert-manager

The operator Kubernetes manifest installation files use cert-manger `v1` custom resources to provision certificates for admission webhooks.

## Community

* Reach out to us on [#tracestore-operator](https://acme.slack.com/archives/C0414EUU39A) Acme Slack channel.
* Participate on [Tracestore community call]({{< relref "../../community" >}}).
