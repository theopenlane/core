# Entitlements Catalog

## Overview

## JSONSchema generation

`go run pkg/catalog/genjsonschema/catalog_schema.go` -> generates a jsonschema from the structs inside of catalog.go.

This jsonschema is how we validate that the yaml input is indeed valid and coforms to the specification. The `LoadCatalog` function will fail if the yaml input does not conform to the schema specification, offering some guardrails in the event of misconfiguration / bad yaml / missing fields, etc.

` go run ./cmd/catalog --stripe-key="[insertstrpekey]"` will take the catalog file, pull products and prices from stripe, and prompt you as to whether or not:
- the definitions in your catalog file have corresponding products and prices in the stripe instance matching the key you provided
-


```bash
➜  core git:(feat-modules) ✗ go run ./cmd/catalog --stripe-key=""
module trust_center         product:false missing_prices:2
module base                 product:false missing_prices:1
module compliance           product:true missing_prices:1
addon domain_scanning      product:false missing_prices:2
Create missing products and prices? (y/N):
```

After you've selected `y` to create the missing products and prices, you'll see an output in your terminal and you can confirm in the Stripe UI product catalog that the products and prices were created successfully. The way you can easily distinguish which products and prices were created through this process is by looking for the `managed_by` metadata and seeing that the `managed_by` includes `module-manager`. The goal of this piece of metadata is work similarly to how kubernetes controllers (and other systems) work where they use annotations / labels to indicate that the resource is controlled declaratively. This helps create the distinction between products or prices that were created manually, and helps prevent drift of products and prices which may have been manually changed in the system and do not match the declarative configuration.


```bash
➜  core git:(feat-modules) ✗ go run ./cmd/catalog --stripe-key=""
module base                 product:true missing_prices:0
module compliance           product:true missing_prices:0
module trust_center         product:true missing_prices:2
addon domain_scanning      product:false missing_prices:2
Create missing products and prices? (y/N):

```

If there are already matching prices in the system which share a lookupkey but do NOT have the managed metadata, you will be prompted as to whether or not you want to "take over" management of these resources. The implication being that not only will the metadata be updated, but if the price's product association does not match the declarative file, it will be changed

```bash
➜  core git:(feat-modules) ✗ go run ./cmd/catalog --stripe-key=""
module base                 product:true missing_prices:0
module compliance           product:true missing_prices:0
module trust_center         product:true missing_prices:0
addon domain_scanning      product:false missing_prices:2
┌──────────────┬────────────────────────────┬────────────────────────────────┬─────────┐
│ FEATURE      │ LOOKUP KEY                 │ PRICE ID                       │ MANAGED │
├──────────────┼────────────────────────────┼────────────────────────────────┼─────────┤
│ trust_center │ price_trust_center_monthly │ price_1RaiEgR7q8Ny5Jw0bGDA8we8 │         │
├──────────────┼────────────────────────────┼────────────────────────────────┼─────────┤
│ trust_center │ price_trust_center_yearly  │ price_1RaiFmR7q8Ny5Jw0PvAVtiGX │         │
└──────────────┴────────────────────────────┴────────────────────────────────┴─────────┘

Take over these prices by adding metadata? (y/N):
```

Once you've successfully reconciled the products and prices in an instance, you should be able to re-run the command and see there are no actions which need to be taken. If, however, you don't have any products or prices which need updated but do NOT have the price ID included inside of the catalog, the default behavior is to write back to the catalog file to ensure the price ID's are included.

```bash
➜  core git:(feat-modules) ✗ go run ./cmd/catalog --stripe-key=""
module base                 product:true missing_prices:0
module compliance           product:true missing_prices:0
module trust_center         product:true missing_prices:0
addon domain_scanning      product:true missing_prices:0
```

If you didn't have price ID's already inside, you should see something like this in your git diff:

