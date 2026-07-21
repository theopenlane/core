-- +goose Up
-- modify "standards" table
ALTER TABLE "standards" ADD COLUMN "priority" bigint NOT NULL DEFAULT 0;
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "is_suggested" boolean NOT NULL DEFAULT false, ADD COLUMN "priority" bigint NOT NULL DEFAULT 0, ADD COLUMN "available_at" timestamptz NULL, ADD COLUMN "source" character varying NULL, ADD COLUMN "source_key" character varying NULL;
-- create index "task_owner_id_idempotency_key" to table: "tasks"
CREATE UNIQUE INDEX "task_owner_id_idempotency_key" ON "tasks" ("owner_id", "idempotency_key") WHERE ((deleted_at IS NULL) AND (idempotency_key IS NOT NULL));
-- create index "task_owner_id_is_suggested_available_at_priority" to table: "tasks"
CREATE INDEX "task_owner_id_is_suggested_available_at_priority" ON "tasks" ("owner_id", "is_suggested", "available_at", "priority") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "task_owner_id_is_suggested_available_at_priority" to table: "tasks"
DROP INDEX "task_owner_id_is_suggested_available_at_priority";
-- reverse: create index "task_owner_id_idempotency_key" to table: "tasks"
DROP INDEX "task_owner_id_idempotency_key";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "source_key", DROP COLUMN "source", DROP COLUMN "available_at", DROP COLUMN "priority", DROP COLUMN "is_suggested", DROP COLUMN "metadata";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP COLUMN "priority";
