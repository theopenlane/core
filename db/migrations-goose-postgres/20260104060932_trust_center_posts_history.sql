-- +goose Up
-- modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "trust_center_id" character varying NULL;

-- +goose Down
-- reverse: modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "trust_center_id";
