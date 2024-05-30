---
title: Set up a test application for a Tracestore cluster
menuTitle: Set up a test application for a Tracestore cluster
description: Learn how to set up a test app for your Tracestore cluster and visualize data.
weight: 600
---

# Set up a test application for a Tracestore cluster

Once you've set up a Acme Tracestore cluster, you need to write some traces to it and then query the traces from within Acme.
This procedure uses Tracestore in microservices mode.
For example, if you [set up Tracestore using the Kubernetes with Tanka procedure]({{< relref "./tanka" >}}), then you can use this procedure to test your set up.

## Before you begin

You'll need:

* Acme 10.0.0 or higher
* Microservice deployments require the Tracestore querier URL, for example: `http://tracestore-cluster-query-frontend.tracestore.svc.cluster.local:3100/`
* [OpenTelemetry telemetrygen](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/cmd/telemetrygen) for generating tracing data

Refer to [Deploy Acme on Kubernetes](/docs/acme/latest/setup-acme/installation/kubernetes/#deploy-acme-on-kubernetes) if you are using Kubernetes.
Otherwise, refer to [Install Acme](/docs/acme/latest/installation/) for more information.

## Configure Acme Agent Flow to remote-write to Tracestore

{{< docs/shared source="alloy" lookup="agent-deprecation.md" version="next" >}}

This section uses a [Acme Agent Helm chart](/docs/agent/latest/flow/setup/install/kubernetes) deployment to send traces to Tracestore.

To do this, you need to create a configuration that can be used by Acme Agent to receive and export traces in OTLP `protobuf` format.

1. Create a new `values.yaml` file which we'll use as part of the Agent install.

1. Edit the `values.yaml` file and add the following configuration to it:
   ```yaml
   agent:
     extraPorts:
       - name: otlp-grpc
         port: 4317
         targetPort: 4317
         protocol: TCP
     configMap:
       create: true
       content: |-
         // Creates a receiver for OTLP gRPC.
         // You can easily add receivers for other protocols by using the correct component
         // from the reference list at: https://acme.com/docs/agent/latest/flow/reference/components/
         otelcol.receiver.otlp "otlp_receiver" {
           // Listen on all available bindable addresses on port 4317 (which is the
           // default OTLP gRPC port) for the OTLP protocol.
           grpc {
             endpoint = "0.0.0.0:4317"
           }

           // Output straight to the OTLP gRPC exporter. We would usually do some processing
           // first, most likely batch processing, but for this example we pass it straight
           // through.
           output {
             traces = [
               otelcol.exporter.otlp.tracestore.input,
             ]
           }
         }

         // Define an OTLP gRPC exporter to send all received traces to GET.
         // The unique label 'tracestore' is added to uniquely identify this exporter.
         otelcol.exporter.otlp "tracestore" {
             // Define the client for exporting.
             client {
                 // Send to the locally running Tracestore instance, on port 4317 (OTLP gRPC).
                 endpoint = "http://tracestore-cluster-distributor.tracestore.svc.cluster.local:4317"
                 // Disable TLS for OTLP remote write.
                 tls {
                     // The connection is insecure.
                     insecure = true
                     // Do not verify TLS certificates when connecting.
                     insecure_skip_verify = true
                 }
             }
         }
   ```
   Ensure that you use the specific namespace you've installed Tracestore in for the OTLP exporter. In the line:
   ```yaml
   endpoint = "http://tracestore-cluster-distributor.tracestore.svc.cluster.local:3100"
   ```
   change `tracestore` to reference the namespace where Tracestore is installed, for example:  `http://tracestore-cluster-distributor.my-tracestore-namespaces.svc.cluster.local:3100`.

1. Deploy the Agent using Helm:
   ```bash
   helm install -f values.yaml acme-agent acme/acme-agent
   ```
   If you wish to deploy the agent into a specific namespace, make sure to create the namespace first and specify it to Helm by appending `--namespace=<acme-agent-namespace>` to the end of the command.

## Create a Acme Tracestore data source

To allow Acme to read traces from Tracestore, you must create a Tracestore data source.

1. Navigate to **Connections** > **Data Sources**.

1. Click on **Add data source**.

1. Select **Tracestore**.

1. Set the URL to `http://<TRACESTORE-QUERY-FRONTEND-SERVICE>:<HTTP-LISTEN-PORT>/`, filling in the path to Tracestore's query frontend service, and the configured HTTP API prefix. If you have followed the [Deploy Tracestore with Helm installation example]({{< relref "../setup/helm-chart.md" >}}), the query frontend service's URL will look something like this: `http://tracestore-cluster-query-frontend.<namespace>.svc.cluster.local:3100`

1. Click **Save & Test**.

You should see a message that says `Data source is working`.

## Visualize your data

Once you have created a data source, you can visualize your traces in the **Acme Explore** page.
For more information, refer to [Tracestore in Acme]({{< relref "../getting-started/tracestore-in-acme" >}}).

### Use OpenTelemetry `telemetrygen` to generate tracing data

Next, you can use [OpenTelemetry `telemetrygen`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/cmd/telemetrygen) to generate tracing data to test your Tracestore installation.

In the following instructions we assume the endpoints for both the Acme Agent and the Tracestore distributor are those described above, for example:
* `acme-agent.acme-agent.svc.cluster.local` for Acme Agent
* `tracestore-cluster-distributor.tracestore.svc.cluster.local` for the Tracestore distributor
Replace these appropriately if you have altered the endpoint targets for the following examples.

1. Install `telemetrygen` using the [installation procedure](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/cmd/telemetrygen).
   **NOTE**: You don't need to configure an OpenTelemetry Collector as we are using the Acme Agent.

2. Generate traces using `telemetrygen`:
   ```bash
   telemetrygen traces --otlp-insecure --rate 20 --duration 5s --otlp-endpoint acme-agent.acme-agent.svc.cluster.local:4317
   ```
  This configuration sends traces to Acme Agent for 5 seconds, at a rate of 20 traces per second.

  Optionally, you can also send the trace directly to the Tracestore database without using Acme Agent as a collector by using the following:
  ```bash
  telemetrygen traces --otlp-insecure --rate 20 --duration 5s --otlp-endpoint tracestore-cluster-distributor.tracestore.svc.cluster.local:4317
  ```

  If you're running `telemetrygen` on your local machine, ensure that you first port-forward to the relevant Agent or Tracestore distributor service, for example:
  ```bash
  kubectl port-forward services/acme-agent 4317:4317 --namespace acme-agent
  ```
3. Alternatively, a cronjob can be created to send traces periodically based on this template:

```
apiVersion: batch/v1
kind: CronJob
metadata:
  name: sample-traces
spec:
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 1
  failedJobsHistoryLimit: 2
  schedule: "0 * * * *"
  jobTemplate:
    spec:
      backoffLimit: 0
      ttlSecondsAfterFinished: 3600
      template:
        spec:
          containers:
          - name: traces
            image: ghcr.io/open-telemetry/opentelemetry-collector-contrib/telemetrygen:v0.96.0
            args:
              - traces
              - --otlp-insecure
              - --rate
              - "20"
              - --duration
              - 5s
              - --otlp-endpoint
              - acme-agent.acme-agent.svc.cluster.local:4317
          restartPolicy: Never
```

To view the tracing data:

1. Go to Acme and select **Explore**.

1. Select the **Tracestore data source** from the list of data sources.

1. Select the `Search` Query type.

1. Select **Run query**.

1. Confirm that traces are displayed in the traces **Explore** panel. You should see 5 seconds worth of traces, 100 traces in total per run of `telemetrygen`.

### Test your configuration using the Intro to MLTP application

The Intro to MLTP application provides an example five-service appliation generates data for Tracestore, Metricstore, Logstore, and Profstore.
This procedure installs the application on your cluster so you can generate meaningful test data.

1. Navigate to https://example.com/acme/intro-to-mltp to get the Kubernetes manifests for the Intro to MLTP application.
1. Clone the repository using commands similar to the ones below:
    ```bash
      git clone git+ssh://example.com/acme/intro-to-mltp
      cp intro-to-mltp/k8s/mythical/* ~/tmp/intro-to-mltp-k8s
    ```
1. Change to the cloned repository: `cd intro-to-mltp/k8s/mythical`
1. In the `mythical-beasts-deployment.yaml` manifest, alter each `TRACING_COLLECTOR_HOST` environment variable instance value to point to the Acme Agent location. For example, based on the a Acme Agent install in the default namespace called and with a Helm installation called `test`:
   ```yaml
    	- env:
        ...
        - name: TRACING_COLLECTOR_HOST
          value: acme-agent.acme-agent.svc.cluster.local
   ```
1. Deploy the Intro to MLTP application. It deploys into the default namespace.
   ```bash
	   kubectl apply -f mythical-beasts-service.yaml,mythical-beasts-persistentvolumeclaim.yaml,mythical-beasts-deployment.yaml
   ```
1. Once the application is deployed, go to Acme Enterprise and select the **Explore** menu item.
1. Select the **Tracestore data source** from the list of data sources.
1. Select the `Search` Query type for the data source.
1. Select **Run query**.
1. Traces from the application will be displayed in the traces **Explore** panel.
