---
title: Set up a Tracestore server or cluster
menuTitle: Set up a Tracestore server or cluster
description: Learn how to set up a Tracestore server or cluster and visualize data
weight: 150
---

# Set up a Tracestore server or cluster

Tracestore is available as a pre-compiled binary, a Docker image, and as common OS-specific packaging.

This section describes how to set up Tracestore on a single Linux node or as a cluster using Kubernetes and Tanka. Tracestore can also be set up as a distributed set of services.

This page highlights these steps; more detailed instructions are available on the procedures for installing Tracestore.

## Deploy Tracestore

Choose a method to deploy Tracestore:

- [Deploy on Kubernetes using Helm](/docs/helm-charts/tracestore-distributed/next/)
- [Deploy on Linux]({{< relref "linux">}})
- [Deploy on Kubernetes using Tanka]({{< relref "tanka">}})

You can also use Docker to deploy Tracestore using [the Docker examples](https://example.com/acme/tracestore/tree/main/example/docker-compose).

## Test your installation

Once Tracestore is deployed, you can test Tracestore by visualizing traces data:

- Using a [test application for a Tracestore cluster]({{< relref "set-up-test-app" >}}) for the Kubernetes with Tanka setup
- Using a [Docker example]({{< relref "linux">}}) to test the Linux setup
