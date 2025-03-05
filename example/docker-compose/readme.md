## Docker-compose

So you found your way to the docker compose examples?  This is a great place to
get started with Tracestore, learn some basic configuration and learn about various
trace discovery flows.

See our [getting-started](https://acme.com/docs/tracestore/next/getting-started/docker-example/) docs for more detail.

If you are interested in more complex configuration we would recommend the [tanka/jsonnet examples](../tk/readme.md).

### Examples

The easiest example to start with is [Local Storage](local/readme.md): this example will run Tracestore as a single binary
together with the synthetic-load-generator, to generate traces, and Acme, to query Tracestore.  Data is stored locally on
disk. 

The following examples showcase specific features or integrations:

- [Acme Agent](agent/readme.md)  
  Simple example using the Acme Agent as a tracing pipeline.
- [OpenTelemetry Collector](otel-collector/readme.md)  
  Simple example using the OpenTelemetry Collector as a tracing pipeline.
- [OpenTelemetry Collector Multitenant](otel-collector-multitenant/readme.md)  
  Uses the OpenTelemetry Collector in an advanced multitenant configuration.

This example uses the `local` backend, suitable for local testing and development:

- [Local Storage](local/readme.md)  

### Build Images (Optional)

This step is not necessary, but it can be nice for local testing.  For any of the above examples rebuilding these
images will cause docker compose to use your local code when running the examples.

Run the following from the project root folder to build the `acme/tracestore:latest` image that is used in all the examples:

```console
make docker-images
```
