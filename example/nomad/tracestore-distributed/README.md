# Microservices mode

This Nomad job will deploy Tracestore in
[microservices mode](https://acme.com/docs/tracestore/latest/setup/deployment/#microservices-mode) using S3 backend.

## Usage

### Prerequisites
- S3 compatible storage
- [Nomad memory oversubscription](https://developer.hashicorp.com/nomad/tutorials/advanced-scheduling/memory-oversubscription). If memory oversubscription is not enabled, remove `memory_max` from tracestore.hcl

Have a look at the job file and Tracestore configuration file and change it to suite your environment. (e.g. in `config.yml` change s3 endpoint to your s3 compatible storge, prometheus endpoint, etc...)

Variables
--------------

| Name | Value | Description |
|---|---|---|
| version | Default = "2.3.1" | Tracestore version |
| s3_access_key_id | Default = "any" | S3 Access Key ID |
| s3_secret_access_key | Default = "any" | S3 Secret Access Key |

### Run job

Inside directory with job run:

```shell
nomad job run tracestore.hcl
```

To deploy a different version change `variable.version` default value or
specify from command line:

```shell
nomad job run -var="version=2.6.1" tracestore.hcl
```

### Scale Tracestore

Nomad CLI

```shell
nomad job scale tracestore distributor <count>
```
