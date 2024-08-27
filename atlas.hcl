variable "cloud_token" {
  type    = string
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


variable "token" {
  type    = string
  default = getenv("TURSO_TOKEN")
}
