# SSO and Webfinger Test

Run a simple file server (e.g. `docker compose up sso-ui`) and visit [http://localhost:3001](http://localhost:3001) to test the `.well-known/webfinger` endpoint.

The sample UI expects the API server to be listening on port `17608`. Ensure
`CORE_AUTH_PROVIDERS_REDIRECTURL` is set to `http://localhost:17608` (without the callback path).

To test the full SSO flow with Dex acting as the IdP, run:

```sh
docker compose up dex sso-ui
```

Dex will start with a demo user (`user@example.com` / `password`) and a client
configured to redirect back to the API server. Dex is configured with an issuer
of `http://localhost:5556/dex` so the discovery URL and client settings work
when the API server runs locally. After logging in, the page will
display the authentication response.

The login page mimics the production flow: enter an email address, it queries
the `/\.well-known/webfinger` endpoint, and if SSO is enforced the browser is
redirected to the configured IdP. After authentication you will be returned to
the page and the email address is shown.

After starting Dex, configure the organization settings using the CLI:

```sh
task -d cli orgsetting:enforce-sso
```

This task applies the static client credentials from `docker/files/dex.yaml` so
the login page redirects through Dex correctly.
