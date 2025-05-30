-- +goose Up
-- modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" ADD COLUMN "job_handle" character varying NOT NULL;
-- modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" ADD COLUMN "job_handle" character varying NOT NULL;

-- +goose Down
-- reverse: modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" DROP COLUMN "job_handle";
-- reverse: modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" DROP COLUMN "job_handle";
