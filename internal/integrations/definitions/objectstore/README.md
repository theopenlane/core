# Object storage integration

Definition id `def_01K0OBJECTSTORE00000000001`. Two modes share one client type (the
`storage.Provider` interface from `pkg/objects/storage`, backed by the GCS, S3 or R2 provider) and the
same operations and mappings. The credential a customer configures is the provider selection: each
credential slot maps to exactly one provider.

| Mode | Who provisions it | Credentials | Owner of ingested rows |
|---|---|---|---|
| Customer installation | An organization, through the integrations API or console | One of: GCS workload identity federation (recommended), GCS service account key, AWS IAM role, AWS access keys, Cloudflare R2 keys; stored in the keystore | The installing organization |
| Runtime integration | The operator, through `integrations.objectstoreruntime` config with `provider` set to `gcs`, `s3` or `r2` | GCS: application default credentials on the host (workload identity on GKE). S3: static keys, or the default credential chain (IRSA / instance role) when the keys are empty. R2: account id and keys | None: rows are `system_owned = true` with no `owner_id` |

Operations (names are the Go config type names):

| Operation | Path | What it does |
|---|---|---|
| `HealthCheck` | inline | Lists one object under the bucket root to prove the credentials reach the bucket |
| `ImportRecords` | reconcile (customer only) | Lists `*.json` objects under a prefix, maps each record through the definition's CEL mapping, upserts into the internal model |
| `WriteObject` | inline | Writes a document to the bucket under a key |
| `SystemImport` | scheduled (runtime only, internal) | `ImportRecords` for the operator bucket, persisting system-owned rows |

## Record files

Every object whose key ends in `.json` under the prefix is read. A file is either a JSON array of
records or one JSON object. Record keys are the target schema's snake_case input keys; see
`examples/entities.json` for Entity. `external_id` is the upsert key and must be present.

Only `Entity` is accepted as `schema` today. `variant: "vendor"` additionally links each entity to the
entity type named `vendor` in the same owner scope (the organization's for a customer import, the
system-owned one for the runtime import).

## Customer installation

The installation identity is `<scheme>` + bucket (`gs://`, `s3://` or `r2://`), so the same bucket name
on two providers is two distinct installations.

### Google Cloud Storage

#### Workload identity federation

The platform mints an RS256 assertion signed with its token keys and exchanges it at Google STS. The
assertion carries:

| Claim | Value |
|---|---|
| `iss` | the platform issuer (`auth.token.issuer` in config). Shown to the customer as **Openlane Issuer URI** on the connection |
| `sub` | the installing organization id |
| `organization_id` | the installing organization id |
| `aud` | `//iam.googleapis.com/projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/openlane/providers/openlane` |

The pool and provider ids are fixed to `openlane`; the customer supplies only the project number that
hosts them. Google fetches the platform's keys from `<issuer>/.well-known/openid-configuration` and
`<issuer>/.well-known/jwks.json`, so the issuer must be reachable from Google over public HTTPS.

`examples/setup-gcp-federation.sh` is GCS-only: it performs the whole customer-side setup against a
test project (APIs, bucket, pool, provider, service account or direct binding, example records) and
prints the configure request to submit. The equivalent steps by hand, with `ORG_ID` the Openlane
organization id and `ISSUER` the issuer URI:

```bash
gcloud iam workload-identity-pools create openlane \
  --project="$PROJECT_ID" --location=global

gcloud iam workload-identity-pools providers create-oidc openlane \
  --project="$PROJECT_ID" --location=global --workload-identity-pool=openlane \
  --issuer-uri="$ISSUER" \
  --allowed-audiences="//iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/openlane/providers/openlane" \
  --attribute-mapping="google.subject=assertion.sub,attribute.organization_id=assertion.organization_id" \
  --attribute-condition="assertion.sub=='$ORG_ID'"
```

Then grant bucket access to the federated principal directly:

```bash
gcloud storage buckets add-iam-policy-binding "gs://$BUCKET" \
  --member="principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/openlane/subject/$ORG_ID" \
  --role=roles/storage.objectUser
```

or, to act as a service account instead, grant the service account the bucket role and let the
principal impersonate it, then set `serviceAccountEmail` on the credential:

```bash
gcloud storage buckets add-iam-policy-binding "gs://$BUCKET" \
  --member="serviceAccount:$SA_EMAIL" --role=roles/storage.objectUser

gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --member="principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/openlane/subject/$ORG_ID" \
  --role=roles/iam.workloadIdentityUser
```

