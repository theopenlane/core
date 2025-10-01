# Catalog CLI Tools

This directory contains two small command line utilities used for working with the
billing catalog stored in [`pkg/catalog/catalog.yaml`](../../pkg/catalog/catalog.yaml)
and Stripe. This is a USE AT YOUR OWN RISK type of tool; we use it for ourselves at Openlane,
but it's not created with the intent of being a generic integration utility. Use of this tool
requires interactions with the `pkg/catalog` and `pkg/entitlements` packages in this repo.

High level:

- `catalog` – reconciles the local catalog file with your Stripe account
- `migrate` – tags a replacement price and optionally migrates active
  subscriptions from one price ID to another

Both binaries rely on a Stripe API key which can be supplied via the
`--stripe-key` flag or the `STRIPE_API_KEY` environment variable.

## `catalog`

```
Usage: catalog [options]
```

The catalog command compares the catalog file with the existing products and
prices in your Stripe account. It prints a table summarizing which features have
matching products and prices. If a product or price is missing you may be
prompted to create it. Existing prices that match by lookup key but are missing
the `managed_by` metadata can also be "taken over" so future runs treat them as
managed by the catalog.

Flags:

- `--catalog` – Path to the catalog YAML file. Default is
  `./pkg/catalog/catalog.yaml`.
- `--stripe-key` – Stripe API key. Can also be set via `STRIPE_API_KEY`.
- `--takeover` – Automatically take over unmanaged prices without prompting.
- `--write` – After reconciliation, write discovered price IDs back to the
  catalog file.

A typical flow is:

1. Run `catalog` with your Stripe key to see a report of the current state.
2. Choose whether to create missing products/prices and take over unmanaged ones.
3. If `--write` is specified, any new price IDs are persisted back into the
   catalog file along with a bumped version number.

### Catalog Validation

One of the first things that occurs within the code logic is that the Catalog yaml file is loaded and compared against the JSONSchema published for it. If

## `pricemigrate`

```
Usage: pricemigrate --old-price <id> --new-price <id> [options]
```

This utility assists with migrating customers from one recurring price to
another. It first tags the old and new prices so that reports can identify them
as a migration pair. It can then update the provided customer subscriptions to
use the new price.

Flags:

- `--old-price` – Price ID customers are currently subscribed to. **Required**.
- `--new-price` – Replacement price ID. **Required**.
- `--customers` – Comma separated list of customer IDs to migrate.
- `--stripe-key` – Stripe API key (or set `STRIPE_API_KEY`).
- `--no-migrate` – Only tag the prices; do not update any subscriptions.
- `--dry-run` – List subscriptions that would be migrated without changing them.

Example dry run:

```
pricemigrate --old-price price_old --new-price price_new \
  --customers cus_A,cus_B --dry-run
```

This prints a table of customers and subscription IDs that reference the old
price. When run without `--dry-run` and with `--no-migrate` omitted, each
subscription containing the old price is updated in place to reference the new
one.

### Catalog example

When the catalog command is run, the output will depend on your specific circumstance. This tool was made with the following circumstances in mind:

- You have 0 products and prices in your stripe account (or are seeding a sandbox stripe instance)
- You have a >0 amount of products in prices in your stripe account, made by this tool
- You have a >0 amounnt of products and prices in your stripe account, not made by this tool
- Some combination of the above

#### Accounts with products and prices already

The logic of the CLI should follow roughly this matching pattern:

- If a stripe price ID is found, we'll use that to find a corresponding price - if one is found with that ID, but does not match the catalog, you'll receive a warning
- If no stripe ID is for price supplied, we'll attempt a lookup into stripe by order of most unique to least unique variables - lookup key is our primary mechanism because while we can control product ID's created in the system, we cannot control price ID's so the next most unique ref we can use is lookup key, then metadata we control, then display name

This is what the output will look like if you've already created some products and prices (created by this tool or not):

```bash
➜  core git:(feat-modules) ✗ go run ./cmd/catalog --stripe-key=""
module trust_center         product:false missing_prices:2
module base                 product:false missing_prices:1
module compliance           product:true missing_prices:1
addon domain_scanning      product:false missing_prices:2
Create missing products and prices? (y/N):
```

After you've selected `y` to create the missing products and prices, you'll see an output in your terminal and you can confirm in the Stripe UI product catalog that the products and prices were created successfully. The way you can easily distinguish which products and prices were created through this process is by looking for the `managed_by` metadata and seeing that the `managed_by` includes `module-manager`. The goal of this piece of metadata is work similarly to how kubernetes controllers (and other systems) work where they use annotations / labels to indicate that the resource is controlled declaratively. This helps create the distinction between products or prices that were created manually, and helps prevent drift of products and prices which may have been manually changed in the system and do not match the declarative configuration.

For the sake of this example, the command was re-run after manually creating products and prices using stripe's UI, but showing that the utility should correctly match if the metadata lines up:

```bash
➜  core git:(feat-modules) ✗ go run ./cmd/catalog --stripe-key=""
module base                 product:true missing_prices:0
module compliance           product:true missing_prices:0
module trust_center         product:true missing_prices:2
addon domain_scanning      product:false missing_prices:2
Create missing products and prices? (y/N):

```

If there are already matching prices in the system which share a lookupkey but do NOT have the managed metadata, you will be prompted as to whether or not you want to "take over" management of these resources. This is a NON-DESTRUCTIVE operation, it's intent is to update the metadata to show its `managed_by` the module-manager (and make it easier for it to reconcile point in time forward):

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

This means if the tool is run as a part of a CI/CD process, you'll have a failing build step (or want to commit the diff back) to ensure you're retaining the price ID for future use.

#### Modifying existing files

You should consider Stripe prices as immutable. Once they've been created, they cannot be modified. Given that, if you alter a price in the catalog which already exists in stripe and has been sync'd by this tool, it will warn you that modifying a price is not possible and to modify you must create a new price and migrate subscriptions over to it.

> WARNING: TEST THE MIGRATION OF PRICES THOROUGHLY - you want to ensure you understand not just re-attaching prices to an account, but whether or not the billing cycle needs to be re-set and other considerations based on your circumstance

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

## Add module to customer subscription

In the event you're wanting to test adding a module to a subscription (in either a sandbox or production) you can use this CLI interactively to find + update a customer subscription. You need to pass in the organization ID you're trying to update (which is the customer name in stripe) and your API key. You'll be presented with an interactive prompt displaying available modules to add to the subscription (which will automatically exclude the ones on the subscription already) and the price you'd like to associate (which should be the monthly and yearly options for the product).

```bash
➜  core git:(feat-cataloglocal) ✗ go run cmd/catalog/main.go --add-module --org-id="01K5EYSC7B20GJY6NJ2ZCF2YYC" --stripe-key="yourkey"
Found customer: 01K5EYSC7B20GJY6NJ2ZCF2YYC
Current subscription: sub_1S8lrpJIzM4Pa2ZcKei6suR6 (status: trialing)
✅ Vulnerability Tracking and Management - Vulnerability Tracking and Management
✅ $100.00/month (price_vulnerability_mgmt_monthly)
Successfully added module 'Vulnerability Tracking and Management' to subscription
```

While it's not difficult to add a module + price via stripe's UI, it be somewhat annoying to have to set folks up with access to a sandbox account and also isn't conducive for basic automation with something like Taskfile.
