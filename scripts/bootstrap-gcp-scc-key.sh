#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  cat <<'USAGE' >&2
Usage: bootstrap-gcp-scc-key.sh <PROJECT_ID> <ORG_ID> [SERVICE_ACCOUNT_NAME]

Example:
  ./scripts/bootstrap-gcp-scc-key.sh my-scc-project 123456789012 openlane-scc-runner
USAGE
  exit 1
fi

PROJECT_ID="$1"
ORG_ID="$2"
SA_NAME="${3:-openlane-scc-runner}"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
KEY_FILE="$(mktemp -t "${SA_NAME}-key.XXXXXX.json")"

echo "Ensuring APIs (securitycenter, iam) are enabled in ${PROJECT_ID}..."
gcloud services enable securitycenter.googleapis.com iam.googleapis.com \
  --project "${PROJECT_ID}" >/dev/null

echo "Creating service account ${SA_EMAIL} (if it doesn't already exist)..."
gcloud iam service-accounts create "${SA_NAME}" \
  --project "${PROJECT_ID}" \
  --display-name "Openlane SCC Runner" \
  --description "Temporary SCC integration account" >/dev/null || true

echo "Granting SCC viewer roles at organization ${ORG_ID}..."
gcloud organizations add-iam-policy-binding "${ORG_ID}" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/securitycenter.findingsViewer" >/dev/null
gcloud organizations add-iam-policy-binding "${ORG_ID}" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/securitycenter.sourcesViewer" >/dev/null

echo "Generating service-account key..."
gcloud iam service-accounts keys create "${KEY_FILE}" \
  --iam-account="${SA_EMAIL}" \
  --project="${PROJECT_ID}"

cat <<EOF

Done. Paste the contents of ${KEY_FILE} into the "Service Account Key JSON" field
in the Openlane SCC integration form, along with:
  projectId              = ${PROJECT_ID}
  organizationId         = ${ORG_ID}
  serviceAccountEmail    = ${SA_EMAIL} (optional but recommended)

Remember to delete the key when finished testing:
  gcloud iam service-accounts keys delete KEY_ID --iam-account ${SA_EMAIL}

EOF
