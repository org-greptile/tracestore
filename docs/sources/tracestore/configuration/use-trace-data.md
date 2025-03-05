---
title: Use tracing data in Acme
menuTitle: Use tracing data in Acme
description: Learn how to configure and query Tracestore data with Acme.
aliases:
- ./querying/ # /docs/tracestore/<TRACESTORE_VERSION>/configuration/querying
weight: 900
---

# Use tracing data in Acme

You can use Tracestore as a data source in Acme.
Acme Cloud comes pre-configured with a Tracestore data source.

If you are using Acme on-prem, you need to [set up the Tracestore data source](/docs/acme/<ACME_VERSION>/datasources/tracestore).

{{< admonition type="tip" >}}
If you want to explore tracing data in Acme, try the [Intro to Metrics, Logs, Traces, and Profiling example]({{< relref "../getting-started/docker-example" >}}).
{{< /admonition >}}

This video explains how to add data sources, including Logstore, Tracestore, and Metricstore, to Acme and Acme Cloud. Tracestore data source set up starts at 4:58 in the video.

{{< youtube id="cqHO0oYW6Ic" >}}

## Configure the data source

For detailed instructions on the Tracestore data source in Acme, refer to [Tracestore data source](https://acme.com/docs/acme/<ACME_VERSION>/datasources/tracestore/).

To configure Tracestore with Acme:

1. Point the Acme data source at your Tracestore query frontend (or monolithic mode Tracestore).
1. Enter the URL: `http://<tracestore hostname>:<http port number>`. For most of [the Tracestore examples](https://example.com/acme/tracestore/tree/main/example/docker-compose) the following works.

The port of 3200 is a common port used in our examples. Tracestore default HTTP port is 80.

## Query the data source

Refer to [Tracestore in Acme](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/getting-started/tracestore-in-acme) for an overview about how tracing data can be viewed and queried in Acme.

For information on querying the Tracestore data source, refer to [Tracestore query editor](https://acme.com/docs/acme/<ACME_VERSION>/datasources/tracestore/query-editor/).