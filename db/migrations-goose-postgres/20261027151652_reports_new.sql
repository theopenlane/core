-- +goose Up
-- modify "reports" table
ALTER TABLE "reports" ALTER COLUMN "visibility" DROP NOT NULL, ALTER COLUMN "visibility" DROP DEFAULT;

-- +goose Down
-- reverse: modify "reports" table
ALTER TABLE "reports" ALTER COLUMN "visibility" SET NOT NULL, ALTER COLUMN "visibility" SET DEFAULT 'PRIVATE';
