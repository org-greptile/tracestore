FROM jaegertracing/jaeger-query:1.43.0

ENV SPAN_STORAGE_TYPE=grpc-plugin \
    GRPC_STORAGE_PLUGIN_BINARY=/tmp/tracestore-query

# This is silly, but it's important that tracestore-query gets copied into /tmp
#  b/c it forces a /tmp dir to exist which hashicorp plugins depend on
ARG TARGETARCH
COPY bin/linux/tracestore-query-${TARGETARCH} /tmp/tracestore-query
