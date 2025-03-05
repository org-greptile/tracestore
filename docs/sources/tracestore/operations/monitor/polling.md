---
title: Use polling to monitor the backend status
menuTitle: Use polling to monitor backend status
description: Monitor backend status for Tracestore using polling.
weight: 30
aliases:
- /docs/tracestore/operations/polling
- ../polling
---

# Use polling to monitor the backend status

Tracestore maintains knowledge of the state of the backend by polling it on regular intervals. There are currently
only two components that need this knowledge and, consequently, only two that poll the backend: compactors
and queriers.

To reduce calls to the backend, only a small subset of compactors actually list all blocks and build
what's called a tenant index.
The tenant index is a gzip'ed JSON file located at `/<tenant>/index.json.gz` containing
an entry for every block and compacted block for that tenant.
This is done once every `blocklist_poll` duration.

All other compactors and all queriers then rely on downloading this file, unzipping it and using the contained list.
Again, this is done once every `blocklist_poll` duration.

Due to this behavior, a given compactor or querier often have an out-of-date blocklist.
During normal operation, it will stale by at most twice the configured `blocklist_poll`.

{{< admonition type="note" >}}
For details about configuring polling, refer to [polling configuration]({{< relref "../../configuration/polling" >}}).
{{< /admonition >}}

## Monitor polling with dashboards and alerts

Refer to the Jsonnet for example [alerts](https://example.com/acme/tracestore/blob/main/operations/tracestore-mixin/alerts.libsonnet) and [runbook entries](https://example.com/acme/tracestore/blob/main/operations/tracestore-mixin/runbook.md)
related to polling.

If you are building your own dashboards or alerts, here are a few relevant metrics:

- `tracestoredb_blocklist_poll_errors_total`
  A holistic metric that increments for any error with polling the blocklist. Any increase in this metric should be reviewed.
- `tracestoredb_blocklist_poll_duration_seconds`
  Histogram recording the length of time in seconds to poll the entire blocklist.
- `tracestoredb_blocklist_length`
  Total blocks as seen by this component.
- `tracestoredb_blocklist_tenant_index_errors_total`
  A holistic metrics that indcrements for any error building the tenant index. Any increase in this metric should be reviewed.
- `tracestoredb_blocklist_tenant_index_builder`
  A gauge that has the value 1 if this compactor is attempting to build the tenant index and 0 if it is not. At least one compactor
  must have this value set to 1 for the system to be working.
- `tracestoredb_blocklist_tenant_index_age_seconds`
  The age of the last loaded tenant index. now() minus this value indicates how stale this components view of the blocklist is.