```bash
-        - interval: month
-          unit_amount: 10000
-          nickname: price_compliance_month
-          lookup_key: price_compliance_month
-          metadata:
-            tier: base
-        - interval: year
-          unit_amount: 100000
-          nickname: compliance-annual
-          lookup_key: compliance-annual
-          metadata:
-            tier: base
+      - interval: month
+        unit_amount: 10000
+        nickname: price_compliance_month
+        lookup_key: price_compliance_month
+        metadata:
+          tier: base
+        price_id: price_1RbjHUR7q8Ny5Jw09vtTo1rQ
+      - interval: year
+        unit_amount: 100000
+        nickname: compliance-annual
+        lookup_key: compliance-annual
+        metadata:
+          tier: base
+        price_id: price_1RdciXR7q8Ny5Jw0hfaLuSef
+    audience: public
```

### Modifying existing files

If you've modified the price defined in a catalog which already exists in stripe, you cannot modify the existing price (especially if it's been used in a subscription). You have to create a new price and migrate customers over to it. You'll see warnings within the terminal output to this effect:

```bash
module base                 product:true missing_prices:0
[WARN] price price_1RbjHUR7q8Ny5Jw09vtTo1rQ for feature compliance does not match catalog; to modify an existing price create a new one and update subscriptions
module compliance           product:true missing_prices:0
[WARN] price price_1RaiEgR7q8Ny5Jw0bGDA8we8 for feature trust_center does not match catalog; to modify an existing price create a new one and update subscriptions
[WARN] price price_1RaiFmR7q8Ny5Jw0PvAVtiGX for feature trust_center does not match catalog; to modify an existing price create a new one and update subscriptions
module trust_center         product:true missing_prices:0
[WARN] price price_1RddewR7q8Ny5Jw0MQVf1GOp for feature domain_scanning does not match catalog; to modify an existing price create a new one and update subscriptions
addon domain_scanning      product:true missing_prices:0
```

## Creating webhook endpoints

Use `CreateWebhookEndpoint` to programmatically register a webhook URL with Stripe. Provide the URL and the events you want delivered. The call returns the created endpoint including the signing secret that should be used to verify incoming requests.

## Catalog Validation

```bash
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Catalog validation failed - ensure you have generated the latest schema file if you have modified the catalog structs"}
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Validation error: (root): version is required"}
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Validation error: modules.entity-management: description is required"}
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Validation error: modules.entity-management: Additional property Description is not allowed"}
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Validation error: modules.entity-management.billing.prices.0: unit_amount is required"}
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Validation error: modules.entity-management.billing.prices.0: Additional property Unit_amount is not allowed"}
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Validation error: modules.entity-management.billing.prices.0: Additional property Nickname is not allowed"}
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Validation error: modules.entity-management.billing.prices.0: Additional property Lookup_key is not allowed"}
{"level":"debug","time":"2025-06-25T13:04:20-05:00","message":"Validation error: modules.entity-management.billing.prices.1: unit_amount is required"}

load catalog: catalog validation failed
exit status 1
```

## Catalog Versioning

The `SaveCatalog` method is responsible for saving a `Catalog` struct to disk in YAML format, while also managing versioning and integrity via a SHA hash. The function first checks if the receiver (`c`) is `nil`, in which case it returns immediately with no error. It then attempts to read the existing catalog file from the provided path. If the file does not exist, that's acceptable, but any other read error is returned.

If the file exists and contains data, it is unmarshaled from YAML into a temporary `Catalog` struct called `orig`. This allows the function to compare the current catalog with the previous version. If the current catalog's version is unset but the original has one, the version is carried over. Similarly, if the SHA is missing, it is computed based on the version string.

The catalog is then marshaled into YAML. If the new YAML differs from the original file's contents, the function attempts to bump the patch version (using semantic versioning) and recomputes the SHA. The catalog is re-marshaled to reflect these changes. A unified diff is generated between the old and new YAML representations, providing a summary of what changed.

Finally, the new YAML data is written to disk with standard permissions. The function returns the diff string, which can be used for logging or review. This approach ensures that the catalog file is always up-to-date, versioned, and its integrity can be verified via the SHA. A subtle point is that the version is only bumped if the contents have changed, helping to avoid unnecessary version increments.
