-- +goose Up
-- create "invite_groups" table
CREATE TABLE "invite_groups" ("invite_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "group_id"), CONSTRAINT "invite_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "invite_groups_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "invite_groups" table
DROP TABLE "invite_groups";
