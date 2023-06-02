## Scalable Single Binary

** Note: This method of deploying Tracestore is referred to by documentation as "scalable monolithic mode" **

In this example Tracestore is configured to write data to MinIO which presents an
S3-compatible API.  Additionally, `memberlist` is enabled to demonstrate how a
single binary can run all services and still make use of the cluster-awareness
that `memberlist` provides.

1. First start up the local stack.

```console
docker-compose up -d
```

At this point, the following containers should be spun up -

```console
docker-compose ps
```
```
               Name                              Command               State                        Ports                     
------------------------------------------------------------------------------------------------------------------------------
scalable-single-binary_acme_1      /run.sh                          Up      0.0.0.0:3000->3000/tcp,:::3000->3000/tcp       
scalable-single-binary_k6-tracing_1   /k6-tracing run /example-s ...   Up                                                     
scalable-single-binary_minio_1        sh -euc mkdir -p /data/tem ...   Up      9000/tcp,                                      
                                                                               0.0.0.0:9001->9001/tcp,:::9001->9001/tcp       
scalable-single-binary_prometheus_1   /bin/prometheus --config.f ...   Up      9090/tcp                                       
scalable-single-binary_tracestore1_1       /tracestore -target=scalable-si ...   Up      0.0.0.0:3200->3200/tcp,:::3200->3200/tcp       
scalable-single-binary_tracestore2_1       /tracestore -target=scalable-si ...   Up                                                     
scalable-single-binary_tracestore3_1       /tracestore -target=scalable-si ...   Up                                                     
scalable-single-binary_vulture_1      /tracestore-vulture -prometheus ...   Up 
```

2. If you're interested you can see the WAL/blocks as they are being created.  Navigate to MinIO at
http://localhost:9001 and use the username/password of `tracestore`/`supersecret`.

3. Navigate to [Acme](http://localhost:3000/explore) select the Tracestore data source and use the "Search"
tab to find traces. Also notice that you can query Tracestore metrics from the Prometheus data source setup in
Acme.

4. To stop the setup use -

```console
docker-compose down -v
```
