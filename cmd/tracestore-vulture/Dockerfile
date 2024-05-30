FROM alpine:3.20 as certs
RUN apk --update add ca-certificates
ARG TARGETARCH
COPY bin/linux/tracestore-vulture-${TARGETARCH} /tracestore-vulture
ENTRYPOINT ["/tracestore-vulture"]
