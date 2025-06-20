# Configuration

You will need to perform a 1-time action of creating a `.config.yaml` file based on the `.example` files.
The Taskfiles will also source a `.dotenv` files which match the naming conventions called for `{{.ENV}}` to ease the overriding of environment variables. These files are intentionally added to the `.gitignore` within this repository to prevent you from accidentally committing secrets or other sensitive information which may live inside the server's environment variables.

All settings in the `yaml` configuration can also be overwritten with environment variables prefixed with `CORE_`. For example, to override the Google `client_secret` set in the yaml configuration with an environment variable you can use:

```
export CORE_AUTH_PROVIDERS_GOOGLE_CLIENTSECRET
```

Configuration precedence is as follows, the latter overriding the former:

1. `default` values set in the config struct within the code
1. `.config.yaml` values
1. Environment variables

## Regenerating

If you've made changes to the code in this code base (specifically interfaces referenced in the `config.go`) and want to regenerate the configuration, run `task config:generate`

## Token key configuration

JWT signing keys are provided via the `token.keys` map. Each entry maps a ULID
(`kid`) to the path of a PEM encoded RSA private key. Keys can also be supplied
through the environment variable `CORE_AUTH_TOKEN_KEYS` using a comma separated
list in the form `kid=/path/key.pem`. When running inside Kubernetes you can
mount a secret containing one or more PEM files and set
`CORE_AUTH_TOKEN_KEYDIR` to the mount path to automatically load all keys from
that directory. When `CORE_AUTH_TOKEN_KEYDIR` is set the server also watches the
directory for changes and reloads the key set without needing a restart.

To rotate keys, create a new PEM file with a new ULID in the directory or update
the `CORE_AUTH_TOKEN_KEYS` variable with the additional entry. Keep the previous
key until all issued tokens expire.

## Module Catalog

`moduleCatalogFile` defines the YAML file containing module and add-on definitions. Use environment variable `CORE_MODULECATALOGFILE` to override the path. Each feature lists one or more price options in `billing.prices`. The server searches Stripe for a matching price based on interval, amount, nickname, and metadata. If none is found a new product and price are created automatically during startup and catalog reloads. If you want to "dry run" confirm what resources would or would not be created via this process, you can run `go run ./cmd/catalog` and verify the output
