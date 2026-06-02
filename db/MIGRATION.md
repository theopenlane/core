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

## Prerequisites

Install the `generate_ulid()` function first:

```sql
CREATE OR REPLACE FUNCTION generate_ulid() RETURNS TEXT AS $$
DECLARE
  encoding  BYTEA = '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
  output    TEXT = '';
  unix_time BIGINT;
  ulid      BYTEA;
BEGIN
  unix_time = (EXTRACT(EPOCH FROM NOW()) * 1000)::BIGINT;
  ulid = substring(int8send(unix_time) FROM 3);
  ulid = ulid || gen_random_bytes(10);

  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 0) & 224) >> 5));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 0) & 31)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 1) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 1) & 7) << 2) | ((GET_BYTE(ulid, 2) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 2) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 2) & 1) << 4) | ((GET_BYTE(ulid, 3) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 3) & 15) << 1) | ((GET_BYTE(ulid, 4) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 4) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 4) & 3) << 3) | ((GET_BYTE(ulid, 5) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 5) & 31)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 6) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 6) & 7) << 2) | ((GET_BYTE(ulid, 7) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 7) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 7) & 1) << 4) | ((GET_BYTE(ulid, 8) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 8) & 15) << 1) | ((GET_BYTE(ulid, 9) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 9) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 9) & 3) << 3) | ((GET_BYTE(ulid, 10) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 10) & 31)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 11) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 11) & 7) << 2) | ((GET_BYTE(ulid, 12) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 12) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 12) & 1) << 4) | ((GET_BYTE(ulid, 13) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 13) & 15) << 1) | ((GET_BYTE(ulid, 14) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 14) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 14) & 3) << 3) | ((GET_BYTE(ulid, 15) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 15) & 31)));

  RETURN output;
END
$$ LANGUAGE plpgsql VOLATILE;
```

Requires `pgcrypto` (`CREATE EXTENSION IF NOT EXISTS pgcrypto;`).

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

INSERT INTO tuple (store, object_type, object_id, relation, _user, user_type, ulid, inserted_at, condition_name, condition_context)
SELECT
    store,
    object_type,
    object_id,
    'parent_context',
    _user,
    user_type,
    generate_ulid(),
    NOW(),
    condition_name,
    condition_context
FROM tuple
WHERE relation = 'parent'
  AND _user LIKE 'organization:%'
  AND object_type != 'file'
ON CONFLICT (store, object_type, object_id, relation, _user) DO NOTHING;

-- record this value; you will need it for rollback if something goes wrong
SELECT NOW() AS migration_end;
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

Substitute `$migration_start` and `$migration_end` with the timestamps recorded in step 2.

```sql
DELETE FROM tuple
WHERE relation = 'parent_context'
  AND _user LIKE 'organization:%'
  AND inserted_at BETWEEN '$migration_start' AND '$migration_end';
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
