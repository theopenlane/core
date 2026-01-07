-- +goose Up
-- modify "dns_verifications" table
ALTER TABLE "dns_verifications" ALTER COLUMN "dns_verification_status" SET DEFAULT 'PENDING', ALTER COLUMN "acme_challenge_status" SET DEFAULT 'INITIALIZING';
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "subprocessor_url" character varying NULL;

-- +goose Down
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP COLUMN "subprocessor_url";
-- reverse: modify "dns_verifications" table
ALTER TABLE "dns_verifications" ALTER COLUMN "acme_challenge_status" SET DEFAULT 'initializing', ALTER COLUMN "dns_verification_status" SET DEFAULT 'pending';
