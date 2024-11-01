---
title: Troubleshoot Tracestore
menuTitle: Troubleshoot
description: Learn how to troubleshoot operational issues for Acme Tracestore.
weight: 700
aliases:
  - ../operations/troubleshooting/
---

# Troubleshoot Tracestore

This section helps with day zero operational issues that may come up when getting started with Tracestore.
The documents walk you through debugging each part of the ingestion and query pipeline to diagnose issues.

In addition, the [Tracestore runbook](https://example.com/acme/tracestore/blob/main/operations/tracestore-mixin/runbook.md) can help with remediating operational issues.

## Sending traces

- [Spans are being refused with "pusher failed to consume trace data"](https://acme.com/docs/tracestore/<TEMMPO_VERSION>/troubleshooting/max-trace-limit-reached/)
- [Is Acme Alloy sending to the backend?](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/alloy/)

## Querying

- [Unable to find my traces in Tracestore](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/unable-to-see-trace/)
- [Error message "Too many jobs in the queue"](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/too-many-jobs-in-queue/)
- [Queries fail with 500 and "error using pageFinder"](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/bad-blocks/)
- [I can search traces, but there are no service name or span name values available](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/search-tag)
- [Error message `response larger than the max (<number> vs <limit>)`](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/response-too-large/)
- [Search results don't match trace lookup results with long-running traces](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/long-running-traces/)

## Metrics-generator

- [Metrics or service graphs seem incomplete](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/metrics-generator/)
