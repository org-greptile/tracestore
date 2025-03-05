variable "version" {
  type        = string
  description = "Tracestore version"
  default     = "2.7.1"
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

  group "metrics-generator" {
    count = 1

    network {
      port "http" {}
      port "grpc" {}
      port "otlp" { to = 4317 }
    }

    service {
      name = "tracestore-metrics-generator"
      port = "http"
      tags = []
      check {
        name     = "metrics-generator"
        port     = "http"
        type     = "http"
        path     = "/ready"
        interval = "20s"
        timeout  = "1s"
      }
    }

    service {
      name = "tracestore-metrics-generator-grpc"
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

    task "metrics-generator" {
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
          "-target=metrics-generator",
          "-config.file=/local/config.yml",
          "-config.expand-env=true",
        ]
      }

      template {
        data        = file("config.yml")
        destination = "local/config.yml"
      }

      template {
        data = <<-EOH
        S3_ACCESS_KEY_ID=${var.s3_access_key_id}
        S3_SECRET_ACCESS_KEY=${var.s3_secret_access_key}
        EOH

        destination = "secrets/s3.env"
        env         = true
      }

      resources {
        cpu        = 200
        memory     = 128
        memory_max = 1024
      }
    }
  }

  group "query-frontend" {
    count = 1

    network {
      port "http" {}
      port "grpc" { static = 9095}
      port "otlp" { to = 4317 }
    }

    service {
      name = "tracestore-query-frontend"
      port = "http"
      tags = []
    }

    service {
      name = "tracestore-query-frontend-grpc"
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

    task "query-frontend" {
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
          "-target=query-frontend",
          "-config.file=/local/config.yml",
          "-config.expand-env=true",
        ]
      }

      template {
        data        = file("config.yml")
        destination = "local/config.yml"
      }

      template {
        data = <<-EOH
        S3_ACCESS_KEY_ID=${var.s3_access_key_id}
        S3_SECRET_ACCESS_KEY=${var.s3_secret_access_key}
        EOH

        destination = "secrets/s3.env"
        env         = true
      }

      resources {
        cpu        = 200
        memory     = 128
        memory_max = 1024
      }
    }
  }

  group "ingester" {
    count = 3

    network {
      port "http" {}
      port "grpc" {}
      port "otlp" { to = 4317 }
    }

    service {
      name = "tracestore-ingester"
      port = "http"
      tags = []
      check {
        name     = "Tracestore ingester"
        port     = "http"
        type     = "http"
        path     = "/ready"
        interval = "20s"
        timeout  = "1s"
      }
    }

    service {
      name = "tracestore-ingester-grpc"
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

    task "ingester" {
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
          "-target=ingester",
          "-config.file=/local/config.yml",
          "-config.expand-env=true",
        ]
        network_mode = "host"
      }

      template {
        data        = file("config.yml")
        destination = "local/config.yml"
      }

      template {
        data = <<-EOH
        S3_ACCESS_KEY_ID=${var.s3_access_key_id}
        S3_SECRET_ACCESS_KEY=${var.s3_secret_access_key}
        EOH

        destination = "secrets/s3.env"
        env         = true
      }

      resources {
        cpu        = 300
        memory     = 128
        memory_max = 2048
      }
    }
  }

  group "compactor" {
    count = 1

    ephemeral_disk {
      size    = 1000
      sticky  = true
    }

    network {
      port "http" {}
      port "grpc" {}
      port "otlp" { to = 4317 }
    }

    service {
      name = "tracestore-compactor"
      port = "http"
      tags = []
      check {
        name     = "Tracestore compactor"
        port     = "http"
        type     = "http"
        path     = "/ready"
        interval = "20s"
        timeout  = "1s"
      }
    }

    service {
      name = "tracestore-compactor-grpc"
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

    task "compactor" {
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
          "-target=compactor",
          "-config.file=/local/config.yml",
          "-config.expand-env=true",
        ]
      }

      template {
        data        = file("config.yml")
        destination = "local/config.yml"
      }

      template {
        data = <<-EOH
        S3_ACCESS_KEY_ID=${var.s3_access_key_id}
        S3_SECRET_ACCESS_KEY=${var.s3_secret_access_key}
        EOH

        destination = "secrets/s3.env"
        env         = true
      }

      resources {
        cpu        = 3000
        memory     = 256
        memory_max = 1024
      }
    }
  }
  group "distributor" {
    count = 1

    network {
      port "http" {}
      port "grpc" {}
      port "otpl" { to = 4317 }
    }

    service {
      name = "tracestore-distributor"
      port = "http"
      tags = []
      check {
        name     = "Tracestore distributor"
        port     = "http"
        type     = "http"
        path     = "/ready"
        interval = "20s"
        timeout  = "1s"
      }
    }

    service {
      name = "tracestore-distributor-otpl"
      port = "otpl"
      tags = []
    }

    service {
      name = "tracestore-distributor-grpc"
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

    task "distributor" {
      driver       = "docker"
      user         = "nobody"
      kill_timeout = "90s"

      config {
        image = "acme/tracestore:${var.version}"
        ports = [
          "http",
          "grpc",
          "otpl",
        ]

        args = [
          "-target=distributor",
          "-config.file=/local/config.yml",
          "-config.expand-env=true",
        ]
      }

      template {
        data        = file("config.yml")
        destination = "local/config.yml"
      }

      template {
        data = <<-EOH
        S3_ACCESS_KEY_ID=${var.s3_access_key_id}
        S3_SECRET_ACCESS_KEY=${var.s3_secret_access_key}
        EOH

        destination = "secrets/s3.env"
        env         = true
      }

      resources {
        cpu        = 200
        memory     = 128
        memory_max = 1024
      }
    }
  }
  group "querier" {
    count = 1

    network {
      port "http" {}
      port "grpc" {}
      port "otlp" { to = 4317 }
    }

    service {
      name = "tracestore-querier"
      port = "http"
      tags = []
      check {
        name     = "Tracestore querier"
        port     = "http"
        type     = "http"
        path     = "/ready"
        interval = "50s"
        timeout  = "1s"
      }
    }

    service {
      name = "tracestore-querier-grpc"
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

    task "querier" {
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
          "-target=querier",
          "-config.file=/local/config.yml",
          "-config.expand-env=true",
        ]
      }

      template {
        data        = file("config.yml")
        destination = "local/config.yml"
      }

      template {
        data = <<-EOH
        S3_ACCESS_KEY_ID=${var.s3_access_key_id}
        S3_SECRET_ACCESS_KEY=${var.s3_secret_access_key}
        EOH

        destination = "secrets/s3.env"
        env         = true
      }

      resources {
        cpu        = 200
        memory     = 128
        memory_max = 2048
      }
    }
  }
}