Configure the installation (`POST /v1/integrations/def_01K0OBJECTSTORE00000000001/config`):

```json
{
  "credentialRef": "WorkloadIdentityCredentialSchema",
  "body": {
    "projectNumber": "123456789012",
    "serviceAccountEmail": "openlane-import@my-project.iam.gserviceaccount.com",
    "bucket": "my-records",
    "projectId": "my-project"
  },
  "userInput": {
    "import": { "prefix": "vendors/", "schema": "Entity", "variant": "vendor" }
  }
}
```

The connection health check runs during configuration and the installation is marked connected when
it passes. `ImportRecords` is a reconcile operation, so it is dispatched on first connection and on
the recurring reconcile loop using the `userInput.import` spec. Run it on demand with
`POST /v1/integrations/<integrationID>/operations/run`:

```json
{ "operation": "ImportRecords", "config": { "prefix": "vendors/", "schema": "Entity", "variant": "vendor" } }
```

#### Service account key

Same flow with `"credentialRef": "ServiceAccountCredentialSchema"` and a body of
`{ "serviceAccountKey": "<key json>", "bucket": "...", "projectId": "..." }`. The key is stored in the
keystore; the service account needs `roles/storage.objectUser` (or read-only `objectViewer` when only
importing) on the bucket.

### AWS (IAM role or access keys)

Both slots build an S3 client for one bucket in one region. The IAM policy attached to the role or
user needs `s3:ListBucket` on the bucket and `s3:GetObject` and `s3:PutObject` on its objects:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    { "Effect": "Allow", "Action": "s3:ListBucket", "Resource": "arn:aws:s3:::my-records" },
    { "Effect": "Allow", "Action": ["s3:GetObject", "s3:PutObject"], "Resource": "arn:aws:s3:::my-records/*" }
  ]
}
```

#### IAM role

Openlane assumes a role in the customer account from its own AWS identity (the operator's
`integrations.objectstore` source credentials, or the host's default credential chain when those are
empty). The role's trust policy names the Openlane principal, shown on the connection as **Openlane
Principal ARN**, and requires the external id generated for the installation:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "AWS": "<Openlane Principal ARN>" },
      "Action": "sts:AssumeRole",
      "Condition": { "StringEquals": { "sts:ExternalId": "<externalId>" } }
    }
  ]
}
```

Configure with `"credentialRef": "AWSAssumeRoleCredentialSchema"`:

```json
{
  "credentialRef": "AWSAssumeRoleCredentialSchema",
  "body": {
    "roleArn": "arn:aws:iam::123456789012:role/openlane-storage",
    "externalId": "<generated>",
    "region": "us-east-1",
    "bucket": "my-records"
  },
  "userInput": {
    "import": { "prefix": "vendors/", "schema": "Entity", "variant": "vendor" }
  }
}
```

`sessionName` optionally overrides the STS session name.

#### Access keys

Same flow with `"credentialRef": "AWSAccessKeyCredentialSchema"` and a body of
`{ "accessKeyId": "...", "secretAccessKey": "...", "region": "us-east-1", "bucket": "my-records" }`.
The keys belong to an IAM user carrying the bucket policy above and are stored in the keystore.

### Cloudflare R2

Create an R2 API token with object read and write on the bucket; the token's S3-compatible access key
id and secret are the credential. Configure with `"credentialRef": "R2CredentialSchema"` and a body of
`{ "accountId": "<cloudflare account id>", "accessKeyId": "...", "secretAccessKey": "...", "bucket": "my-records" }`.
The endpoint is derived from the account id (`https://<accountId>.r2.cloudflarestorage.com`).

## Runtime integration

To run only the runtime integration on a deployment, set `integrations.objectstore.runtimeonly: true`.
The definition then registers with `RuntimeOnly` set: it is hidden from the catalog and the configure
and auth-flow endpoints refuse organization installs, while the system import stays available to triggers.

Operator config, under `integrations` in the server config. `provider` selects the backend and the
remaining fields are read per provider (`projectid` for `gcs`; `region`, `accesskeyid` and
`secretaccesskey` for `s3`; `accountid`, `accesskeyid` and `secretaccesskey` for `r2`; `endpoint` is a
GCS emulator, a custom S3 endpoint, or an R2 endpoint override):

