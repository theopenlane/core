# Migration to v2 permissions

Migrate `parent` → `parent_context` for Organization Tuples

## Background

The FGA model is being updated to distinguish between structural parent ownership (e.g. a `control` owns a `control_implementation`) and organizational context (e.g. "this object belongs to organization X"). Previously both used the generic `parent` relation; the new model uses `parent_context` specifically for organization-scoped relations.

This allows permission rules to correctly differentiate between "inherit from owning object" and "inherit from org context", which is required for permission scopes and no longer requiring a user "owner" tuple for ownership, and instead can inherit from roles.

The exception being `files`, since files have a more complex relationship and are not always owned by orgs, and can also be owned by users, these are staying as `parent` even for `organization`.

### What to migrate

All tuples where:
- `relation = 'parent'`
- `_user LIKE 'organization:%'`
- `object_type != 'file'` (files use a separate ownership model)

### Steps

**1. Preview the rows that will be migrated**

```sql
SELECT
    store,
    object_type,
    object_id,
    'parent_context' AS relation,
    _user,
    user_type
FROM tuple
WHERE relation = 'parent'
  AND _user LIKE 'organization:%'
  AND object_type != 'file';
```

**2. Run the migration**

```sql
INSERT INTO tuple (store, object_type, object_id, relation, _user, user_type, ulid, inserted_at)
SELECT
    store,
    object_type,
    object_id,
    'parent_context',
    _user,
    user_type,
    md5(store || object_type || object_id || 'parent_context' || _user),
    NOW()
FROM tuple
WHERE relation = 'parent'
  AND _user LIKE 'organization:%'
  AND object_type != 'file'
ON CONFLICT (store, object_type, object_id, relation, _user) DO NOTHING;
```

The `ulid` is derived deterministically from the natural key so the insert is idempotent — safe to re-run.

**3. Verify**

Spot-check that `parent_context` rows now exist for the same objects that had `parent` rows:

```sql
SELECT object_type, COUNT(*)
FROM tuple
WHERE relation = 'parent_context'
  AND _user LIKE 'organization:%'
GROUP BY object_type
ORDER BY object_type;
```

**4. Deploy the updated FGA model**

The new model must be deployed after the tuples are written. Deploying the model before the migration means objects will temporarily lose org-context permissions.

**5. (Optional) Clean up old `parent` tuples**

Once the new model is live and verified, the old `parent` + `organization:*` tuples are no longer used and can be deleted:

```sql
DELETE FROM tuple
WHERE relation = 'parent'
  AND _user LIKE 'organization:%'
  AND object_type != 'file';
```
