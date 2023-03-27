---
title: Querying with Acme
weight: 40
---

The way Acme queries Tracestore changed from 7.4.x to 7.5.x. This document aims to explain the difference between the two
and help you set up your datasources appropriately.

## Acme 7.5.x and higher (easy mode)

Acme 7.5.x and higher can query Tracestore directly. Point the Acme data source at your Tracestore query frontend (or monolithic mode Tracestore) and enter the URL: `http://<tracestore hostname>:<http port number>`. For most of [our examples](https://example.com/acme/tracestore/tree/main/example/docker-compose) the following works.

<p align="center"><img src="../ds75.png" alt="Acme 7.5.x datasource"></p>

Note that the port of 3200 is a common port used in our examples. Tracestore default for http is 80.


## Acme 7.4.x

Acme 7.4.x is *not* able to query Tracestore directly and requires the tracestore-query component as an intermediary. In this case
you need to run Tracestore-Query and direct it at Tracestore proper. Check out [the Acme 7.4.x example](https://example.com/acme/tracestore/tree/main/example/docker-compose/acme7.4) to help with configuration.

The url entered will be `http://<tracestore-query hostname>:16686/`.
