-- Modify "job_results" table
ALTER TABLE "job_results" ADD COLUMN "log" text NULL;
-- Drop index "job_runners_ip_address_key" from table: "job_runners"
DROP INDEX "job_runners_ip_address_key";
-- Modify "job_runners" table
ALTER TABLE "job_runners" ALTER COLUMN "ip_address" DROP NOT NULL, ADD COLUMN "last_seen" timestamptz NULL, ADD COLUMN "version" character varying NULL, ADD COLUMN "os" character varying NULL;
