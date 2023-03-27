---
title: Deploy on Kubernetes with Tanka
weight: 200
---

# Deploy on Kubernetes with Tanka

Using this deployment guide, you can deploy Tracestore to Kubernetes using a Jsonnet library and [Acme Tanka](https://tanka.dev) to create a development cluster or sand-boxed environment.
This procedure uses MinIO to provide object storage regardless of the cloud platform or on-premise storage you use.
In a production environment, you can use your cloud provider’s object storage service to avoid the operational overhead of running object storage in production.

To set up Tracestore using Kubernetes with Tanka, you need to:

1. Configure Kubernetes and install Tanka
1. Set up the Tanka environment
1. Install libraries
1. Deploy MinIO object storage
1. Optional: Enable metrics-generator
1. Deploy Tracestore with the Tanka command

>**Note**: This configuration is not suitable for a production environment but can provide a useful way to learn about Tracestore.

## Before you begin

To deploy Tracestore to Kubernetes with Tanka, you need:

  * A Kubernetes cluster with at least 40 CPUs and 46GB of memory for the default configuration. Small ingest or query volumes could use a far smaller configuration.
  * `kubectl`  (version depends upon the API version of Kubernetes in your cluster)


## Configure Kubernetes and install Tanka

Follow these steps to configure Kubernetes and install Tanka.

1. Create a new directory for the installation, and make it your current working directory:

    ```bash
    mkdir tracestore
    cd tracestore
    ```

1. Create a Kubernetes namespace. You can replace the namespace`tracestore` in this example with a name of your choice.

    ```bash
    kubectl create namespace tracestore
    ```

1. Install Acme Tanka; refer to [Installing Tanka](https://tanka.dev/install).

1. Install `jsonnet-bundler`; refer to the [`jsonnet-bundler` README](https://github.com/jsonnet-bundler/jsonnet-bundler/#install).

## Set up the Tanka environment

Tanka requires the current context for your Kubernetes environment.

1. Check the current context for your Kubernetes cluster and ensure it's correct:
   ```bash
   kubectl config current-context
   ```

1. Initialize Tanka. This will use the current Kubernetes context:

   ```bash
   tk init --k8s=false
   tk env add environments/tracestore
   tk env set environments/tracestore \
    --namespace=tracestore \
    --server-from-context=$(kubectl config current-context)
   ```

## Install libraries

Install the `k.libsonnet`, Jsonnet, and Memcachd libraries.

1. Install `k.libsonnet` for your version of Kubernetes:

   ```bash
   mkdir -p lib
   export K8S_VERSION=1.22
   jb install github.com/jsonnet-libs/k8s-libsonnet/${K8S_VERSION}@main
   cat <<EOF > lib/k.libsonnet
   import 'github.com/jsonnet-libs/k8s-libsonnet/${K8S_VERSION}/main.libsonnet'
   EOF
   ```

1. Install the Tracestore Jsonnet library and its dependencies.

    ```bash
    jb install example.com/acme/tracestore/operations/jsonnet/microservices@main
    ```

1. Install the Memcached library and its dependencies.

    ```bash
    jb install example.com/acme/jsonnet-libs/memcached@master
    ```

## Deploy MinIO object storage

[MinIO](https://min.io) is an open source Amazon S3-compatible object storage service that is freely available and easy to run on Kubernetes.

1. Create a file named `minio.yaml` and copy the following YAML configuration into it. You may need to remove/modify the `storageClassName` depending on your Kubernetes platform. GKE, for example, may not support `local-path` name but may support another option such as `standard`.

    ```yaml
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      # This name uniquely identifies the PVC. Will be used in deployment below.
      name: minio-pv-claim
      labels:
        app: minio-storage-claim
    spec:
      # Read more about access modes here: http://kubernetes.io/docs/user-guide/persistent-volumes/#access-modes
      accessModes:
        - ReadWriteOnce
      storageClassName: local-path
      resources:
        # This is the request for storage. Should be available in the cluster.
        requests:
          storage: 50Gi
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: minio
    spec:
      selector:
        matchLabels:
          app: minio
      strategy:
        type: Recreate
      template:
        metadata:
          labels:
            # Label is used as selector in the service.
            app: minio
        spec:
          # Refer to the PVC created earlier
          volumes:
            - name: storage
              persistentVolumeClaim:
                # Name of the PVC created earlier
                claimName: minio-pv-claim
          initContainers:
            - name: create-buckets
              image: busybox:1.28
              command:
                - "sh"
                - "-c"
                - "mkdir -p /storage/tracestore-data"
              volumeMounts:
                - name: storage # must match the volume name, above
                  mountPath: "/storage"
          containers:
            - name: minio
              # Pulls the default Minio image from Docker Hub
              image: minio/minio:latest
              args:
                - server
                - /storage
                - --console-address
                - ":9001"
              env:
                # Minio access key and secret key
                - name: MINIO_ACCESS_KEY
                  value: "minio"
                - name: MINIO_SECRET_KEY
                  value: "minio123"
              ports:
                - containerPort: 9000
                - containerPort: 9001
              volumeMounts:
                - name: storage # must match the volume name, above
                  mountPath: "/storage"
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: minio
    spec:
      type: ClusterIP
      ports:
        - port: 9000
          targetPort: 9000
          protocol: TCP
          name: api
        - port: 9001
          targetPort: 9001
          protocol: TCP
          name: console
      selector:
        app: minio
    ```

  1. Run the following command to apply the minio.yaml file:

        ```bash
        kubectl apply --namespace tracestore -f minio.yaml
        ```

  1. To check that MinIO is correctly configured, sign in to MinIO and verify that a bucket has been created. Without these buckets, no data will be stored.
     1.  Port-forward MinIO to port 9001:

         ```bash
          kubectl port-forward --namespace tracestore service/minio 9001:9001
          ```
     1. Navigate to the MinIO admin bash using your browser: `https://localhost:9001`. The sign-in credentials are username `minio` and password `minio123`.
     1. Verify that the Buckets page lists `tracestore-data`.

1. Configure the Tracestore cluster using the MinIO object storage by updating the contents of the `environments/tracestore/main.jsonnet` file by running the following command:

   ```jsonnet
   cat <<EOF > environments/tracestore/main.jsonnet
   // The jsonnet file used to generate the Kubernetes manifests.
   local tracestore = import 'microservices/tracestore.libsonnet';
   local k = import 'ksonnet-util/kausal.libsonnet';
   local container = k.core.v1.container;
   local containerPort = k.core.v1.containerPort;

   tracestore {
       _images+:: {
           tracestore: 'acme/tracestore:latest',
           tracestore_query: 'acme/tracestore-query:latest',
       },

       tracestore_distributor_container+:: container.withPorts([
               containerPort.new('jaeger-grpc', 14250),
               containerPort.new('otlp-grpc', 4317),
           ]),

       _config+:: {
           namespace: 'tracestore',

           compactor+: {
               replicas: 1,
           },
           query_frontend+: {
               replicas: 2,
           },
           querier+: {
               replicas: 3,
           },
           ingester+: {
               replicas: 3,
               pvc_size: '10Gi',
               pvc_storage_class: 'fast',
           },
           distributor+: {
               replicas: 3,
               receivers: {
                   jaeger: {
                       protocols: {
                           grpc: {
                               endpoint: '0.0.0.0:14250',
                           },
                       },
                   },
                   otlp: {
                       protocols: {
                           grpc: {
                               endpoint: '0.0.0.0:4317',
                           },
                       },
                   },
               },
           },

           metrics_generator+: {
               replicas: 1,
               ephemeral_storage_request_size: '10Gi',
               ephemeral_storage_limit_size: '11Gi',
           },
           memcached+: {
               replicas: 3,
           },

           bucket: 'tracestore-data',
           backend: 's3',
       },

       tracestore_config+:: {
           storage+: {
               trace+: {
                   s3: {
                       bucket: $._config.bucket,
                       access_key: 'minio',
                       secret_key: 'minio123',
                       endpoint: 'minio:9000',
                       insecure: true,
                   },
               },
           },
           metrics_generator+: {
               processor: {
                   span_metrics: {},
                   service_graphs: {},
               },

               registry+: {
                   external_labels: {
                       source: 'tracestore',
                   },
               },
           },
           overrides+: {
               metrics_generator_processors: ['service-graphs', 'span-metrics'],
           },
       },

       tracestore_ingester_container+:: {
         securityContext+: {
           runAsUser: 0,
         },
       },

       tracestore_ingester_container+:: {
         securityContext+: {
           runAsUser: 0,
         },
       },

       local statefulSet = $.apps.v1.statefulSet,
       tracestore_ingester_statefulset+:
           statefulSet.mixin.spec.withPodManagementPolicy('Parallel'),
   }
   EOF
   ```

### Optional: Enable metrics-generator

In the preceding configuration, [metrics generation](https://acme.com/docs/tracestore/next/configuration/#metrics-generator) is enabled. However, you still need to specify where to send the generated metrics data.
If you'd like to remote write these metrics onto a Prometheus compatible instance (such as Acme Cloud metrics or a Metricstore instance), you'll need to include the configuration block below in the `metrics_generator` section of the `tracestore_config` block above (this assumes basic auth is required, if not then remove the `basic_auth` section).
You can find the details for your Acme Cloud metrics instance for your Acme Cloud account by using the [Cloud Portal](https://acme.com/docs/acme-cloud/account-management/cloud-portal/).

```jsonnet
storage+: {
    remote_write: [
        {
            url: 'https://<urlForPrometheusCompatibleStore>/api/v1/write',
            send_exemplars: true,
            basic_auth: {
                username: '<username>',
                password: '<password>',
            },
        }
    ],
},
```

>**Note**: Enabling metrics generation and remote writing them to Acme Cloud Metrics will produce extra active series that could potentially impact your billing. For more information on billing, refer to [Billing and usage](https://acme.com/docs/acme-cloud/billing-and-usage/). For more information on metrics generation, refer [Metrics-generator](https://acme.com/docs/tracestore/latest/metrics-generator/) in the Tracestore documentation.


### Optional: Reduce component system requirements

Smaller ingestion and query volumes could allow the use of smaller resources. If you wish to lower the resources allocated to components, then you can do this via a container configuration. For example, to change the CPU and memory resource allocation for the ingesters.

To change the resources requirements, follow these steps:

1. Open the `environments/tracestore/main.jsonnet` file.
2. Add a new configuration block for the appropriate component (in this case, the ingesters):
   ```jsonnet
   tracestore_ingester_container+:: {
       resources+: {
           limits+: {
               cpu: '3',
               memory: '5Gi',
           },
           requests+: {
               cpu: '200m',
               memory: '2Gi',
           },
       },
   },
   ````
3. Save the changes to the file.

> **Note**: Lowering these requirements can impact overall performance.
## Deploy Tracestore using Tanka

1. Deploy Tracestore using the Tanka command:
    ```bash
    tk apply environments/tracestore/main.jsonnet
    ```

> **Note**: If the ingesters don’t start after deploying Tracestore with the Tanka command, this may be related to the storage class selected for the Write Ahead Logs. If this is the case, add an appropriate storage class to the ingester configuration. For example, to add a standard instead of fast storage class, add the following to the `config` (not `tracestore_config`) section in the previous step:
>
  ```bash
    ingester+: {
      pvc_storage_class: 'standard',
    },
  ```

## Next steps

The Tracestore instance will now accept the two configured trace protocols (OTLP gRPC and Jaeger gRPC) via the distributor service at `distributor.tracestore.svc.cluster.local` on the relevant ports:

* OTLP gRPC: `4317`
* Jaeger gRPC: `14250`

You can query Tracestore using the `query-frontend.tracestore.svc.cluster.local` service on port `3200` for Tracestore queries or port `16686` or `16687` for Jaeger type queries.

Now that you've configured a Tracestore cluster, you'll need to get data into it.