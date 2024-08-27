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