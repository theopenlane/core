docker "postgres" "dev" {
  image  = "postgres:18"
  schema = "public"
  baseline = file("baseline/extensions.sql")
}

env "dev" {
  dev = docker.postgres.dev.url
}
