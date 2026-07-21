-- +goose Up
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "observation_period_start_date" timestamptz NULL, ADD COLUMN "observation_period_end_date" timestamptz NULL, ADD COLUMN "fieldwork_start_date" timestamptz NULL, ADD COLUMN "fieldwork_end_date" timestamptz NULL;
-- modify "standards" table
ALTER TABLE "standards" ADD COLUMN "priority" bigint NOT NULL DEFAULT 0;
-- modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD CONSTRAINT "trust_center_nda_requests_users_approved_by_user" FOREIGN KEY ("approved_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" DROP CONSTRAINT "trust_center_nda_requests_users_approved_by_user";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP COLUMN "priority";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "fieldwork_end_date", DROP COLUMN "fieldwork_start_date", DROP COLUMN "observation_period_end_date", DROP COLUMN "observation_period_start_date";
