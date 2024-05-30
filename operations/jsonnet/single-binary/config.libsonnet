{
  _images+:: {
    tracestore: 'acme/tracestore:latest',
    tracestore_query: 'acme/tracestore-query:latest',
    tracestore_vulture: 'acme/tracestore-vulture:latest',
  },

  _config+:: {
    tracestore: {
      port: 3200,
      replicas: 1,
      headless_service_name: 'tracestore-members',
    },
    // disable tracestore-query by default
    tracestore_query: {
      enabled: false,
    },
    pvc_size: error 'Must specify a pvc size',
    pvc_storage_class: error 'Must specify a pvc storage class',
    receivers: error 'Must specify receivers',
    ballast_size_mbs: '1024',
    jaeger_ui: {
      base_path: '/',
    },
  },
}
