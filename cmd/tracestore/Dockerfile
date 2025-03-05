FROM alpine:latest AS ca-certificates
RUN apk add --update --no-cache ca-certificates

FROM gcr.io/distroless/static-debian12:debug

# we need this because some docker-compose files call chown assuming there's a shell
SHELL ["/busybox/sh", "-c"]

RUN ["/busybox/addgroup", "-g", "10001", "-S", "tracestore"]
RUN ["/busybox/adduser", "-u", "10001", "-S", "tracestore", "-G", "tracestore"]
RUN ["/busybox/mkdir", "-p", "/var/tracestore", "-m", "0700"]
RUN ["/busybox/chown", "-R", "tracestore:tracestore", "/var/tracestore"]

ARG TARGETARCH
COPY bin/linux/tracestore-${TARGETARCH} /tracestore
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 10001:10001

ENTRYPOINT ["/tracestore"]
