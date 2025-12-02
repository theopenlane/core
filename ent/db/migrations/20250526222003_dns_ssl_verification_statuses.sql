-- Modify "dns_verification_history" table
ALTER TABLE "dns_verification_history" ALTER COLUMN "dns_verification_status" SET DEFAULT 'pending', ALTER COLUMN "acme_challenge_status" SET DEFAULT 'initializing';
-- Modify "dns_verifications" table
ALTER TABLE "dns_verifications" ALTER COLUMN "dns_verification_status" SET DEFAULT 'pending', ALTER COLUMN "acme_challenge_status" SET DEFAULT 'initializing';
