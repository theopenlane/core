-- Modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD CONSTRAINT "trust_center_nda_requests_users_approved_by_user" FOREIGN KEY ("approved_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
