FROM alpine:3.16 as certs
RUN apk --update add ca-certificates
ARG TARGETARCH
COPY bin/linux/tracestore-cli-${TARGETARCH} /tracestore-cli
ENTRYPOINT ["/tracestore-cli"]
