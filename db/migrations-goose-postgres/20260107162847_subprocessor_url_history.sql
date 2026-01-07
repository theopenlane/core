-- +goose Up
-- modify "dns_verification_history" table
ALTER TABLE "dns_verification_history" ALTER COLUMN "dns_verification_status" SET DEFAULT 'PENDING', ALTER COLUMN "acme_challenge_status" SET DEFAULT 'INITIALIZING';
-- modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "subprocessor_url" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_history" table
ALTER TABLE "trust_center_history" DROP COLUMN "subprocessor_url";
-- reverse: modify "dns_verification_history" table
ALTER TABLE "dns_verification_history" ALTER COLUMN "acme_challenge_status" SET DEFAULT 'initializing', ALTER COLUMN "dns_verification_status" SET DEFAULT 'pending';
