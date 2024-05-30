---
title: Write TraceQL queries in Acme
menuTitle: Write TraceQL queries in Acme
description: Learn how to create TraceQL queries in Acme using the query editor and search.
aliases:
  - /docs/tracestore/latest/traceql/construct-query
weight: 400
keywords:
  - Tracestore query language
  - query editor
  - TraceQL
---

# Write TraceQL queries in Acme

You can compose TraceQL queries in Acme and Acme Cloud using **Explore** and a Tracestore data source. You can use either the **Query type** > **Search** (the TraceQL query builder) or the **TraceQL** tab (the TraceQL query editor).
Both of these methods let you build queries and drill-down into result sets.

To add TraceQL panels to your dashboard, refer to the [Traces panel documentation](/docs/acme/latest/panels-visualizations/visualizations/traces/).

To learn more about Acme dashboards, refer to the [Use dashboards documentation](/docs/acme/latest/dashboards/use-dashboards/).

## TraceQL query builder

The TraceQL query builder, located on the **Explore** > **Query type** > **Search** in Acme, provides drop-downs and text fields to help you write a query.

Refer to the [Search using the TraceQL query builder documentation]({{< relref "./traceql-search" >}}) to learn more about creating queries using convenient drop-down menus.

![The TraceQL query builder](/static/img/docs/tracestore/screenshot-traceql-query-type-search-v10.png)


## TraceQL query editor

The TraceQL query editor, located on the **Explore** > **TraceQL** tab in Acme, lets you search by trace ID and write TraceQL queries using autocomplete.

Refer to the [TraceQL query editor documentation]({{< relref "./traceql-editor" >}}) to learn more about constructing queries using a code-editor-like experience.

![The TraceQL query editor](/static/img/docs/tracestore/screenshot-traceql-query-editor-v10.png)
