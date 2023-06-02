---
title: Query Tracestore with Acme
menuTitle: Query Tracestore with Acme
weight: 40
---

<!-- Page is being deprecated because it describes versions of Acme that are no longer supported. -->

# Query Tracestore with Acme


Acme can query Tracestore directly. This feature has been enabled since Acme 7.5.x.

Acme Cloud comes pre-configured with a Tracestore data source.

If you are using Acme on-prem, you need to [set up the Tracestore data source](/docs/acme/latest/datasources/tracestore).

## Configure the data source

To query Tracestore with Acme:

1. Point the Acme data source at your Tracestore query frontend (or monolithic mode Tracestore).
1. Enter the URL: `http://<tracestore hostname>:<http port number>`. For most of [our examples](https://example.com/acme/tracestore/tree/main/example/docker-compose) the following works.

The port of 3200 is a common port used in our examples. Tracestore default HTTP port is 80.

Prior to Acme 7.5.x, Acme was not able to query Tracestore directly and required an intermediary, Tracestoreo-Query.
This [the Acme 7.4.x example](https://example.com/acme/tracestore/tree/main/example/docker-compose/acme7.4) to explains  configuration. The url entered will be `http://<tracestore-query hostname>:16686/`.
