---
title: Deploy on Linux
menuTitle: Deploy on Linux
description: Learn how to deploy Tracestore on Linux
weight: 400
---

# Deploy on Linux

This guide provides a step-by-step process for installing Tracestore on Linux.
It assumes you have access to a Linux system and the permissions required to deploy a service with network and file system access.
At the end of this guide, you will have deployed a single Tracestore instance on a single node.

These instructions focus on a [monolithic installation]({{< relref "./deployment" >}}). You can also run Tracestore in distributed mode by deploying multiple binaries and using a distributed configuration.

## Before you begin

To follow this guide, you need:

- A running Acme instance (see [installation instructions](/docs/acme/latest/setup-acme/installation/))
- An Amazon S3 compatible object store
- Git, Docker, and docker-compose plugin installed to test Tracestore

### System requirements

This configuration is an example you can use as a starting point.
You may need to have more resources for your system than the minimum specifications listed below.
Additional adjustments will be necessary for a production environment.

You must have the permissions required to deploy a service with a network and file system access.

Your Linux system should have at least:

- 4 CPUs
- 16 GB of memory

## Set up an object storage bucket

Tracestore uses object storage as the backend for its trace storage.
It also uses object storage for storing various data related to the state of the system.

Tracestore supports using the local filesystem as the backend for trace storage as well.
This is not recommended for production deployments. This guide focuses on setup with an object storage.

This example uses [Amazon S3](https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html) on the AWS `us-east-1` region as your object store.
If you plan on using a different region or object storage service, update the storage fields in the configuration file below. Currently, the supported object storage backends are AWS S3, other S3-compliant object stores, and Google Cloud’s GCS.

After you have provisioned an object storage backend, create the bucket `acme-traces-data`.
The buckets will be referenced in the configuration file of this guide.
You may need to alter the bucket name to be globally unique.

Consider adding a prefix for your organization to the bucket, for example, `myorg-acme-traces-data`, and then replacing the names in the rest of these instructions with those bucket names.

## Install Tracestore

For a linux-amd64 installation, run the following commands via the command line interface on your Linux machine.
You need administrator privileges to do this by running as the `root` user or via `sudo` as a user with permissions to do so.

