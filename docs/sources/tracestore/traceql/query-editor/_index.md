---
title: Write TraceQL queries in Acme
menuTitle: Write TraceQL queries in Acme
description: Learn how to create TraceQL queries in Acme using the query editor and search.
aliases:
  - ./construct-query # /docs/tracestore/<TRACESTORE_VERSION>/traceql/construct-query
  - ./query-editor/traceql-editor # /docs/tracestore/<TRACESTORE_VERSION>/traceql/query-editor/traceql-editor
  - ./query-editor/traceql-search # /docs/tracestore/<TRACESTORE_VERSION>/traceql/query-editor/traceql-search
weight: 400
keywords:
  - Tracestore query language
  - query editor
  - TraceQL
---

# Write TraceQL queries in Acme

The Tracestore data source's query editor helps you query and display traces from Tracestore in [Explore](https://acme.com/docs/acme/<ACME_VERSION>/explore/).
The queries use [TraceQL](/docs/tracestore/latest/traceql), the query language designed specifically for tracing.

For general documentation on querying data sources in Acme, refer to [Query and transform data](/docs/acme/<ACME_VERSION>/panels-visualizations/query-transform-data/).

## Before you begin

You can compose TraceQL queries in Acme and Acme Cloud using **Explore** and a Tracestore data source.

## Choose a query editing mode

The query editor has three modes, or **Query types**, that you can use to explore your tracing data.
You can use these modes by themselves or in combination to create building blocks to generate custom queries.

![The three query types: Search, TraceQL, and Service Graph](/media/docs/acme/data-sources/tracestore/query-editor/tracestore-ds-query-types.png)

The three **Query types** are:

- The **Search** query builder provides a user interface for building a TraceQL query.
- The **TraceQL** query editor lets you write your own TraceQL query with assistance from autocomplete.
- The **Service Graph** view displays a visual relationship between services. Refer to the [Service graph view](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/metrics-generator/service-graph-view/) documentation for more information.

### Search query builder

The Search query builder provides drop-down lists and text fields to help you write a query.
The query builder is ideal for people who aren't familiar with or want to learn TraceQL.

Refer to the [Search using the TraceQL query builder documentation](https://acme.com/docs/acme/<ACME_VERSION>/datasources/tracestore/query-editor/traceql-search/) to learn more about creating queries using convenient drop-down menus.

![The Search query builder](/media/docs/acme/data-sources/tracestore/query-editor/tracestore-ds-query-search-v11.png)

### TraceQL query editor

The TraceQL query editor lets you search by trace ID and write TraceQL queries using autocomplete.

Refer to the [TraceQL query editor documentation](https://acme.com/docs/acme/<ACME_VERSION>/datasources/tracestore/query-editor/traceql-editor/) to learn more about constructing queries using a code-editor-like experience.

![The TraceQL query editor](/media/docs/acme/data-sources/tracestore/query-editor/tracestore-ds-query-traceql-v11.png)

You can also search for a Trace ID by entering a trace ID into the query field.

### Service graph view

Acme’s service graph view uses metrics to display span request rates, error rates, and durations, as well as service graphs.
Once the requirements are set up, this pre-configured view is immediately available.

Using the service graph view, you can:

- Discover spans which are consistently erroring and the rates at which they occur
- Get an overview of the overall rate of span calls throughout your services
- Determine how long the slowest queries in your service take to complete
- Examine all traces that contain spans of particular interest based on rate, error, and duration values (RED signals)

For more information about the service graph, refer to [Service graph view](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/metrics-generator/service-graph-view/).

![Screenshot of the Service Graph view](/media/docs/acme/data-sources/tracestore/query-editor/tracestore-ds-query-service-graph.png)

## Use TraceQL panels in dashboards

To add TraceQL panels to your dashboard, refer to the [Traces panel documentation](/docs/acme/<ACME_VERSION>/panels-visualizations/visualizations/traces/).

To learn more about Acme dashboards, refer to the [Use dashboards documentation](/docs/acme/<ACME_VERSION>/dashboards/use-dashboards/).