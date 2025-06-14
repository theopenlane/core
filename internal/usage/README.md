# Usage Tracking

There could be many reasons why you might want to cap usage of a system - from enforcing limits related to an organization's entitlements based on purchases to preventing run-away record creation (whether unintentional or malicious), having built-in mechansism is ideal. To accomplish this, we've introduced a `Usage` schema object; the usage schema and setup is intended to provide a flexible way to cap the amount of resources that an organization can consume (for example total evidence storage or the number of assets they can create).

Examples of how we've initially structured usage limits:
- Object storage is tracked via `HookUsageStorage` registered on the `File` schema. The hook reads the `persisted_file_size` field and adjusts the
organization's `STORAGE` usage whenever files are created, updated or deleted.
- User counts are updated through `HookUsageUsers` on the `OrgMembership` schema
- Program counts are handled by `HookUsagePrograms` on the `Program` schema.

Both the user and program hooks decrement usage when the corresponding records are removed so the limit reflects the current total, but you can also implement usage tracking in a way that enforces a "global total" (e.g. each creation of a record counts towards the limit, deleting the record doesn't change the usage tracking).

## Usage model

Usage will be tracked per organization and per resource. A new `Usage` entity will hold the current amount used and the allowed limit. Example fields:

```go
func (Usage) Fields() []ent.Field {
    return []ent.Field{
        field.String("organization_id").Comment("owner organization"),
        field.Enum("resource_type").GoType(enums.UsageType("")),
        field.Int64("used").Default(0),
        field.Int64("limit").Default(0),
    }
}
```

The `UsageType` enum lives in `pkg/enums` and defines resource categories such as `STORAGE`, `RECORDS`, `USERS`, and `PROGRAMS`.

## Usage mixin

Instead of manually adding hooks to every schema we want to track, a small `UsageMixin` can embed the logic. Schemas provide the `UsageType` they consume and the mixin registers `HookUsage`:

```go
type File struct {
    ent.Schema
    mixin.UsageMixin{Type: enums.UsageStorage}
}
```

In the above example, any creation of a `File` record will increment the organization's `STORAGE` usage.

## Updating usage

When a feature creates or removes resources it should increment/decrement the corresponding `Usage` row in a transaction. Hooks on the ent schema can enforce that the `used` value does not exceed `limit` and return a descriptive error if the cap has been hit.

## Integrating with entitlements

Subscription tiers or product modules can set default limits for each `UsageType`. When a new organization is created its `Usage` rows are populated from the tier. Limits can then be changed manually or through an upgrade path.

## Setting limits

Limits for each resource type come from a basic configuration. The `subscription` section of `core`'s config accepts a `usageLimits` map keyed by the `UsageType` enum.

```yaml
subscription:
  usageLimits:
    STORAGE: 10737418240   # 10 GiB
    RECORDS: 1000
    USERS: 50
    PROGRAMS: 5
```

During organization provisioning these values should be applied using the helper `usage.InitializeUsageLimits`:

```go
limits := cfg.Entitlements.UsageLimits
_ = usage.InitializeUsageLimits(ctx, client, org.ID, limits)
```

In the future, this will likely be expanded to set default usage limits either per "tier" or per "module".

## Overriding limits

While subscriptions / entitlements can define the starting limits for new organizations, the general goal is that we can override or change them per customer as we desire (e.g. offering credits or expansions for whatever reason we want). Use `usage.SetUsageLimit` to replace a limit or `usage.AddUsageLimit` to grant additional capacity. This is intentionally basic but could be expanded as desired.

```go
// give org extra 50 jigga bytes - those mega gigga bytes son!
err := usage.AddUsageLimit(ctx, client, orgID, enums.UsageStorage, 50<<30)
```

## Unlimited usage

Setting a limit to `0` disables enforcement for that resource. You can remove a cap entirely with `usage.ClearUsageLimit`:

```go
// organization has unlimited storage - power level 9000!
err := usage.ClearUsageLimit(ctx, client, orgID, enums.UsageStorage)
```

Calling `usage.IsUnlimited(u)` on a `Usage` row will return `true` when no limit is configured.

## Surfacing usage

GraphQL resolvers will expose the current usage and remaining capacity so that our front end / clients can display warnings before a hard limit is reached. The `Usage` schema skips mutation generation via `entgql` annotations, ensuring customers cannot alter their usage records through the API. Backend hooks are responsible for updating usage metrics.

## Enforcement hooks

Ent hooks check usage before creating records. For example `HookUsage(enums.UsageRecords)` increments the `RECORDS` counter whenever a new record is inserted. This hook does **not** subtract when records are deleted, allowing lifetime limits to be enforced. A matching privacy rule
`AllowIfWithinUsage` denies the mutation once the limit is reached.

Bulk GraphQL mutations should check limits using usage.CheckUsageDelta before creating records. The helper accepts the number of items to create and ensures the organization has capacity for the entire batch.

## Checking storage usage

To determine how much object storage an organization currently consumes the `usage` package provides a helper:

```go
size, err := usage.OrganizationStorageUsage(ctx, client, orgID)
```

It sums the `persisted_file_size` column of all `File` records owned by the organization and returns the total in bytes.

## Usage threshold events

The usage package can publish events whenever an organization's consumption crosses certain percentages of its limit. Register an event emitter and desired thresholds at startup:

```go
usage.RegisterThresholdEmitter(eventer.Emitter)
usage.RegisterThresholds(enums.UsageStorage, []int{50, 80, 100})
```

Listeners can subscribe to `usage.TopicUsageThreshold` to take action when a threshold is reached (for example sending warning emails):

```go
eventer.Emitter.On(usage.TopicUsageThreshold, myListener)
```

`EmitThresholdEvents` is automatically called when usage is updated, so events fire as soon as limits are approached.
