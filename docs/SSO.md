# Configuring SSO Providers

Openlane allows you to use external identity providers (IdPs) so that all authentication for your organization flows through a provider you control. The following guide explains how to configure Okta and Google Workspace as SSO providers and how to update your organization settings in Openlane.

## Prerequisites

- Admin access to your identity provider (Okta or Google Workspace)
- An organization owner role in Openlane

## 1. Configure Okta

1. Log in to Okta as an administrator and navigate to **Applications → Create App Integration**.
2. Choose **OIDC - Web Application** and select **Next**.
3. Provide a name (for example, `Openlane`) and set the following redirect URIs:
   - `https://<your-openlane-host>/oidc/callback` (replace with your domain)
4. After the application is created, copy the **Client ID** and **Client Secret**.
5. From the application's **General** tab, locate the **Okta domain** and note the **Issuer URI** (for example `https://dev-123456.okta.com/oauth2/default`). This is your OIDC discovery endpoint.

### Update Organization Settings

In Openlane, update your organization settings with the values obtained from Okta:

```graphql
mutation UpdateOrg {
  updateOrganizationSetting(
    id: "<orgSettingID>"
    input: {
      identityProvider: OKTA
      identityProviderClientID: "<client-id>"
      identityProviderClientSecret: "<client-secret>"
      oidcDiscoveryEndpoint: "<issuer-uri>"
      identityProviderLoginEnforced: true
    }
  ) {
    organizationSetting { id }
  }
}
```

Setting `identityProviderLoginEnforced` to `true` requires all organization members to authenticate through Okta.

## 2. Configure Google Workspace

1. In the [Google Cloud Console](https://console.cloud.google.com/), create a new project or select an existing one.
2. Navigate to **APIs & Services → Credentials** and choose **Create OAuth client ID**.
3. Select **Web application** as the application type and add `https://<your-openlane-host>/oidc/callback` to the list of authorized redirect URIs.
4. Download the newly created credentials or copy the generated **Client ID** and **Client Secret**.
5. Google publishes the discovery URL at `https://accounts.google.com/.well-known/openid-configuration`.

### Update Organization Settings

Use the values from Google to update your organization settings:

```graphql
mutation UpdateOrg {
  updateOrganizationSetting(
    id: "<orgSettingID>"
    input: {
      identityProvider: GOOGLEWORKSPACE
      identityProviderClientID: "<client-id>"
      identityProviderClientSecret: "<client-secret>"
      oidcDiscoveryEndpoint: "https://accounts.google.com"
      identityProviderLoginEnforced: true
    }
  ) {
    organizationSetting { id }
  }
}
```

When the update is saved, users will be redirected to Google Workspace for sign-in.

## 3. Verifying the Configuration

After updating the organization settings, log out of Openlane and attempt to log in again. You should be redirected to the configured identity provider. If the login process succeeds, your Openlane session will start normally.

If you need to disable SSO enforcement temporarily, update `identityProviderLoginEnforced` to `false` in your organization settings.

---

For other providers supported by Openlane (such as OneLogin or GitHub), use a similar process to create an OAuth/OIDC application and update the organization settings accordingly.
