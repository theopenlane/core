-- Modify "notes" table
ALTER TABLE "notes" ADD COLUMN "trust_center_posts" character varying NULL, ADD CONSTRAINT "notes_trust_centers_posts" FOREIGN KEY ("trust_center_posts") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
