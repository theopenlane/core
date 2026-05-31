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

---

## Pre-release steps

**1. Preview what will be migrated**

```sql
SELECT object_type, COUNT(*) AS row_count
FROM tuple
WHERE relation = 'parent'
  AND _user LIKE 'organization:%'
  AND object_type != 'file'
GROUP BY object_type
ORDER BY object_type;
```

**2. Record the migration start time, then run the migration**

```sql
-- record this value; you will need it for rollback if something goes wrong
SELECT NOW() AS migration_start;

INSERT INTO tuple (store, object_type, object_id, relation, _user, user_type, ulid, inserted_at)
SELECT
    store,
    object_type,
    object_id,
    'parent_context',
    _user,
    user_type,
    generate_ulid(),
    NOW()
FROM tuple
WHERE relation = 'parent'
  AND _user LIKE 'organization:%'
  AND object_type != 'file'
ON CONFLICT (store, object_type, object_id, relation, _user) DO NOTHING;
```

**3. Verify row counts match step 1**

```sql
SELECT object_type, COUNT(*) AS migrated
FROM tuple
WHERE relation = 'parent_context'
  AND _user LIKE 'organization:%'
GROUP BY object_type
ORDER BY object_type;
```

**If something looks wrong — rollback before releasing**

Substitute `$migration_start` with the timestamp recorded in step 2.

```sql
DELETE FROM tuple
WHERE relation = 'parent_context'
  AND _user LIKE 'organization:%'
  AND inserted_at >= '$migration_start';
```

---

## Release

Deploy the updated FGA model after the tuples are written. Deploying the model before the migration means objects will temporarily lose org-context permissions.

---

## Post-release steps

**4. Verify the new model is using `parent_context`**

Spot-check a known object in a known org and confirm permissions resolve correctly. Then confirm the counts are still what you expect:

```sql
SELECT object_type, COUNT(*) AS migrated
FROM tuple
WHERE relation = 'parent_context'
  AND _user LIKE 'organization:%'
GROUP BY object_type
ORDER BY object_type;
```

**5. Clean up old `parent` tuples**

Once the new model is live and verified, the old `parent` + `organization:*` tuples are no longer read by the model and can be deleted:

```sql
DELETE FROM tuple
WHERE relation = 'parent'
  AND _user LIKE 'organization:%'
  AND object_type != 'file';
```

Confirm the expected number of rows were removed:

```sql
SELECT COUNT(*)
FROM tuple
WHERE relation = 'parent'
  AND _user LIKE 'organization:%'
  AND object_type != 'file';
-- should return 0
```