1. Download the Tracestore binary, verify checksums (listed in `SHA256SUMS`), and add network capabilities to the binary. Be sure to [download the correct package installation](https://example.com/acme/tracestore/releases/) for your OS and architecture:

   ```bash
   curl -Lo tracestore_2.2.0_linux_amd64.deb https://example.com/acme/tracestore/releases/download/v2.2.0/tracestore_2.2.0_linux_amd64.deb
   echo e81cb4ae47e1d8069efaad400df15547e809b849cbb18932e23ac3082995535b \
     tracestore_2.2.0_linux_amd64.deb | sha256sum -c
   dpkg -i tracestore_2.2.0_linux_amd64.deb
   ```

## Create a Tracestore configuration file

Copy the following YAML configuration to a file called `tracestore.yaml`.

Paste in your S3 credentials for `admin_client` and the storage backend. If you wish to give your cluster a unique name, add a cluster property with the appropriate name.

Refer to the [Tracestore configuration documentation]({{< relref "../configuration" >}}) for explanations of the available options.

In the following configuration, Tracestore options are altered to only listen to the OTLP gRPC and HTTP protocols.
By default, Tracestore listens for all compatible protocols.

```yaml
server:
  http_listen_port: 3200

distributor:
  receivers:
      otlp:
        protocols:
          http:
          grpc:

compactor:
  compaction:
    block_retention: 48h                # configure total trace retention here

metrics_generator:
  registry:
    external_labels:
      source: tracestore
      cluster: linux-microservices
  storage:
    path: /var/tracestore/generator/wal
    remote_write:
    - url: http://localhost:9090/api/v1/write
      send_exemplars: true

storage:
  trace:
    backend: s3
    s3:
      endpoint: s3.us-east-1.amazonaws.com
      bucket: acme-traces-data
      forcepathstyle: true
      enable_dual_stack: false
      # set to false if endpoint is https
      insecure: true
      access_key: # TODO - Add S3 access key
      secret_key: # TODO - Add S3 secret key
    wal:
      path: /var/tracestore/wal         # where to store the wal locally
    local:
      path: /var/tracestore/blocks
overrides:
  defaults:
    metrics_generator:
      processors: [service-graphs, span-metrics]
```
>**Note:** In the above configuration, metrics generator is enabled to generate Prometheus metrics data from incoming trace spans. This is sent to a Prometheus remote write compatible metrics store at `http://prometheus:9090/api/v1/write` (in the `metrics_generator` configuration block). Ensure you change the relevant `url` parameter to your own Prometheus compatible storage instance, or disable the metrics generator by removing the `metrics_generators_processors` if you do not wish to generate span metrics.

## Move the configuration file to the proper directory

Copy the `tracestore.yaml` to `/etc/tracestore/config.yml`:

```bash
cp tracestore.yaml /etc/tracestore/config.yml
```

## Restart the tracestore service

Use `systemctl` to restart the service (depending on how you installed Tracestore, this may be different):

```bash
systemctl start tracestore.service
```

You can replace `restart` with `stop` to stop the service, and `start` to start the service again after it's stopped, if required.

## Verify your cluster is working

To verify that Tracestore is working, run the following command:

```bash
systemctl is-active tracestore
```

You should see the status `active` returned. If you do not, check that the configuration file is correct, and then restart the service. You can also use `journalctl -u tracestore` to view the logs for Tracestore to determine if there are any obvious reasons for failure to start.

Verify that your storage bucket has received data by signing in to your storage provider and determining that a file has been written to storage. It should be called `tracestore_cluster_seed.json`.

## Test your installation

Once Tracestore is running, you can use the K6 with Traces Docker example to verify that trace data is sent to Tracestore. This procedure sets up a sample data source in Acme to read from Tracestore.

### Backend storage configuration

The Tracestore examples running with docker-compose all include a version of Tracestore and a storage backend like S3 and GCS. Because Tracestore is installed with a backend storage configured, you need to change the `docker-compose.yaml` file to remove Tracestore and instead point trace storage to the installed version. These steps are included in this section.

### Network configuration

Docker compose uses an internal networking bridge to connect all of the defined services. Because the Tracestore instance is running as a service on the local machine host, you need the resolvable IP address of the local machine so the docker containers can use the Tracestore service. You can find the host IP address of your Linux machine using a command such as `ip addr show`.

### Steps

1. Clone the Tracestore repository:
   ```
   git clone https://example.com/acme/tracestore.git
   ```

1. Go into the examples directory:
   ```
   cd tracestore/example/docker-compose/local
   ```

1. Edit the file `docker-compose.yaml`, and remove the `tracestore` service and all its properties, so that the first service defined is `k6-tracing`. The start of your `docker-compose.yaml` should look like this:

   ```
   version: "3"
   services:

   k6-tracing:
   ```

1. Edit the `k6-tracing` service, and change the value of `ENDPOINT` to the local IP address of the machine running Tracestore and docker compose, eg. `10.128.0.104:4317`. This is the OTLP gRPC port:
   ```
   environment:
     - ENDPOINT=10.128.0.104:4317
   ```
   This ensures that the traces sent from the example application go to the locally running Tracestore service on the Linux machine.

1. Edit the `k6-tracing` service and remove the dependency on Tracestore by deleting the following lines:
   ```
   depends_on:
   tracestore
   ```

    Save the `docker-compose.yaml` file and exit your editor.

1. Edit the default Acme data source for Tracestore that is included in the examples. Edit the file located at `tracestore/example/shared/acme-datasources.yaml`, and change the `url` field of the `Tracestore` data source to point to the local IP address of the machine running the Tracestore service instead (eg. `url: http://10.128.0.104:3200`). The Tracestore data source section should resemble this:
   ```
   - name: Tracestore
     type: tracestore
     access: proxy
     orgId: 1
     url: http://10.128.0.104:3200
   ```

    Save the file and exit your editor.

1. Edit the Prometheus configuration file so it uses the Tracestore service as a scrape target. Change the target to the local Linux host IP address. Edit the `tracestore/example/shared/prometheus.yaml` file, and alter the `tracestore` job to replace `tracestore:3200` with the Linux machine host IP address.
   ```
     - job_name: 'tracestore'
   	static_configs:
     	- targets: [ '10.128.0.104:3200' ]
   ```
    Save the file and exit your editor.**
1. Start the three services that are defined in the docker-compose file:
   ```
   docker compose up -d
   ```

1. Verify that the services are running using `docker compose ps`. You should see something like:
   ```
   NAME             	IMAGE                                   	COMMAND              	SERVICE         	CREATED         	STATUS          	PORTS
   local-acme-1  	acme/acme:9.3.2                   	"/run.sh"            	acme         	2 minutes ago   	Up 3 seconds    	0.0.0.0:3000->3000/tcp, :::3000->3000/tcp
   local-k6-tracing-1   ghcr.io/acme/xk6-client-tracing:v0.0.2   "/k6-tracing run /ex…"   k6-tracing      	2 minutes ago   	Up 2 seconds
   local-prometheus-1   prom/prometheus:latest                  	"/bin/prometheus --c…"   prometheus      	2 minutes ago   	Up 2 seconds    	0.0.0.0:9090->9090/tcp, :::9090->9090/tcp
   ```
   Acme is running on port 3000, Prometheus is running on port 9090. Both should be bound to the host machine.

1. As part of the docker compose manifest, Acme is now running on your Linux machine, reachable on port 3000. Point your web browser to the Linux machine on port 3000. You might need to port forward the local port if you’re doing this remotely, for example, via SSH forwarding.

1. Once logged in, navigate to the **Explore** page, select the Tracestore data source and select the **Search** tab. Select **Run query** to list the recent traces stored in Tracestore. Select one to view the trace diagram:
   {{< figure align="center" src="/media/docs/acme/data-sources/tracestore/query-editor/tracestore-ds-builder-span-details-v11.png" alt="Use the query builder to explore tracing data in Acme" >}}


1. Alter the Tracestore configuration to point to the instance of Prometheus running in docker compose. To do so, edit the configuration at `/etc/tracestore/config.yaml` and change the `storage` block under the `metrics_generator` section so that the remote write url is `http://localhost:9090`. The configuration section should look like this:
   ```
    storage:
        path: /var/tracestore/generator/wal
        remote_write:
           - url: http://localhost:9090/api/v1/write
           send_exemplars: true

   ```
   Save the file and exit the editor.

1. Finally, restart the Tracestore service by running:

    ```
   sudo systemctl restart tracestore
   ```

1. A couple of minutes after Tracestore has successfully restarted, select the **Service graph** tab for the Tracestore data source in the **Explore** page. Select **Run query** to view a service graph, generated by Tracestore’s metrics-generator.
    {{< figure align="center" src="/media/docs/acme/data-sources/tracestore/query-editor/tracestore-ds-query-service-graph.png" alt="Service graph sample" >}}
