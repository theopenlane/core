-- Modify "invites" table
ALTER TABLE "invites" ADD COLUMN "ownership_transfer" boolean NULL DEFAULT false;
