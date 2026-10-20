#!/usr/bin/env bash
# configures a Google Cloud project for the objectstore workload identity federation flow and prints the
# integration config to submit; safe to re-run, every step is describe-or-create
#
#   PROJECT_ID=my-test-project ORG_ID=01JXXXXXXXXXXXXXXXXXXXXXXX ISSUER=https://api.example.dev BUCKET=my-records \
#     ./setup-gcp-federation.sh
#
#   SERVICE_ACCOUNT   name of a service account to impersonate (default openlane-storage); set to "" to grant the
#                     federated principal the bucket role directly instead
#   BUCKET_LOCATION   location for a bucket that does not exist yet (default us-central1)
#   PREFIX            object prefix the example records are uploaded under (default vendors/)
#   SEED              upload the example vendor records under PREFIX (default true)

set -euo pipefail

: "${PROJECT_ID:?PROJECT_ID is required}"
: "${ORG_ID:?ORG_ID (the Openlane organization id) is required}"
: "${ISSUER:?ISSUER (the platform auth.token.issuer, publicly reachable) is required}"
: "${BUCKET:?BUCKET is required}"

SERVICE_ACCOUNT=${SERVICE_ACCOUNT-openlane-storage}
BUCKET_LOCATION=${BUCKET_LOCATION:-us-central1}
PREFIX=${PREFIX:-vendors/}
SEED=${SEED:-true}

# the pool and provider ids are fixed in the definition; the assertion audience is derived from them
POOL=openlane
PROVIDER=openlane

log() { printf -- '----> %s\n' "$*" >&2; }

log "checking the issuer is reachable: $ISSUER"
curl -fsS "${ISSUER%/}/.well-known/openid-configuration" >/dev/null
curl -fsS "${ISSUER%/}/.well-known/jwks.json" >/dev/null

log "enabling APIs on $PROJECT_ID"
gcloud services enable iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com storage.googleapis.com \
	--project="$PROJECT_ID" >/dev/null

PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
AUDIENCE="//iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL/providers/$PROVIDER"
PRINCIPAL="principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL/subject/$ORG_ID"

log "bucket gs://$BUCKET"
if ! gcloud storage buckets describe "gs://$BUCKET" --project="$PROJECT_ID" >/dev/null 2>&1; then
	gcloud storage buckets create "gs://$BUCKET" --project="$PROJECT_ID" --location="$BUCKET_LOCATION" \
		--uniform-bucket-level-access >/dev/null
fi

log "workload identity pool $POOL"
if ! gcloud iam workload-identity-pools describe "$POOL" --project="$PROJECT_ID" --location=global >/dev/null 2>&1; then
	gcloud iam workload-identity-pools create "$POOL" --project="$PROJECT_ID" --location=global \
		--display-name="Openlane" >/dev/null
fi

log "oidc provider $PROVIDER trusting $ISSUER"
if ! gcloud iam workload-identity-pools providers describe "$PROVIDER" --project="$PROJECT_ID" --location=global \
	--workload-identity-pool="$POOL" >/dev/null 2>&1; then
	gcloud iam workload-identity-pools providers create-oidc "$PROVIDER" --project="$PROJECT_ID" --location=global \
		--workload-identity-pool="$POOL" \
		--issuer-uri="$ISSUER" \
		--allowed-audiences="$AUDIENCE" \
		--attribute-mapping="google.subject=assertion.sub,attribute.organization_id=assertion.organization_id" \
		--attribute-condition="assertion.sub=='$ORG_ID'" >/dev/null
fi

SA_EMAIL=""

if [ -n "$SERVICE_ACCOUNT" ]; then
	SA_EMAIL="$SERVICE_ACCOUNT@$PROJECT_ID.iam.gserviceaccount.com"

	log "service account $SA_EMAIL"
	if ! gcloud iam service-accounts describe "$SA_EMAIL" --project="$PROJECT_ID" >/dev/null 2>&1; then
		gcloud iam service-accounts create "$SERVICE_ACCOUNT" --project="$PROJECT_ID" \
			--display-name="Openlane storage import" >/dev/null
	fi

	log "granting the service account object access on the bucket"
	gcloud storage buckets add-iam-policy-binding "gs://$BUCKET" --project="$PROJECT_ID" \
		--member="serviceAccount:$SA_EMAIL" --role=roles/storage.objectUser >/dev/null

	log "allowing the federated principal to impersonate the service account"
	gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" --project="$PROJECT_ID" \
		--member="$PRINCIPAL" --role=roles/iam.workloadIdentityUser >/dev/null
else
	log "granting the federated principal object access on the bucket directly"
	gcloud storage buckets add-iam-policy-binding "gs://$BUCKET" --project="$PROJECT_ID" \
		--member="$PRINCIPAL" --role=roles/storage.objectUser >/dev/null
fi

if [ "$SEED" = "true" ]; then
	log "uploading example records to gs://$BUCKET/$PREFIX"
	gcloud storage cp "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"/*.json "gs://$BUCKET/$PREFIX" >/dev/null
fi

log "done"
echo
echo "assertion audience: $AUDIENCE"
echo "federated principal: $PRINCIPAL"
echo
echo "POST /v1/integrations/def_01K0OBJECTSTORE00000000001/config"
cat <<EOF
{
  "credentialRef": "WorkloadIdentityCredentialSchema",
  "body": {
    "projectNumber": "$PROJECT_NUMBER",
    "serviceAccountEmail": "$SA_EMAIL",
    "bucket": "$BUCKET",
    "projectId": "$PROJECT_ID"
  },
  "userInput": {
    "import": { "prefix": "$PREFIX", "schema": "Entity", "variant": "vendor" }
  }
}
EOF
