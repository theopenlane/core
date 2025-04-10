-- +goose Up
-- modify "invites" table
ALTER TABLE "invites" ALTER COLUMN "send_attempts" SET DEFAULT 1;
-- modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "send_attempts" bigint NOT NULL DEFAULT 1;

-- +goose Down
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "send_attempts";
-- reverse: modify "invites" table
ALTER TABLE "invites" ALTER COLUMN "send_attempts" SET DEFAULT 0;
