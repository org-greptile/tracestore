FROM jaegertracing/jaeger-query:1.50.0

ENV SPAN_STORAGE_TYPE=grpc-plugin \
    GRPC_STORAGE_PLUGIN_BINARY=/tracestore-query

# Ensure /tmp dir exists, hashicorp plugins need a /tmp dir to exist.
RUN mkdir -p /tmp

ARG TARGETARCH
COPY bin/linux/tracestore-query-${TARGETARCH} /tracestore-query
