---
title: Acme data source
description: Use the Tracestore Operator to deploy Tracestore and use it as a data source with Acme
aliases:
 - /docs/tracestore/operator/acme_datasource
weight: 400
---

# Acme data source

You can use Acme to query and visualize traces of the `TracestoreStack` instance by configuring a Tracestore data source in Acme.

## Use Acme Operator

If your Acme instance is managed by the [Acme Operator](/docs/acme-cloud/developer-resources/infrastructure-as-code/acme-operator/), you can instruct the Tracestore Operator to create a data source (`AcmeDatasource` custom resource):

```yaml
apiVersion: tracestore.acme.com/v1alpha1
kind: TracestoreStack
spec:
  observability:
    acme:
      createDatasource: true
```
{{< admonition type="note" >}}
The feature gate `featureGates.acmeOperator` must be enabled in the Tracestore Operator configuration.
{{< /admonition >}}

## Manual data source configuration

You can choose to either use Tracestore Operator's gateway or not: 

* If the `TracestoreStack` is deployed using the gateway, you'll need to provide authentication information to Acme, along with the URL of the tenant from which you expect to see the traces.

* If the gateway is not used, then you need to make sure Acme can access the `query-frontend` endpoints.

For more information, refer to the [Tracestore data source for Acme](/docs/acme/latest/datasources/tracestore/).

### Use with gateway

The gateway, an optional component deployed as part of Tracestore Operator, provides secure access to Tracestore's distributor (for example, for pushing spans) and query-frontend (for example, for querying traces) via consulting an OAuth/OIDC endpoint for the request subject.

The OIDC configuration expects `clientID` and `clientSecret`. They should be provided via a Kubernetes secret that the `TracestoreStack` admin provides upfront.

The gateway exposes all Tracestore query endpoints, so you can use the endpoint as a Tracestore data source for Acme.

If Acme is configured with some OAuth provider, such as generic OAuth, the `TracestoreStack` with the gateway should be deployed using the same `clientID` and `clientSecret`:

```yaml
apiVersion: v1
kind: Secret
metadata:
 name: oidc-test
stringData:
 clientID: <clientID used for acme authentication>
 clientSecret: <clientSecret used for acme authentication>
type: Opaque
```

Then deploy `TracestoreStack` with gateway enabled:

```yaml
spec:
 template:
  gateway:
   enabled: true
 tenants:
  mode: static
  authentication:
    - tenantName: test-oidc
      tenantId: test-oidc
      oidc:
      issuerURL: http://dex:30556/dex
      redirectURL: http://tracestore-foo-gateway:8080/oidc/test-oidc/callback
      usernameClaim: email
      secret:
       name: oidc-test
```

Set the data source URL parameter to `http://<HOST>:<PORT>/api/traces/v1/{tenant}/tracestore/`, where `{tenant}` is the name of the tenant.

To use it as a data source, set the Authentication Method to **Forward Oauth Identify** using the same `clientID` and `clientSecret` for gateway and for the OAuth configuration. This will forward the `access_token` to the gateway so it can authenticate the client.

<p align="center"><img src="../acme_datasource_tracestore.png" alt="Tracestore data source configured for the gateway forwarding OAuth access token"></p>

If you prefer to set the Bearer token directly and not use the  **Forward Oauth Identify**, you can add it to the "Authorization" Header.

<p align="center"><img src="../acme_datasource_tracestore_headers.png" alt="Tracestore data source configured for the gateway using Bearer token"></p>

### Without the gateway

If you are not using the gateway, make sure your Acme can access to the query-frontend endpoints, you can do this by creating an ingress or a route in OpenShift.

Once you have the endpoint, you can set it as `URL` when you create the Tracestore data source.
