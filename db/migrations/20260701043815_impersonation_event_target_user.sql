-- Modify "impersonation_events" table
ALTER TABLE "impersonation_events" DROP CONSTRAINT "impersonation_events_users_targeted_impersonations", ALTER COLUMN "target_user_id" DROP NOT NULL, ADD CONSTRAINT "impersonation_events_users_targeted_impersonations" FOREIGN KEY ("target_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
