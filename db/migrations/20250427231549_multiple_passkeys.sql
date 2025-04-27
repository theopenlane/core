-- Modify "webauthns" table
ALTER TABLE "webauthns" DROP CONSTRAINT "webauthns_users_webauthn", ADD CONSTRAINT "webauthns_users_webauthns" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
