# Catalog CLI Tools

This directory contains two small command line utilities used for working with the
billing catalog stored in [`pkg/catalog/catalog.yaml`](../../pkg/catalog/catalog.yaml)
and Stripe.

- `catalog` – reconciles the local catalog file with your Stripe account.
- `pricemigrate` – tags a replacement price and optionally migrates active
  subscriptions from one price ID to another.

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
