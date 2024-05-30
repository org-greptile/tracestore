FROM alpine:3.20 as certs
RUN apk --update add ca-certificates
ARG TARGETARCH
COPY bin/linux/tracestore-${TARGETARCH} /tracestore

RUN addgroup -g 10001 -S tracestore && \
    adduser -u 10001 -S tracestore -G tracestore

RUN mkdir -p /var/tracestore && \
    chown -R tracestore:tracestore /var/tracestore

USER 10001:10001

ENTRYPOINT ["/tracestore"]
