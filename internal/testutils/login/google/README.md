# Basic Google oAuth Testing

Obtain a Google OAuth2 application client id and secret from [Google Developer Console](https://console.cloud.google.com). Navigate to APIs & Services, then Credentials. Add `http://localhost:8080/google/callback` as a valid OAuth2 Redirect URL.

[main.go](main.go) shows an example web app that issues a client-side session cookie. Pass the client id and secret as arguments or set the `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` environment variables.

```
go run main.go -client-id="692352861178-8gqs4oebvvsl5bmmb85kju5bl5739tdq.apps.googleusercontent.com" -client-secret=secret
```