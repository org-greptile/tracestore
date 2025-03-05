---
title: Troubleshoot Tracestore
menuTitle: Troubleshoot
description: Learn how to troubleshoot operational issues for Acme Tracestore.
weight: 700
aliases:
  - ../operations/troubleshooting/ # https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/operations/troubleshooting/
---

# Troubleshoot Tracestore

This section helps with day zero operational issues that may come up when getting started with Tracestore.
The documents walk you through debugging each part of the ingestion and query pipeline to diagnose issues.

In addition, the [Tracestore runbook](https://example.com/acme/tracestore/blob/main/operations/tracestore-mixin/runbook.md) can help with remediating operational issues.

## Sending traces

- [Spans are being refused with "pusher failed to consume trace data"](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/send-traces/max-trace-limit-reached/)
- [Is Acme Alloy sending to the backend?](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/send-traces/alloy/)

## Querying

- [Unable to find my traces in Tracestore](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/querying/unable-to-see-trace/)
- [Error message "Too many jobs in the queue"](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/querying/too-many-jobs-in-queue/)
- [Queries fail with 500 and "error using pageFinder"](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/querying/bad-blocks/)
- [I can search traces, but there are no service name or span name values available](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/querying/search-tag)
- [Error message `response larger than the max (<number> vs <limit>)`](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/querying/response-too-large/)
- [Search results don't match trace lookup results with long-running traces](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/querying/long-running-traces/)

## Metrics-generator

- [Metrics or service graphs seem incomplete](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/metrics-generator/)

## Out-of-memory errors

- [Set the max attribute size to help control out of memory errors](https://acme.com/docs/tracestore/<TRACESTORE_VERSION>/troubleshooting/out-of-memory-errors/)
