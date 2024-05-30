[
  {
	// import the microservices example
    local tracestore = import '../tracestore-microservices/main.jsonnet',

    name: 'cluster name',
    apiServer: 'https://0.0.0.0:6443',
    namespace: 'namespace',
    
    data: tracestore,

    dataOverride: {
      _images+:: {
        // images can be overridden here if desired
      },

      _config+:: {

        // config can be overridden here if desired
        
      },

    },

  },
]
