-- +goose Up
-- modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "pirsch_access_link" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_history" table
ALTER TABLE "trust_center_history" DROP COLUMN "pirsch_access_link";
