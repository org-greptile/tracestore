---
title: Set up Tracestore
menuTitle: Set up
description: Learn how to set up a Tracestore server or cluster and visualize data.
aliases:
- /docs/tracestore/setup
weight: 300
---

# Set up Tracestore

To set up Tracestore, you need to:

1. Plan your deployment
1. Deploy Tracestore
1. Test your installation
1. (Optional) Configure Tracestore services

## Plan your deployment

How you choose to deploy Tracestore depends upon your tracing needs.
Tracestore has two deployment modes: monolithic or microservices.

Read [Plan your deployment]({{< relref "./deployment" >}}) to determine the best method to deploy Tracestore.

## Deploy Tracestore

Once you have decided how to deploy Tracestore, you can install and set up Tracestore. For additional samples, refer to the [Example setups]({{< relref "../getting-started/example-demo-app" >}}) topic.

Acme Tracestore is available as a [pre-compiled binary, OS_specific packaging](https://example.com/acme/tracestore/releases), and [Docker image](https://example.com/acme/tracestore/tree/main/example/docker-compose).

The following procedures provide example Tracestore deployments that you can use as a starting point:

- [Deploy with Helm]({{< relref "./helm-chart" >}}) (microservices and monolithic)
- [Deploy with Tracestore Operator]({{< relref "./operator" >}}) (microservices)
- [Deploy on Linux]({{< relref "./linux" >}}) (monolithic)
- [Deploy on Kubernetes using Tanka]({{< relref "./tanka" >}}) (microservices)

You can also use Docker to deploy Tracestore using [the Docker examples](https://example.com/acme/tracestore/tree/main/example/docker-compose).

## Test your installation

Once Tracestore is deployed, you can test Tracestore by visualizing traces data:

- Using a [test application for a Tracestore cluster]({{< relref "./set-up-test-app" >}}) for the Kubernetes with Tanka setup
- Using a [Docker example]({{< relref "./linux" >}}) to test the Linux setup

These visualizations test Kubernetes with Tanka and Linux procedures. They do not check optional configuration you have enabled.

## (Optional) Configure Tracestore services

Explore Tracestore's features by learning about [available features and configurations]({{< relref "../configuration" >}}).

If you would to see a simplied, annotated example configuration for Tracestore, the [Introduction To MLT](https://example.com/acme/intro-to-mlt) example repository contains a [configuration](https://example.com/acme/intro-to-mlt/blob/main/tracestore/tracestore.yaml) for a monolithic instance.
