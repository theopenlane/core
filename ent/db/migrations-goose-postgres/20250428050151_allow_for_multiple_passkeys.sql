-- +goose Up
-- modify "webauthns" table
ALTER TABLE "webauthns" DROP CONSTRAINT "webauthns_users_webauthn", ADD CONSTRAINT "webauthns_users_webauthns" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- reverse: modify "webauthns" table
ALTER TABLE "webauthns" DROP CONSTRAINT "webauthns_users_webauthns", ADD CONSTRAINT "webauthns_users_webauthn" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
