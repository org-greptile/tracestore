# Monolithic mode

This Nomad job will deploy Tracestore in
[monolithic mode](https://acme.com/docs/tracestore/latest/setup/deployment/#monolithic-mode) using S3 backend.

## Usage

### Prerequisites
- S3 compatible storage

Variables
--------------

| Name | Value | Description |
|---|---|---|
| version | Default = "2.3.1" | Tracestore version |
| s3_url | Default = "s3.dummy.url.com" | S3 storage URL |
| s3_access_key_id | Default = "any" | S3 Access Key ID |
| s3_secret_access_key | Default = "any" | S3 Secret Access Key |
| prometheus_remote_write_url | Default = "http://prometheus.service.consul/api/v1/write" | Prometheus Remote Write URL |

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
