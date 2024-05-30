---
description: Learn how to create TraceQL queries in Acme using Explore > Search.
keywords:
  - acme
  - tracestore
  - traces
  - queries
menuTitle: Search traces
title: Search traces using TraceQL query builder
---

# Search traces using TraceQL query builder

Inspired by PromQL and LogQL, TraceQL is a query language designed for selecting traces.
TraceQL provides a method for formulating precise queries so you can zoom in to the data you need.
Query results are returned faster because the queries limit what is searched.

To learn more about how to query by TraceQL, refer to the [TraceQL documentation](/docs/tracestore/latest/traceql).

The TraceQL query builder, located on the **Explore** > **Query type** > **Search** in Acme, provides drop-downs and text fields to help you write a query.

![The TraceQL query builder](/static/img/docs/tracestore/screenshot-traceql-query-type-search-v10.png)

## Enable Search with the query builder

This feature is automatically available in Acme 10 (and newer) and Acme Cloud.

To enable the TraceQL query builder in self-hosted Acme through version 10.1, [enable the `traceqlSearch` feature toggle](/docs/acme/latest/setup-acme/configure-acme/feature-toggles/).

[//]: # "Shared content for the Search - TraceQL query builder"

{{< docs/shared source="acme" lookup="datasources/tracestore-search-traceql.md" leveloffset="+1" version="<ACME_VERSION>" >}}
