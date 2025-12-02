-- +goose Up
-- modify "dns_verification_history" table
ALTER TABLE "dns_verification_history" ALTER COLUMN "dns_verification_status" SET DEFAULT 'pending', ALTER COLUMN "acme_challenge_status" SET DEFAULT 'initializing';
-- modify "dns_verifications" table
ALTER TABLE "dns_verifications" ALTER COLUMN "dns_verification_status" SET DEFAULT 'pending', ALTER COLUMN "acme_challenge_status" SET DEFAULT 'initializing';

-- +goose Down
-- reverse: modify "dns_verifications" table
ALTER TABLE "dns_verifications" ALTER COLUMN "acme_challenge_status" SET DEFAULT 'PENDING', ALTER COLUMN "dns_verification_status" SET DEFAULT 'PENDING';
-- reverse: modify "dns_verification_history" table
ALTER TABLE "dns_verification_history" ALTER COLUMN "acme_challenge_status" SET DEFAULT 'PENDING', ALTER COLUMN "dns_verification_status" SET DEFAULT 'PENDING';
