# Stripe integration



## Redis cache

The CheckFeatureTuple function is responsible for determining whether a specific feature tuple exists, using a two-step approach: it first checks a Redis cache, and if the tuple is not found there, it queries an external FGA (Fine-Grained Authorization) system. This method is part of the TupleChecker struct, which holds references to a Redis client, an FGA client, and a cache time-to-live (TTL) duration.

The function begins by validating that both the Redis client and the FGA client are properly configured. If either is missing, it returns an error immediately, preventing further execution. Next, it generates a cache key based on the tuple's contents using the cacheKey method, which ensures that each tuple maps to a unique Redis key.

It then attempts to retrieve the tuple's existence status from Redis. If the key is found (err == nil), it returns true if the cached value is "1", or false otherwise. If the error is anything other than a cache miss (redis.Nil), it returns the error, as this indicates a problem with the Redis operation.

If the tuple is not present in the cache, the function queries the FGA system by calling CheckTuple. If this check fails, it returns the error. Otherwise, it caches the result in Redis for future requests, storing "1" for a positive result and "0" for a negative one, using the configured TTL. Finally, it returns the result of the FGA check.

This approach optimizes performance by reducing the number of expensive FGA checks, relying on Redis as a fast, in-memory cache. It also ensures that the cache is kept up-to-date with the latest results, improving efficiency for repeated queries. One subtlety is that the function does not handle errors from the Redis Set operation, which could be a point of improvement if cache consistency is critical.

## Webhook

To test webhooks locally:

`brew install stripe/stripe-cli/stripe`

`stripe login` and then use your associated credentials

`stripe listen --forward-to localhost:17608/v1/stripe/webhook`

`stripe trigger payment_intent.succeeded`

### Unprocessed

If your webhook endpoint temporarily can’t process events, Stripe automatically resends the undelivered events to your endpoint for up to three days, increasing the time for your webhook endpoint to eventually receive and process all events.

stripe fixtures ./fixtures.json

stripe trigger payment_intent.succeeded --override payment_intent:amount=5000 --override payment_intent:currency=usd --add payment_intent:customer=cus_xxx

https://github.com/stripe/stripe-cli/tree/master/pkg/fixtures/triggers

### Subscription status

Basic information pulled up out of stripe's docs indicating the various scenarios in which a subscription can be in a given status

https://stripe.com/docs/billing/subscriptions/overview#subscription-status
