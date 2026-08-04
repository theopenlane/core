-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "observation_period_start_date" timestamptz NULL, ADD COLUMN "observation_period_end_date" timestamptz NULL, ADD COLUMN "fieldwork_start_date" timestamptz NULL, ADD COLUMN "fieldwork_end_date" timestamptz NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "is_suggested" boolean NOT NULL DEFAULT false, ADD COLUMN "priority" bigint NOT NULL DEFAULT 0, ADD COLUMN "source" character varying NULL, ADD COLUMN "source_key" character varying NULL;
-- Create index "task_owner_id_idempotency_key" to table: "tasks"
CREATE UNIQUE INDEX "task_owner_id_idempotency_key" ON "tasks" ("owner_id", "idempotency_key") WHERE ((deleted_at IS NULL) AND (idempotency_key IS NOT NULL));
-- Create index "task_owner_id_is_suggested_priority" to table: "tasks"
CREATE INDEX "task_owner_id_is_suggested_priority" ON "tasks" ("owner_id", "is_suggested", "priority") WHERE (deleted_at IS NULL);
-- Modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD CONSTRAINT "trust_center_nda_requests_users_approved_by_user" FOREIGN KEY ("approved_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
