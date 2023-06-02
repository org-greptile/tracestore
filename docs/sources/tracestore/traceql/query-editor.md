---
title: TraceQL query editor
menuTitle: TraceQL query editor
description: Learn how to use the TraceQL query editor
aliases:
  - /docs/tracestore/latest/traceql/construct-query
weight: 400
keywords:
  - Tracestore query language
  - query editor
  - TraceQL
---

# TraceQL query editor

You can use the TraceQL viewer and query editor in the Tracestore data source to build queries and drill-down into result sets. The editor is available in Acme’s Explore interface.

{{% admonition type="note" %}}
To use the TraceQL query editor in Acme, you need to enable the `traceqlEditor` feature flag. This feature is available starting in Acme 9.3.2. The query editory is available automatically in Acme Cloud.
{{% /admonition %}}

![Query editor showing request for http.method](/static/img/docs/tracestore/query-editor-http-method.png)

Using the query editor, you can use the editor’s autocomplete suggestions to write queries. The editor detects span sets to provide relevant autocomplete options. It uses regular expressions (regex) to detect where it is inside a spanset and provide attribute names, scopes, intrinsic names, logic operators, or attribute values from Tracestore's API, depending on what is expected for the current situation.

![Query editor showing the auto-complete feature](/static/img/docs/tracestore/query-editor-auto-complete.png)

Query results are returned in a table. Selecting the Trace ID or Span ID provides more detailed information.

![Query editor showing span results](/static/img/docs/tracestore/query-editor-results-span.png)

Selecting the trace ID from the returned results will open a trace diagram. Selecting a span from the returned results opens a trace diagram and reveals the relevant span in the trace diagram (above, the highlighted blue line).
