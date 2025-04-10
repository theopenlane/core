-- Modify "invites" table
ALTER TABLE "invites" ALTER COLUMN "send_attempts" SET DEFAULT 1;
-- Modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "send_attempts" bigint NOT NULL DEFAULT 1;
