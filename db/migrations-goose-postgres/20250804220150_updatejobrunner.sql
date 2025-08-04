-- +goose Up
-- modify "job_results" table
ALTER TABLE "job_results" ADD COLUMN "log" text NULL;
-- drop index "job_runners_ip_address_key" from table: "job_runners"
DROP INDEX "job_runners_ip_address_key";
-- modify "job_runners" table
ALTER TABLE "job_runners" ALTER COLUMN "ip_address" DROP NOT NULL, ADD COLUMN "last_seen" timestamptz NULL, ADD COLUMN "version" character varying NULL, ADD COLUMN "os" character varying NULL;

-- +goose Down
-- reverse: modify "job_runners" table
ALTER TABLE "job_runners" DROP COLUMN "os", DROP COLUMN "version", DROP COLUMN "last_seen", ALTER COLUMN "ip_address" SET NOT NULL;
-- reverse: drop index "job_runners_ip_address_key" from table: "job_runners"
CREATE UNIQUE INDEX "job_runners_ip_address_key" ON "job_runners" ("ip_address");
-- reverse: modify "job_results" table
ALTER TABLE "job_results" DROP COLUMN "log";
