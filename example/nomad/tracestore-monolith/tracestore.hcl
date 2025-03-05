variable "version" {
  type        = string
  description = "Tracestore version"
  default     = "2.7.1"
}

variable "prometheus_remote_write_url" {
  type        = string
  description = "Prometheus Remote Write URL"
  default     = "http://prometheus.service.consul/api/v1/write"
}

variable "s3_url" {
  type        = string
  description = "S3 URL"
  default     = "s3.dummy.url"
}

variable "s3_access_key_id" {
  type        = string
  description = "S3 Access Key ID"
  default     = "any"
}

variable "s3_secret_access_key" {
  type        = string
  description = "S3 Secret Access Key"
  default     = "any"
}

job "tracestore" {
  datacenters = ["*"]

  group "tracestore" {
    count = 1

    network {
      port "http" { to = 3200 }
      port "grpc" {}
      port "otlp" { to = 4317 }
    }

    service {
      name = "tracestore-http"
      port = "http"
      tags = []
      check {
        name     = "tracestore-http"
        port     = "http"
        type     = "http"
        path     = "/ready"
        interval = "20s"
        timeout  = "1s"
      }
    }

    service {
      name = "tracestore-grpc"
      port = "grpc"
      tags = []
      check {
        port     = "grpc"
        type     = "grpc"
        interval = "20s"
        timeout  = "1s"
        grpc_use_tls = false
        tls_skip_verify = true
      }
    }

    service {
      name = "tracestore-otpl"
      port = "otpl"
      tags = []
    }

    task "tracestore" {
      driver       = "docker"
      user         = "nobody"
      kill_timeout = "90s"

      config {
        image = "acme/tracestore:${var.version}"
        ports = [
          "http",
          "grpc",
        ]

        args = [
          "-target=all",
          "-config.file=/local/config.yml",
          "-config.expand-env=true",
        ]
      }
      template {
        data        = <<-EOC
        server:
          log_level: info
          http_listen_port: {{ env "NOMAD_PORT_http" }}
          grpc_listen_port: {{ env "NOMAD_PORT_grpc" }}

        distributor:
          receivers:                           # this configuration will listen on all ports and protocols that tracestore is capable of.
            otlp:
              protocols:
                http:
                grpc: 0.0.0.0:{{ env "NOMAD_PORT_otlp" }}

        metrics_generator:
          processor:
            service_graphs:
              max_items: 10000

          storage:
            path: {{ env "NOMAD_ALLOC_DIR" }}/tracestore/wal
            remote_write:
              - url: ${var.prometheus_remote_write_url}
                send_exemplars: true

        storage:
          trace:
            backend: s3
            wal:
              path: {{ env "NOMAD_ALLOC_DIR" }}/tracestore/wal
            local:
              path: {{ env "NOMAD_ALLOC_DIR" }}/tracestore/blocks
            s3:
              bucket: tracestore                    # how to store data in s3
              endpoint: ${var.s3_url}
              insecure: true
              access_key: ${var.s3_access_key_id}
              secret_key: ${var.s3_secret_access_key}

        overrides:
          defaults:
            metrics_generator:
              processors:
                - service-graphs
                - span-metrics
        EOC
        destination = "local/config.yml"
      }


      resources {
        cpu        = 300
        memory     = 1024
      }
    }
  }
}