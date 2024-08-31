variable "cloud_token" {
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

