# Integrations TO DO notes

- make sure we have a /disconnect endpoint
- update the keystore secret names / integration references hardened types (and not string concat sillynesss)
- see about collapsing existing social provider login to new devlarative setup
- check to see if there are known good sources for the providers client libraries (existing go projects with lots of plugins, etc).
- look into making credential store a full oauth2 service
- explore possibility of collapsing object store providers into this setup?
- determine appropriate method of versioning a provider integration (possibly based on the external provider's client version?)
- add integration health checks to ensure we have all of the appropriate permissions, scopes, credentials, whatever (kind of like how when you're adding SAML to github, you have to go through a test flow, make sure that our test flow confirms the integration health)
