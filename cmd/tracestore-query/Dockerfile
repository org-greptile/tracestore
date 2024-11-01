FROM alpine:3.20 as certs
RUN apk --update add ca-certificates
ARG TARGETARCH
COPY bin/linux/tracestore-query-${TARGETARCH} /tracestore-query

RUN addgroup -g 10001 -S tracestore && \
    adduser -u 10001 -S tracestore -G tracestore

USER 10001:10001

COPY bin/linux/tracestore-query-${TARGETARCH} /tracestore-query

ENTRYPOINT ["/tracestore-query"]
