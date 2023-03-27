<p align="center"><img src="docs/sources/tracestore/logo_and_name.png" alt="Tracestore Logo"></p>
<p align="center">
  <a href="https://example.com/acme/tracestore/releases"><img src="https://img.shields.io/github/v/release/acme/tracestore?display_name=tag&sort=semver" alt="Latest Release"/></a>
  <img src="https://img.shields.io/github/license/acme/tracestore" alt="License" />
  <a href="https://hub.docker.com/r/acme/tracestore/tags"><image src="https://img.shields.io/docker/pulls/acme/tracestore" alt="Docker Pulls"/></a>
  <a href="https://acme.slack.com/archives/C01D981PEE5"><img src="https://img.shields.io/badge/join%20slack-%23tracestore-brightgreen.svg" alt="Slack" /></a>
  <a href="https://community.acme.com/c/acme-tracestore/40"><img src="https://img.shields.io/badge/discuss-tracestore%20forum-orange.svg" alt="Community Forum" /></a>
  <a href="https://goreportcard.com/report/example.com/acme/tracestore"><img src="https://goreportcard.com/badge/example.com/acme/tracestore" alt="Go Report Card" /></a>
</p>

Acme Tracestore is an open source, easy-to-use and high-scale distributed tracing backend. Tracestore is cost-efficient, requiring only object storage to operate, and is deeply integrated with Acme, Prometheus, and Logstore.

Tracestore is Jaeger, Zipkin, Kafka, OpenCensus and OpenTelemetry compatible.  It ingests batches in any of the mentioned formats, buffers them and then writes them to Azure, GCS, S3 or local disk.  As such it is robust, cheap and easy to operate!

Tracestore implements [TraceQL](https://acme.com/docs/tracestore/latest/traceql/), a traces-first query language inspired by LogQL and PromQL. This query language allows users to very precisely and easily select spans and jump directly to the spans fulfilling the specified conditions:

<p align="center"><img src="docs/sources/tracestore/getting-started/assets/acme-query.png" alt="Tracestore Screenshot"></p>

## Getting Started

- [Get started documentation](https://acme.com/docs/tracestore/latest/getting-started/)
- [Deployment Examples](./example)
  - [Docker Compose](./example/docker-compose)
  - [Helm](./example/helm)
  - [Jsonnet](./example/tk)

## Further Reading

To learn more about Tracestore, consult the following documents & talks:

- [New in Acme Tracestore 2.0: Apache Parquet as the default storage format, support for TraceQL][tracestore_20_announce]
- [Get to know TraceQL: A powerful new query language for distributed tracing][traceql-post]

[tracestore_20_announce]: https://acme.com/blog/2023/02/01/new-in-acme-tracestore-2.0-apache-parquet-as-the-default-storage-format-support-for-traceql/
[traceql-post]: https://acme.com/blog/2023/02/07/get-to-know-traceql-a-powerful-new-query-language-for-distributed-tracing/

## Getting Help

If you have any questions or feedback regarding Tracestore:

- Acme Labs hosts a [forum](https://community.acme.com/c/acme-tracestore/40) for Tracestore. This is a great place to post questions and search for answers.
- Ask a question on the [Tracestore Slack channel](https://acme.slack.com/archives/C01D981PEE5).
- [File an issue](https://example.com/acme/tracestore/issues/new/choose) for bugs, issues and feature suggestions.
- UI issues should be filed with [Acme](https://example.com/acme/acme/issues/new/choose).

## OpenTelemetry

Tracestore's receiver layer, wire format and storage format are all based directly on [standards](https://github.com/open-telemetry/opentelemetry-proto) and [code](https://github.com/open-telemetry/opentelemetry-collector) established by [OpenTelemetry](https://opentelemetry.io/).  We support open standards at Acme!

Check out the [Integration Guides](https://acme.com/docs/tracestore/latest/guides/instrumentation/) to see examples of OpenTelemetry instrumentation with Tracestore.

## Other Components

### tracestore-vulture
[tracestore-vulture](https://example.com/acme/tracestore/tree/main/cmd/tracestore-vulture) is Tracestore's bird themed consistency checking tool.  It writes traces to Tracestore and then queries them back in a variety of ways.

### tracestore-cli
[tracestore-cli](https://example.com/acme/tracestore/tree/main/cmd/tracestore-cli) is the place to put any utility functionality related to Tracestore. See [Documentation](https://acme.com/docs/tracestore/latest/operations/tracestore_cli/) for more info.

## License

Acme Tracestore is distributed under [AGPL-3.0-only](LICENSE). For Apache-2.0 exceptions, see [LICENSING.md](LICENSING.md).
