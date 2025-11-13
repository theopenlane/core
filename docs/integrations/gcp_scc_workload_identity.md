## Google Cloud SCC Workload Identity Setup

Follow these steps to wire Security Command Center into Openlane using Google Workload Identity Federation. Commands assume you have `gcloud` installed and `PROJECT_ID`, `ORG_ID`, etc. filled in with your own values.

### 1. Enable Required APIs

```bash
gcloud services enable \
  securitycenter.googleapis.com \
  iamcredentials.googleapis.com \
  iam.googleapis.com
```

### 2. Choose the Host Project

```bash
export PROJECT_ID="my-scc-project"
gcloud config set project "$PROJECT_ID"
```

### 3. Create the Workload Identity Pool and Provider

```bash
export POOL_ID="openlane-pool"
export PROVIDER_ID="openlane-provider"
export WORKLOAD_ISSUER_URL="https://accounts.example.com"   # Your IdP issuer
export ALLOWED_AUDIENCE="openlane-scc"

gcloud iam workload-identity-pools create "$POOL_ID" \
  --location="global" \
  --display-name="Openlane Integrations"

gcloud iam workload-identity-pools providers create-oidc "$PROVIDER_ID" \
  --workload-identity-pool="$POOL_ID" \
  --location="global" \
  --display-name="Openlane SCC Provider" \
  --issuer-uri="$WORKLOAD_ISSUER_URL" \
  --allowed-audiences="$ALLOWED_AUDIENCE" \
  --attribute-mapping="google.subject=assertion.sub,attribute.email=assertion.email"
```

Record the provider resource:

```
projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL_ID/providers/PROVIDER_ID
```

### 4. Create the Service Account

```bash
export SA_NAME="openlane-scc-runner"
gcloud iam service-accounts create "$SA_NAME" \
  --description="Openlane SCC integration" \
  --display-name="Openlane SCC Runner"
```

Grant SCC viewer roles at the organization level (adjust as needed):

```bash
export ORG_ID="123456789012"
gcloud organizations add-iam-policy-binding "$ORG_ID" \
  --member="serviceAccount:${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/securitycenter.findingsViewer"

gcloud organizations add-iam-policy-binding "$ORG_ID" \
  --member="serviceAccount:${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/securitycenter.sourcesViewer"
```

### 5. Allow the Pool to Impersonate the Service Account

```bash
export PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')

gcloud iam service-accounts add-iam-policy-binding \
  ${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com \
  --role="roles/iam.workloadIdentityUser" \
  --member="principal://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/subject/*"
```

### 6. Fill Out the Openlane Integration Form

| UI Field | Value |
| --- | --- |
| `projectId` | `$PROJECT_ID` |
| `organizationId` | `$ORG_ID` |
| `workloadIdentityProvider` | `projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL_ID/providers/$PROVIDER_ID` |
| `audience` | `//iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL_ID/providers/$PROVIDER_ID` |
| `serviceAccountEmail` | `${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com` |
| `sourceId` | e.g. `organizations/$ORG_ID/sources/1234567890` (`gcloud scc sources list --organization=$ORG_ID`) |
| `findingFilter` | Optional CEL filter (e.g. `severity="HIGH"`) |
| `workloadPoolProject` | `$PROJECT_ID` (only if the pool lives in another project) |
| `tokenLifetime` | Optional duration such as `3600s` |
| `subjectToken` | OIDC/SAML assertion from your IdP (see below) |

### 7. Obtaining a Subject Token

- **Quick test:** run your IdP flow manually, grab the JWT/SAML assertion, and paste it into the `subjectToken` field.
- **Runtime injection:** leave `subjectToken` blank and provide it per mint via request attributes (UI/CLI support coming soon).
- **Using `gcloud` helper:** generate a credential config to guide token exchange:
  ```bash
  gcloud iam workload-identity-pools create-cred-config \
    "projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/providers/${PROVIDER_ID}" \
    --service-account="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --credential-source-file=cred.json
  ```
  Follow the instructions in `cred.json` to exchange your IdP token when needed.

### 8. Verify via the Integrations UI

1. Start the dev stack: `task run-dev`
2. Open `http://localhost:17608/pkg/testutils/integrations/index.html`
3. Submit the SCC configuration
4. Under the SCC card, run:
   - `health.default` to ensure SCC is reachable
   - `findings.collect` with your source/filter
5. Check the activity log for summaries like `Collected N findings…` which confirms the broker minted a token, impersonated the service account, and SCC accepted the call.

### Quick Test Using a Service-Account Key (Optional)

If you just need to validate the SCC integration before wiring workload identity, you can paste a service-account JSON key into the **Service Account Key JSON** field on the integration form. The provider will mint tokens directly from that key. This is handy for demos, but:

- Treat the key like any other secret (store it in hush/secret manager, rotate, delete afterward).
- Remove the key once workload identity is in place to avoid long-lived credentials.
