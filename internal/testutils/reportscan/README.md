# Report Scan Test Page

A standalone page for exercising the report scan flow end to end: upload a SOC 2 report PDF, let the backend parse it through the report scan listener, and inspect the parsed sections stored on the scan.

## Prerequisites

- Core running on `http://localhost:17608` with the integrations runtime enabled
- `integrations.gemini` provisioned for one of the two backends, since the parse operations are disabled otherwise:
  - `backend: gemini` needs `apikey`
  - `backend: vertex` needs `project` and `location`; credentials come from `credentialsjson`
- `integrations.gemini.prompts` populated, since a section with no configured prompt is never parsed
- Either a PAT / API token with write access to the organization

## Running

From the root of the repository:

```bash
task testui:reportscan
```

Open http://localhost:5500/reportscan/ in a browser. Port 5500 is already an allowed CORS origin in the local dev config. The server hosts the whole `internal/testutils` directory so the shared bootstrap stylesheet resolves.

## Using the page

1. Set the API host and, if not using cookies, paste a PAT or API token. A PAT also needs the organization ID.
2. Choose a PDF and click **Submit report**. The page sends `createScan` with `scanType: REPORT`, `performedBy: openlane_report_parse`, and the file under `scanFiles`.
3. The page polls the scan every ten seconds until it reaches `COMPLETED` or `FAILED`, then renders each section from `metadata.report` as JSON.
4. **List report scans** shows the twenty most recent report scans for the organization.

## What the scan stores

- `metadata.report` holds the parsed sections, one key per section
- `metadata.report_summary` holds each section's state, item count, error, and batch progress, updated as jobs land, so it shows progress while the scan is still running
- `metadata.report_validation` holds the SOC 2 validation outcome for the upload
- `metadata.error` holds the failure reason when the scan fails

## What to expect

- An upload that is not a PDF is rejected by the `scanFiles` MIME validator, and a password protected PDF is rejected by the password validator.
- A PDF that does not look like a SOC 2 report is rejected before any parsing, with the reason on `metadata.error` and the detail on `metadata.report_validation`.
- When `integrations.gemini.modelarmortemplate` is set, a PDF carrying instructions aimed at the model is rejected at submit.
- A second submission within 24 hours for the same organization is rejected by the parse operation's rate limit policy unless the previous scan failed.
- A failed parse stores the cause under `metadata.error` and marks the scan `FAILED`, so it can be resubmitted immediately.
- Parsing runs one job per configured section, and the reviews and findings sections fan out into batches of controls, so a large report is dozens of model calls and can take several minutes.
