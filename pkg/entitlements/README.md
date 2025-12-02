# Entitlements

The `entitlements` package wraps Stripe billing APIs and OpenFGA feature
authorization. It manages organization subscriptions, available modules, and
runtime feature checks. This package centralizes the logic for creating and
updating Stripe resources while keeping feature state in OpenFGA with a Redis
cache for _ideally_ faster lookup (not yet benchmarked but a decently safe assumption).

## Purpose

- Provide a Go client for Stripe that exposes helper functions to create
  customers, prices, products and subscriptions
- Process Stripe webhooks to keep feature tuples in sync with purchases
- Offer the `TupleChecker` utility for feature gating via OpenFGA with optional
  Redis caching

A more in‑depth discussion of how modules and entitlements flow through the
system can be found in [`docs/entitlements.md`](../../docs/entitlements.md).

## Approach

The package relies on the official [`stripe-go`](https://github.com/stripe/stripe-go)
client. Helper functions use functional options so callers can supply only the
fields they need. When webhooks are received the handler translates Stripe events
into updates to OpenFGA tuples and refreshes the cached feature list. Feature
checks first consult Redis and fall back to OpenFGA if no cache entry exists.

Key types include:

- `StripeClient` – thin wrapper around `stripe.Client` with higher level helpers
- `OrganizationCustomer` and `Subscription` – internal representations of Stripe
  customers/subscriptions
- `TupleChecker` – verifies feature tuples and creates/deletes them in OpenFGA
  while caching results in Redis

## Integration

Application configuration embeds `entitlements.Config` in the global
[`config.Config`](../../config/config.go) struct. When the HTTP server starts it
constructs a `StripeClient` using that config and stores it on the handler:

```go
h := &handlers.Handler{
    Entitlements: stripeClient,
}
```

The webhook receiver in
[`internal/httpserve/handlers/webhook.go`](../../internal/httpserve/handlers/webhook.go)
uses this client to update subscriptions and feature tuples when events arrive.
Ent hooks in [`ent/hooks/organization.go`](../../ent/hooks/organization.go)
create default feature tuples for new organizations.

## Examples

### Creating a client and customer

```go
sc, err := entitlements.NewStripeClient(
    entitlements.WithAPIKey("sk_test_..."),
)
if err != nil {
    log.Fatal(err)
}

cust, err := sc.CreateCustomer(context.Background(), &entitlements.OrganizationCustomer{
    OrganizationID: "org_123",
    Email:          "billing@example.com",
})
```

### Feature checks with `TupleChecker`

```go
redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
tc := entitlements.NewTupleChecker(
    entitlements.WithRedisClient(redisClient),
    entitlements.WithFGAClient(myFGAClient),
)

allowed, err := tc.CheckFeatureTuple(ctx, entitlements.FeatureTuple{
    UserID:  "user1",
    Feature: "compliance-module",
    Context: map[string]any{"org": "org_123"},
})
```

### Local webhook testing

Install the [Stripe CLI](https://github.com/stripe/stripe-cli) and forward events
to your server:

```bash
stripe login
stripe listen --forward-to localhost:17608/v1/stripe/webhook
```

You can then trigger events, for example:

```bash
stripe trigger payment_intent.succeeded
```

This mirrors the flow described in the "End‑to‑End Flow" section of the
[`docs/entitlements.md`](../../docs/entitlements.md) document.
