# Keystore Overview

The `keystore` package is the installation-scoped credential and client lifecycle boundary for integrations.

Today it owns:
- credential persistence for a single installation via hush secrets
- replacement/deletion of installation credentials
- pooled client initialization keyed by installation, client name, credential content, and config payload
- invalidation of pooled clients when credentials change

It does not own:
- definition registration
- auth flow orchestration
- operation dispatch or execution
- ingest

Those concerns remain in `internal/integrations/runtime` and `internal/keymaker`.
