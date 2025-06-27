# Transitioning to Module Catalog Pricing

This document captures remaining tasks for completing the move from the legacy tier-based pricing model to the new module catalog approach.

## Goals

- **Subscription sync** – Keep organization subscriptions and customers aligned with Stripe data at all times.
- **Trial on sign up** – Automatically create a subscription when a new organization signs up and grant a trial of the compliance module.
- **Module purchases** – Allow customers to buy additional modules and update their entitlements accordingly.
- **Feature lookup** – Provide a simple way to query what features an organization currently has enabled.

## Outstanding work

- **Remove tier references**
  - Clean up `product_tier` fields and replace them with module identifiers.
  - Update any privacy rules, interceptors or code paths that still rely on tiers.
- **Finalize catalog loader**
  - Ensure the catalog is loaded at start‑up and validated against `jsonschema/catalog.schema.json`.
  - Auto-create Stripe products/prices for any catalog entries missing in Stripe.
- **Subscription creation flow**
  - Hook organization creation events so that `FindOrCreateCustomer` also provisions the trial compliance subscription.
  - Store resulting Stripe IDs on `org_subscription` records.
- **Webhook handling**
  - Extend `handleSubscriptionUpdated` to map purchased module prices back to feature lookup keys.
  - Maintain OpenFGA tuples via `ensureFeatureTuples` whenever subscriptions change.
- **Feature APIs**
  - Update `AccountFeaturesHandler` to read enabled features from FGA instead of the `org_subscriptions` table.
  - Add integration tests covering feature cache updates and FGA writes.
- **UI / portal links**
  - Surface catalog data and purchase links via the `/v1/catalog` endpoint.
  - Expose Stripe billing portal URLs on subscription records so customers can manage payment methods.

Completing the above will fully migrate the service to module‑based pricing with a declarative catalog while keeping all customer data in sync with Stripe.
