-- +goose Up
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
-- modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';

-- +goose Down
-- reverse: modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" SET DEFAULT 'READY';
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" SET DEFAULT 'READY';