```yaml
integrations:
  objectstoreruntime:
    provider: gcs
    bucket: openlane-system-records
    projectid: openlane-prod
    region: ""
    endpoint: ""
    accountid: ""
    accesskeyid: ""
    secretaccesskey: ""
    import:
      prefix: entities/
      schema: Entity
      variant: vendor
```

The runtime integration registers only when `provider`, `bucket` and `import.prefix` are set. At
registration the client is built for the selected provider; if it cannot be built the server fails to
start, which is the intended gate. For `gcs` the client uses application default credentials. For `s3`
the client uses `accesskeyid` and `secretaccesskey` when both are set, otherwise the default credential
chain (IRSA on EKS, the instance role, or the environment and shared config files). For `r2` the client
uses `accountid` with the keys. An unknown `provider` is refused with `ErrRuntimeProviderUnsupported`.
`SystemImport` is then seeded as a scheduled loop at startup and polls between `mininterval` and
`maxinterval`, backing off while nothing changes. It never runs for an unprovisioned config
(`DisabledForAll`).

Runtime ingest executes as a system-admin caller carrying the virtual integration identity, so the
system-owned hook marks every created row `system_owned = true`, no owner is stamped, and history
tables are skipped. Lookups scope on `system_owned = true`, so re-imports update the same rows. Only
schemas that carry the `system_owned` marker can be imported on this path; anything else is refused
with `ErrIngestSchemaNotSystemOwned`.

Prerequisite for `variant: vendor`: a system-owned `EntityType` named `vendor` must exist (created by a
system admin, so the hook marks it system-owned). Without it the link resolves nothing and entities
are created untyped.

### GKE workload identity (gcs)

Enable Workload Identity Federation for GKE on the cluster and node pool, run the server pod under a
Kubernetes service account, and grant that identity the bucket role. Either bind directly:

```bash
gcloud storage buckets add-iam-policy-binding "gs://$BUCKET" \
  --member="principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/$NAMESPACE/sa/$KSA" \
  --role=roles/storage.objectUser
```

or map the Kubernetes service account to a Google service account (annotate the KSA with
`iam.gke.io/gcp-service-account: $GSA_EMAIL`, grant `roles/iam.workloadIdentityUser` on the GSA to
`serviceAccount:$PROJECT_ID.svc.id.goog[$NAMESPACE/$KSA]`, and give the GSA the bucket role). In both
cases ADC resolves through the metadata server and no config or environment variable is needed.

## Testing locally

Nothing here requires GKE. With `provider: gcs` the runtime client only requires that ADC resolve, and
the GCS client honours the standard local mechanisms in this order:

1. `GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json` for a key file.
2. `gcloud auth application-default login`, which writes user credentials to
   `~/.config/gcloud/application_default_credentials.json`. Add
   `--impersonate-service-account=$GSA_EMAIL` to run with exactly the production identity's
   permissions. Set `objectstoreruntime.projectid` so quota attribution has a project.
3. `STORAGE_EMULATOR_HOST=localhost:4443` with an emulator such as `fsouza/fake-gcs-server`:

   ```bash
   docker run -d -p 4443:4443 fsouza/fake-gcs-server -scheme http -public-host localhost:4443
   export STORAGE_EMULATOR_HOST=localhost:4443
   ```

   When that variable is set the client disables authentication and targets the emulator, so ADC is
   not consulted at all. This is the path for running the scheduled import against local files: create
   the bucket in the emulator, upload `*.json` files under the prefix, set `objectstoreruntime` to
   match, and start the server. The `endpoint` config field is for real alternate endpoints and does
   not disable authentication on its own. Signed URLs do not work against the emulator.

With `provider: s3` and empty keys the runtime client resolves the default credential chain, so
`AWS_PROFILE` or `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` in the environment work locally; set
`endpoint` to point at an S3-compatible emulator. With `provider: r2` the keys are always explicit.

The customer federation path cannot be exercised end to end from a laptop unless Google STS can reach
the issuer's discovery and JWKS endpoints, which means a publicly routable issuer URL (a tunnel to the
local server with `auth.token.issuer` set to the tunnel URL works). Locally, test customer
installations with the service account key, access key or R2 slots, or against the emulator, and
exercise federation in a deployed environment. The assertion minting and STS exchange are covered in isolation by
`pkg/oidc/exchange_test.go`, and the same `auth.FederatedTokenSource` helper is exercised by
`internal/integrations/definitions/gcpscc/federation_test.go`.

Unit tests for the definition (mappings, registration, config resolution, record decoding, the access
key and R2 client builds) and the provider constructors need no credentials and run with the `test`
build tag.
