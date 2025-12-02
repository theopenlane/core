variable "cloud_token" {
  type = string
  default = getenv("ATLAS_CLOUD_TOKEN")
}

atlas {
  cloud {
    token = var.cloud_token
  }
}

data "remote_dir" "migrations" {
  name = "core"
}

docker "postgres" "dev" {
  image  = "postgres:17"
  schema = "public"
  baseline = file("baseline/extensions.sql")
}

env "dev" {
  dev = docker.postgres.dev.url
}
