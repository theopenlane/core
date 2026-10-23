-- +goose Up
-- modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "purpose" character varying NULL DEFAULT 'TRUSTCENTER';

-- +goose Down
-- reverse: modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" DROP COLUMN "purpose";
