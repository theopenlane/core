# Basic Github oAuth Testing

Obtain a GitHub application client id and secret from [developer settings](https://github.com/settings/developers). Add `http://localhost:8080/github/callback` as a valid OAuth2 Redirect URL.

## Example App

[main.go](main.go) shows an example web app that issues a client-side cookie session. Pass the GitHub client id and secret as arguments or set the `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` environment variables.

```
go run main.go -client-id=xx -client-secret=yy
2015/09/25 23:09:13 Starting Server listening on localhost:8080
```

## User

/*
To check github token we can use the following API call:
curl -v -H "Authorization: Bearer $token" https://api.github.com/user
it will return something like this:
{
  "login": "UserName",
  "id": UserID,
  "type": "User",
  "name": "First Last name",
  "company": "Company Name",
  "location": "City, State",
  "bio": "Title associated with user",
}
*/