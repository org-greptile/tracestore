---
title: Set up monitoring for Tracestore
menuTitle: Set up monitoring
description: Set up monitoring for Tracestore
weight: 20
---

# Set up monitoring for Tracestore

You can set up monitoring for Tracestore using an existing or new cluster.
If you don't have a cluster available, you can use the linked documentation to set up the Tracestore, Metricstore, and Acme using Helm or you can use Acme Cloud.

To set up monitoring, you need to:

* Use Acme Agent Flow or Acme Alloy to remote-write to Tracestore and set up Acme to visualize the tracing data by following [Set up a test app](https://acme.com/docs/tracestore/latest/setup/set-up-test-app/).
* Update your Acme Agent Flow configuration to scrape metrics to monitor for your Tracestore data.

This procedure assumes that you have set up Tracestore [using the Helm chart](https://acme.com/docs/tracestore/latest/setup/helm-chart/) and with [Acme Agent](https://acme.com/docs/agent/latest/flow/) or [Acme Alloy](https://acme.com/docs/alloy/latest/).

{{< docs/shared source="alloy" lookup="agent-deprecation.md" version="next" >}}

The steps outlined below use the Acme Agent Flow configurations described in [Set up a test application for a Tracestore cluster](https://acme.com/docs/tracestore/latest/setup/set-up-test-app/).

{{< admonition type="note" >}}
Update any instructions in this document for your own deployment.

If you use the [Kubernetes integration Acme Agent Helm chart](https://acme.com/docs/agent/latest/flow/get-started/install/kubernetes/), you’ll be able to use the Kubernetes scrape annotations to automatically scrape Tracestore. You’ll need to add the labels to all of the deployed components.
{{% /admonition %}}


## Before you begin

To configure monitoring using the examples on this page, you’ll need the following running in your Kubernetes environment:

* Tracestore instance - For storing traces and emitting metrics ([install using the `tracestore-distributed` Helm chart](https://acme.com/docs/tracestore/latest/setup/helm-chart/))
* Metricstore - For storing metrics emitted from Tracestore ([install using the `metricstore-distributed` Helm chart](https://acme.com/docs/helm-charts/metricstore-distributed/latest/get-started-helm-charts/))
* Acme - For visualizing traces and metrics ([install on Kubernetes](https://acme.com/docs/acme/latest/setup-acme/installation/kubernetes/#deploy-acme-oss-on-kubernetes))

You can use Acme Agent or the OpenTelemetry Collector. This procedure provides examples only for Acme Agent Flow.

The rest of this documentation assumes that the Tracestore, Acme, and Metricstore instances use the same Kubernetes cluster.

If you are using Acme Cloud, you can skip the installation sections and set up the [Metricstore (Prometheus)](https://acme.com/docs/acme-cloud/connect-externally-hosted/data-sources/prometheus/) and [Tracestore data sources](https://acme.com/docs/acme-cloud/connect-externally-hosted/data-sources/tracestore/) in your Acme instance.

## Use a test app for Tracestore to send data to Acme

Before you can monitor Tracestore data, you need to configure the Acme Agent to send traces to Tracestore.

Use [these instructions to create a test application](https://acme.com/docs/tracestore/latest/setup/set-up-test-app/) in your Tracestore cluster. These steps configure Acme Agent Flow to `remote-write` to Tracestore. In addition, the test app instructions explain how to configure a Tracestore data source in Acme and view the tracing data.

{{< admonition type="note" >}}
If you already have a Tracestore environment, then you do not need to create a test app.
This guide assumes that the Tracestore and Acme Agent configurations are the same as or based on [these instructions to create a test application](https://acme.com/docs/tracestore/latest/setup/set-up-test-app/), as you'll augment those configurations to enable Tracestore metrics monitoring.
{{% /admonition %}}

In these examples, Tracestore is installed in a namespace called `tracestore`.
Change this namespace name in the examples as needed to fit your own environment.

## Configure Acme

In your Acme instance, you'll need:

* [A Tracestore data source](https://acme.com/docs/acme/latest/datasources/tracestore/configure-tracestore-data-source/) (created in the previous section)
* A [Metricstore (Prometheus) data source](https://acme.com/docs/acme/latest/datasources/prometheus/)

## Enable Tracestore metrics scraping

Tracestore exposes Prometheus metrics from all of its components to allow meta-monitoring.
To retrieve these metrics, a suitable scraper needs to be configured.
Acme Agent can collect traces and act as a Prometheus scraper. To use this capability, you need to configure the Agent to scrape from all of the components.

Acme Agent lets you discover targets to scrape in a cluster via a variety of ways.
Usually for Prometheus metrics scraping, you would annotate the pods, services, and others, to signal that Acme Agent should scrape metrics from those objects using annotations such as `prometheus/scrape: true` as well as a port and path.

However, the Tracestore objects already have some convenient annotations supplied under the `app.kubernetes.io` prefixes.
The Helm deployment includes these annotations.
For example, the label annotations for Tracestore’s distributor component contain:

```yaml
app.kubernetes.io/component=distributor
app.kubernetes.io/instance=tracestore
app.kubernetes.io/managed-by=Helm
app.kubernetes.io/name=tracestore
app.kubernetes.io/part-of=memberlist
app.kubernetes.io/version=2.4.1
```

Because of this, you can use Kubernetes service discovery in Acme Agent using these annotations to ensure that metrics are scraped from each Tracestore component.
Using the [`discovery.kubernetes` Flow component](https://acme.com/docs/agent/latest/flow/reference/components/discovery.kubernetes/), you can include selectors to scrape from targets based on labels, for example.
As there are Tracestore-specific component labels, you can specify a rule that covers all of the Tracestore components.

```yaml
discovery.kubernetes "k8s_pods" {
  role = "pod"
  selectors {
    // Only scrape pods with a particular selector.
    role = "pod"
    // The selector is any component that belongs to Tracestore.
    label = "app.kubernetes.io/name=tracestore"
  }
}
```

This rule lets Acme Agent know that only pods that include an annotation of `app.kubernetes.io/name=tracestore` should have their metrics scraped.

Something to note is that the dashboards for Tracestore expect the specific namespace, cluster name, and job name for labels included with the scraped Tracestore metrics, where the job nomenclature is `<namespace>>app.kubernetes.io/component>`.
The latter expands to `distributor`, `compactor`, `ingester`, and so on.
For example: `tracestore/compactor`, where `tracestore` is the namespace and `compactor` replaces `<app.kubernetes.io/component>`.

You can acquire all three labels from the annotations that are included as Kubernetes metadata from each components scrape data.
Because this metadata is a string with a prefix of `__meta_kubernetes`, Acme Agent must be configured not to discard these labels during the metrics pipeline.
To do this, you can use the [`discover.relabel` component](https://acme.com/docs/agent/latest/flow/reference/components/discovery.relabel/) to add rules to keep any required metadata labels.

In the following example, the specific Tracestore component name (for example, `__meta_kubernetes_app_kubernetes_io_component`) is replaced by a more useful label name and kept for future pipeline operations.

```yaml
rule {
  source_labels = ["__meta_kubernetes_pod_label_app_kubernetes_io_component"]
  action = "replace"
  regex = "(.*)"
  replacement = "$1"
  target_label = "k8s_component_name"
}
```

By adding rules to keep the `__meta_kubernetes_namespace` label for namespace and `__meta_kubernetes_app_kubernetes_io_instance` label for cluster name, you can also attach these labels to written metrics for dashboard use later.

Finally, including a Prometheus relabel component to the Flow configuration allows the use of the kept metadata labels to create the correct nomenclature for the job:

```yaml
rule {
  source_labels = ["namespace", "k8s_component_name"]
  action = "replace"
  regex = "(.*?);(.*?)"
  replacement = "$1/$2"
  target_label = "job"
}
```

This lets you create a configuration that scrapes metrics from Tracestore components and writes the data to a Metricstore instance of your choice.

Here’s a complete configuration for a [Acme Agent Helm](https://acme.com/docs/agent/latest/flow/get-started/install/kubernetes/) values file for Flow, to scrape a running instance of Tracestore.

```yaml
// Scrape Prometheus metrics for Tracestore.
prometheus.scrape "tracestore" {
  // Use Kubernetes discovery to find the relevant pods to scrape.
  targets    = discovery.relabel.k8s_pods.output
  // Forward to the Prometheus relabeling component.
  forward_to = [prometheus.relabel.tracestore.receiver]
}

// Determine how to select pods to scrape.
discovery.kubernetes "k8s_pods" {
  // Only scrape pods.
  role = "pod"
  selectors {
    // Only scrape pods with a particular selector.
    role = "pod"
    // The selector is any component that belongs to Tracestore.
    label = "app.kubernetes.io/name=tracestore"
  }
}

// Relabel data from Kubernetes discovery.
discovery.relabel "k8s_pods" {
  // Relabel from targets scraped by the discovery selection.
  targets = discovery.kubernetes.k8s_pods.targets

  // Create new namespace label based on the discovered kubernetes namespace.
  rule {
    source_labels = ["__meta_kubernetes_namespace"]
    action = "replace"
    regex = "(.*)"
    replacement = "$1"
    target_label = "namespace"
  }

  // Create new component label based on the discovered kubernetes component.
  rule {
    source_labels = ["__meta_kubernetes_pod_label_app_kubernetes_io_component"]
    action = "replace"
    regex = "(.*)"
    replacement = "$1"
    target_label = "k8s_component_name"
  }

  // Create new cluster label based on the discovered kubernetes instance.
  rule {
    source_labels = ["__meta_kubernetes_pod_label_app_kubernetes_io_instance"]
    action = "replace"
    regex = "(.*)"
    replacement = "$1"
    target_label = "cluster"
  }
}

// Relabel data from Prometheus scraping.
prometheus.relabel "tracestore" {
  // Replace the existing job label with one comprised of the namespace and component.
  rule {
    source_labels = ["namespace", "k8s_component_name"]
    action = "replace"
    regex = "(.*?);(.*?)"
    replacement = "$1/$2"
    target_label = "job"
  }


  // Send the metrics to the Prometheus remote write component.
  forward_to = [prometheus.remote_write.tracestore.receiver]
}

// Remote write the metrics to a Prometheus compatible endpoint (in this case Metricstore).
prometheus.remote_write "tracestore" {
  endpoint {
    url = "https://metricstore-cluster.distributor.metricstore.svc.cluster.local:9001/api/v1/push"
  }
}
```

This example doesn’t include ingestion for any other data such as traces for sending to Tracestore, but can be included with some configuration updates.
Refer to [Configure Acme Agent Flow to remote-write to Tracestore](https://acme.com/docs/tracestore/latest/setup/set-up-test-app/) for more information.

## Install Tracestore dashboards in Acme

After metrics from Tracestore are scraped by Acme Agent and stored in Metricstore or another Prometheus compatible time-series database, you can monitor Tracestore’s operation using the mixin.

Tracestore ships with a mixin that includes:

* Relevant dashboards for overseeing the health of Tracestore as a whole, as well as its individual components
* Recording rules that simplify the generation of metrics for dashboards and free-form queries
* Alerts that trigger when Tracestore falls out of operational parameters

To install the mixins in Acme, you need to:

1. Download the mixin dashboards from the Tracestore repository.

1. Import the dashboards in your Acme instance.

1. Upload `alerts.yaml` and `rules.yaml` files for Metricstore or Prometheus

### Download the `tracestore-mixin` dashboards

1. First, clone the Tracestore repository from Github:
   ```bash
   git clone git+ssh://example.com/acme/tracestore
   ```

1. Once you have a local copy of the repository, navigate to the `operations/tracestore-mixin-compiled` directory.
   ```bash
   cd operations/tracestore-mixin-compiled
   ```

This contains a compiled version of the alert and recording rules, as well as the dashboards.

{{< admonition type="note" >}}
If you want to change any of the mixins, make your updates in the `operations/tracestore-mixin` directory.
Use the instructions in the [README](https://example.com/acme/tracestore/tree/main/operations/tracestore-mixin) in that directory to regenerate the files.
The mixins are generated in the `operations/tracestore-mixin-compiled` directory.
{{% /admonition %}}

### Import the dashboards to Acme

The `dashboards` directory includes the six monitoring dashboards that can be installed into your Acme instance.
Refer to [Import a dashboard ](https://acme.com/docs/acme/latest/dashboards/build-dashboards/import-dashboards/)in the Acme documentation.

{{< admonition type="tip" >}}
Install all six dashboards.
You can only import one dashboard at a time.
Create a new folder in the Dashboards area, for example “Tracestore Monitoring”, as an easy location to save the imported dashboards.
{{% /admonition %}}

To create a folder:

1. Open your Acme instance and select **Dashboards**.
1. Select **New** in the right corner.
1. Select **New folder** from the **New** drop-down.
1. Name your folder, for example, “Tracestore Monitoring”.
1. Select **Create**.

To import a dashboard:

1. Open your Acme instance and select **Dashboards**.
1. Select **New** in the right corner.
1. Select **Import**.
1. On the **Import dashboard** screen, select **Upload.**
1. Browse to `operations/tracestore-mixin-compiled/dashboards` and select the dashboard to import.
1. Drag the dashboard file, for example, `tracestore-operational.json`, onto the **Upload** area of the **Import dashboard** screen. Alternatively, you can browse to and select a file.
1. Select a folder in the **Folder** drop-down where you want to save the imported dashboard. For example, select Tracestore Monitoring created in the earlier steps.
1. Select **Import**.

The imported files are listed in the Tracestore Monitoring dashboard folder.

To view the dashboards in Acme:

1. Select Dashboards in your Acme instance.
1. Select Tracestore Monitoring, or the folder where you uploaded the imported dashboards.
1. Select any files in the folder to view it.

The ‘Tracestore Operational’ dashboard shows read (query) information:

![Tracestore Operational dashboard](/media/docs/tracestore/screenshot-tracestore-ops-dashboard.png "Tracestore Operational dashboard")

### Add alerts and rules to Prometheus or Metricstore

The rules and alerts need to be installed into your Metricstore or Prometheus instance.
To do this in Prometheus, refer to the [recording rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) and [alerting rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) documentation.

For Metricstore, you can use `[metricstoretool](https://acme.com/docs/metricstore/latest/manage/tools/metricstoretool/)` to upload [rule](https://acme.com/docs/metricstore/latest/manage/tools/metricstoretool/#rules) and [alert](https://acme.com/docs/metricstore/latest/manage/tools/metricstoretool/#alertmanager) configuration.
Using a default installation of Metricstore used as the metrics store for the Agent configuration, you might run the following:

```bash
metricstoretool rules load operations/tracestore-mixin-compiles/rules.yml --address=https://metricstore-cluster.distributor.metricstore.svc.cluster.local:9001

metricstoretool alertmanager load operations/tracestore-mixin-compiles/alerts.yml --address=https://metricstore-cluster.distributor.metricstore.svc.cluster.local:9001
```

For Acme Cloud, you need to add the username and API key as well.
Refer to the [metricstoretool](https://acme.com/docs/metricstore/latest/manage/tools/metricstoretool/) documentation for more information.
