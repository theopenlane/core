-- +goose Up
-- modify "invites" table
ALTER TABLE "invites" ADD COLUMN "ownership_transfer" boolean NULL DEFAULT false;

-- +goose Down
-- reverse: modify "invites" table
ALTER TABLE "invites" DROP COLUMN "ownership_transfer";
