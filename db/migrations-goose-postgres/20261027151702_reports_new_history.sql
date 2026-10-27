-- +goose Up
-- modify "report_history" table
ALTER TABLE "report_history" ALTER COLUMN "visibility" DROP NOT NULL, ALTER COLUMN "visibility" DROP DEFAULT;

-- +goose Down
-- reverse: modify "report_history" table
ALTER TABLE "report_history" ALTER COLUMN "visibility" SET NOT NULL, ALTER COLUMN "visibility" SET DEFAULT 'PRIVATE';
