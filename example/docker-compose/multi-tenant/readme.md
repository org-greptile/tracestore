## Local Storage

In this example, all data is stored locally in the `tracestore-data` folder. Local storage is fine for experimenting with Tracestore
or when using the single binary, but doesn't work in a distributed or microservices scenario.

1. Start up the local stack.

```console
$ docker compose up -d
Starting multi-tenant_acme_1    ... done
Starting multi-tenant_tracestore_1      ... done
Starting multi-tenant_k6-tracing-2_1 ... done
Starting multi-tenant_k6-tracing_1   ... done
```

At this point, the following containers should be spun up:

```console
$ docker compose ps
           Name                          Command               State                                                                     Ports
------------------------------------------------------------------------------------------------------------------------------------------------------------
multi-tenant_acme_1        /run.sh                          Up      0.0.0.0:3000->3000/tcp,:::3000->3000/tcp
multi-tenant_k6-tracing-2_1   /k6-tracing run /example-s ...   Up
multi-tenant_k6-tracing_1     /k6-tracing run /example-s ...   Up
multi-tenant_tracestore_1          /tracestore -config.file=/etc/t ...   Up      0.0.0.0:14268->14268/tcp,:::14268->14268/tcp, 0.0.0.0:3200->3200/tcp,:::3200->3200/tcp, 0.0.0.0:4317->4317/tcp,:::4317->4317/tcp,
                                                                       0.0.0.0:4318->4318/tcp,:::4318->4318/tcp, 0.0.0.0:9095->9095/tcp,:::9095->9095/tcp, 0.0.0.0:9411->9411/tcp,:::9411->9411/tcp


```

2. If you're interested, you can see the wal/blocks as they are being created.

```console
$ ls tracestore-data/
```

3. Navigate to [Acme](http://localhost:3000/explore) select the Tracestore data source and use the "Search"
tab to find traces. Also notice that you can query Tracestore metrics from the Prometheus data source setup in
Acme.

4. Tail logs of a container (for example, tracestore):
```bash
$ docker logs multi-tenant_tracestore_1 -f
```

5. To stop the setup, use the following command:

```console
docker compose down -v
```

## Streaming and multi-tenant search

- Needs `stream_over_http_enabled: true`, `multitenancy_enabled: true`,
and `query_frontend.multi_tenant_queries_enabled: true` in the Tracestore configuration file, see `tracestore.yaml`

You can use Acme or tracestore-cli to make a query.

**gRPC streaming query using tracestore-cli**
- `$ tracestore-cli query api search "0.0.0.0:3200" --use-grpc --limit 10000 "{}" "2023-12-05T08:11:18Z" "2023-12-05T08:12:18Z" --org-id="test"`

**Multi-tenant streaming queries using tracestore-cli**
- Pass multiple tenant ids with `|` like this `--org-id="test|test2"`

Example:
```
$ ./bin/linux/tracestore-cli-amd64 query api search "0.0.0.0:3200" --use-grpc --limit 10000 "{ true } >> { true }" "2024-01-15T11:00:00Z" "2024-01-19T12:30:00Z" --org-id="test|test2"
```
