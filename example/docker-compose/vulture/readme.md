## Vulture

This example set up a local Tracestore instance and Tracestore vulture.

1. Create the storage directory with the correct permissions and start up the local stack.

```console
mkdir tracestore-data/
docker compose up -d
```

At this point, the following containers should be spun up:

```console
docker compose ps
```
```
NAME                IMAGE                          COMMAND                  SERVICE   CREATED         STATUS         PORTS
vulture-tracestore-1     acme/tracestore:latest           "/tracestore -config.file…"   tracestore     2 minutes ago   Up 2 minutes   0.0.0.0:3200->3200/tcp, 0.0.0.0:14250->14250/tcp
vulture-vulture-1   acme/tracestore-vulture:latest   "/tracestore-vulture -tem…"   vulture   2 minutes ago   Up 2 minutes  
```

2. If you're interested you can see the wal/blocks as they are being created.

```console
ls tracestore-data/
```

3. Tail logs of a container (eg: tracestore)
```bash
docker logs vulture_tracestore_1 -f
```

4. To stop the setup use:

```console
docker compose down -v
```

you can use Acme or tracestore-cli to make a query.

tracestore-cli: `$ tracestore-cli query api search "0.0.0.0:3200" --use-grpc "{}" "2023-12-05T08:11:18Z" "2023-12-05T08:12:18Z" --org-id="test"`
