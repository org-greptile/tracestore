# tracestore-mixin

Once installed via the build instructions below, dashboards, rules, and alerts are located in the [`operations/tracestore-mixins-compiled`](..tracestore-mixins-compiled) folder. Use them directly in Prometheus and Acme to monitor Tracestore.

You can either use the mixins in the `tracestore-mixin-compiled` folder or you can build your own to incorporate your own changes. 

## Build

To regenerate dashboards, rule and alerts, run `make all`.

This requires [jsonnet](https://jsonnet.org/) and [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler) to be installed. 

On macOS, you can install these with the following commands:

```console
brew install jsonnet
brew install jsonnet-bundler 
go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@v0.4.0
```

## Use the mixins

Once you run `make all`, the mixins are created in the `tracestore-mixin-compiled` folder. 

```
➜   make all                                                                                                                          git:(main|)
jb install
jsonnet -J vendor -S dashboards.jsonnet -m ../tracestore-mixin-compiled/dashboards/
../tracestore-mixin-compiled/dashboards/tracestore-operational.json
../tracestore-mixin-compiled/dashboards/tracestore-reads.json
../tracestore-mixin-compiled/dashboards/tracestore-resources.json
../tracestore-mixin-compiled/dashboards/tracestore-rollout-progress.json
../tracestore-mixin-compiled/dashboards/tracestore-tenants.json
../tracestore-mixin-compiled/dashboards/tracestore-writes.json
jsonnet -J vendor -S alerts.jsonnet > ../tracestore-mixin-compiled/alerts.yaml
jsonnet -J vendor -S rules.jsonnet > ../tracestore-mixin-compiled/rules.yaml
```

Alerts and rules are listed in their matching files: 
* Alerts -> `tracestore-mixin-compiled/alerts.yaml`
* Rules -> 'tracestore-mixin-compiled/rules.yaml`

For information on using the dashboards, refer to the [mixin runbook](operations/tracestore-mixin/runbook.md).
