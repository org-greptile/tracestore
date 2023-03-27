---
title: Set up a test app for a Tracestore cluster
menuTitle: Set up a test application for a Tracestore cluster
description: Learn how to set up a test app for your Tracestore cluster and visualize data.
weight: 400
---

# Set up a test application for a Tracestore cluster

Once you've set up a Acme Tracestore cluster, you need to write some traces to it and then query the traces from within Acme.

## Before you begin

You'll need:

* Acme 9.0.0 or higher
* Microservice deployments require the Tracestore querier URL, for example: `http://query-frontend.tracestore.svc.cluster.local:3200`

Refer to [Deploy Acme on Kubernetes](https://acme.com/docs/acme/latest/setup-acme/installation/kubernetes/#deploy-acme-on-kubernetes) if you are using Kubernetes.
Otherwise, refer to [Install Acme](https://acme.com/docs/acme/latest/installation/) for more information.

## Set up `remote_write` to your Tracestore cluster

To enable writes to your cluster:

1. Add a `remote_write` configuration snippet to the configuration file of an existing Acme Agent.

   If you do not have an existing traces collector, refer to [Set up with Acme Agent](https://acme.com/docs/agent/latest/set-up/).
   For Kubernetes, refer to the [Acme Agent Traces Kubernetes quick start guide](https://acme.com/docs/acme-cloud/kubernetes-monitoring/agent-k8s/k8s_agent_traces/).

   The example agent Kubernetes ConfigMap configuration below opens many trace receivers (note that the remote write is onto the Tracestore cluster using OTLP gRPC):

    ```yaml
    kind: ConfigMap
    metadata:
      name: acme-agent-traces
    apiVersion: v1
    data:
      agent.yaml: |
        traces:
            configs:
              - batch:
                    send_batch_size: 1000
                    timeout: 5s
                name: default
                receivers:
                    jaeger:
                        protocols:
                            grpc: null
                            thrift_binary: null
                            thrift_compact: null
                            thrift_http: null
                    opencensus: null
                    otlp:
                        protocols:
                            grpc: null
                            http: null
                    zipkin: null
                remote_write:
                  - endpoint: <tracestoreDistributorServiceEndpoint>
                    insecure: true  # only add this if TLS is not used
                scrape_configs:
                  - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
                    job_name: kubernetes-pods
                    kubernetes_sd_configs:
                      - role: pod
                    relabel_configs:
                      - action: replace
                        source_labels:
                          - __meta_kubernetes_namespace
                        target_label: namespace
                      - action: replace
                        source_labels:
                          - __meta_kubernetes_pod_name
                        target_label: pod
                      - action: replace
                        source_labels:
                          - __meta_kubernetes_pod_container_name
                        target_label: container
                    tls_config:
                        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
                        insecure_skip_verify: false
    ```

    If you have followed the [Tanka Tracestore installation example]({{< relref "../setup/tanka" >}}), then the `endpoint` value would be:

    ```bash
    distributor.tracestore.svc.cluster.local:4317
    ```

1. Apply the ConfigMap with:

    ```bash
    kubectl apply --namespace default -f agent.yaml
    ```

1. Deploy Acme Agent using the procedures from the relevant instructions above.

## Create a Acme Tracestore data source

To allow Acme to read traces from Tracestore, you must create a Tracestore data source.

1. Navigate to **Configuration ≫ Data Sources**.

1. Click on **Add data source**.

1. Select **Tracestore**.

1. Set the URL to `http://<TRACESTORE-HOST>:<HTTP-LISTEN-PORT>/`, filling in the path to your gateway and the configured HTTP API prefix. If you have followed the [Tanka Tracestore installation example]({{< relref "../setup/tanka.md" >}}), this will be: `http://query-frontend.tracestore.svc.cluster.local:3200/`

1. Click **Save & Test**.

You should see a message that says `Data source is working`.

If you see an error that says `Data source is not working: failed to get trace with id: 0`, check your Acme version.

To fix the error, [upgrade your Acme to 9.0 or later](https://acme.com/docs/acme/latest/setup-acme/upgrade-acme/).

## Visualize your data

Once you have created a data source, you can visualize your traces in the **Acme Explore** page.
For more information, refer to [Tracestore in Acme]({{< relref "../getting-started/tracestore-in-acme" >}}).

### Test your configuration using the TNS application

You can use The New Stack (TNS) application to test Tracestore data.

1. Create a new directory to store the TNS manifests.
1. Navigate to `https://example.com/acme/tns/tree/main/production/k8s-yamls` to get the Kubernetes manifests for the TNS application.
1. Clone the repository using commands similar to the ones below (where `<targetDir>` is the directory you used to store the manifests):

    ```bash
      mkdir ~/tmp
      cd ~/tmp
      git clone git+ssh://example.com/acme/tns
      cp tns/production/k8s-yamls/* <targetDir>
    ```

1. Change to the new directory: `cd <targetDir>` .
1. In each of the `-dep.yaml` manifests, alter the `JAEGER_AGENT_HOST` to the Acme Agent location. For example, based on the above Acme Agent install:
   ```yaml
   env:
   - name: JAEGER_AGENT_HOST
     value: acme-agent-traces.default.svc.cluster.local
   ```
1. Deploy the TNS application. It will deploy into the default namespace.
   ```bash
	 kubectl apply -f app-svc.yaml,db-svc.yaml,loadgen-svc.yaml,app-dep.yaml,db-dep.yaml,loadgen-dep.yaml
   ```
1. Once the application is running, look at the logs for one of the pods (such as the App pod) and find a relevant trace ID. For example:
   ```bash
  	kubectl logs $(kubectl get pod -l name=app -o jsonpath="{.items[0].metadata.name}")
    level=debug traceID=50075ac8b434e8f7 msg="GET / (200) 1.950625ms"
    level=info msg="HTTP client success" status=200 url=http://db duration=1.297806ms traceID=2c2fd669c388e76
    level=debug traceID=2c2fd669c388e76 msg="GET / (200) 1.70755ms"
    level=info msg="HTTP client success" status=200 url=http://db duration=1.853271ms traceID=79058bb9cc39acfb
    level=debug traceID=79058bb9cc39acfb msg="GET / (200) 2.300922ms"
    level=info msg="HTTP client success" status=200 url=http://db duration=1.381894ms traceID=7b0e0526f5958549
    level=debug traceID=7b0e0526f5958549 msg="GET / (200) 2.105263ms"
   ```
1. Go to Acme and select the **Explore** menu item.
1. Select the **Tracestore data source** from the list of data sources.
1. Copy the trace ID into the **Trace ID** edit field.
1. Select **Run query**.
1. Confirm that the trace is displayed in the traces **Explore** panel.
