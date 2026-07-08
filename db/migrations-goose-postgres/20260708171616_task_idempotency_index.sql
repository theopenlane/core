-- +goose Up
-- create index "task_owner_id_idempotency_key" to table: "tasks"
CREATE UNIQUE INDEX "task_owner_id_idempotency_key" ON "tasks" ("owner_id", "idempotency_key") WHERE ((deleted_at IS NULL) AND (idempotency_key IS NOT NULL));

-- +goose Down
-- reverse: create index "task_owner_id_idempotency_key" to table: "tasks"
DROP INDEX "task_owner_id_idempotency_key";
