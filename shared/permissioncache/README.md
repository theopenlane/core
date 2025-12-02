# Permissions Cache

This package contains helpers for caching permissions, such as enabled
features for an organization. `Cache` is a small wrapper around Redis that can
store things such as a set of feature names keyed by organization ID, user roles
within an organization, etc.. Entries expire independently from session data
and can be fetched with `GetX` or written with `SetX`.

```go
c := features.NewCache(redisClient, permissioncache.WithCacheTTL(time.Minute))
features := []models.OrgModule{models.OrgModule("evidence"), models.OrgModule("search")}
err := c.SetFeatures(ctx, "org1", features)
if err != nil {
    return err
}

feats, err := c.GetFeatures(ctx, "org1")
if err != nil {
    return err
}
```

Use `permissioncache.WithCache` and `permissioncache.CacheFromContext` to make the cache
available throughout a request lifecycle.