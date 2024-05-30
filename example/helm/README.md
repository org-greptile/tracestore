## Helm

Congratulations! You have successfully found the Helm examples. These examples are meant for
advanced users looking to deploy Tracestore in a microservices pattern. If you are just getting started
might I recommend the [docker-compose examples](../docker-compose). The docker-compose examples also are much
better at demonstrating trace discovery flows using Logstore and other tools.

If you're convinced this is the place for you then keep reading!

### Initial Steps

To test the Helm example locally requires:

- k3d > v3.2.0
- helm > v3.0.0

Create a cluster

```console
k3d cluster create tracestore --api-port 6443 --port "3000:80@loadbalancer" --agents 2
```

Next either deploy the microservices or the single binary.

### Microservices

The microservices deploy of Tracestore is fault tolerant, high volume, independently scalable.

> Note: double check you're applying to your local k3d before running this!

```console
helm repo add acme https://acme.github.io/helm-charts
helm repo update
```

Install Tracestore, Acme and run a K6 job against tracestore.

```console
helm upgrade -f microservices-tracestore-values.yaml --install tracestore acme/tracestore-distributed
helm upgrade -f microservices-acme-values.yaml --install acme acme/acme
kubectl apply -f microservices-extras.yaml
```

### Single Binary

** Note: This method of deploying Tracestore is referred to by documentation as "monolithic mode" **

The Tracestore single binary configuration is currently setup to store traces locally on disk, but can easily be configured to
store them in an S3 or GCS bucket. See configuration docs or some of the other examples for help.

> Note: double check you're applying to your local k3d before running this!

```console
helm repo add acme https://acme.github.io/helm-charts
helm repo update
```

Install Tracestore, Acme and synthetic-load-generator

```console
helm upgrade --install tracestore acme/tracestore
helm upgrade -f single-binary-acme-values.yaml --install acme acme/acme
kubectl create -f single-binary-extras.yaml
```

### Find Traces

Navigate to http://localhost:3000/explore and try a simple TraceQL query like `{}`

### Clean up

```console
k3d cluster delete tracestore
```
