-- +goose Up
-- drop index "webauthns_aaguid_key" from table: "webauthns"
DROP INDEX "webauthns_aaguid_key";

-- +goose Down
-- reverse: drop index "webauthns_aaguid_key" from table: "webauthns"
CREATE UNIQUE INDEX "webauthns_aaguid_key" ON "webauthns" ("aaguid");
