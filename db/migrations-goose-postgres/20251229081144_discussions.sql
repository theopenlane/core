-- +goose Up
-- modify "discussions" table
ALTER TABLE "discussions" ALTER COLUMN "external_id" DROP NOT NULL;
-- drop index "workflowproposal_workflow_object_ref_id_domain_key" from table: "workflow_proposals"
DROP INDEX "workflowproposal_workflow_object_ref_id_domain_key";
-- create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
CREATE UNIQUE INDEX "workflowproposal_workflow_object_ref_id_domain_key" ON "workflow_proposals" ("workflow_object_ref_id", "domain_key") WHERE ((state)::text = ANY ((ARRAY['DRAFT'::character varying, 'SUBMITTED'::character varying])::text[]));
-- modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_discussions_comments", DROP COLUMN "discussion_comments", ADD CONSTRAINT "notes_discussions_comments" FOREIGN KEY ("discussion_id") REFERENCES "discussions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_discussions_comments", ADD COLUMN "discussion_comments" character varying NULL, ADD CONSTRAINT "notes_discussions_comments" FOREIGN KEY ("discussion_comments") REFERENCES "discussions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
DROP INDEX "workflowproposal_workflow_object_ref_id_domain_key";
-- reverse: drop index "workflowproposal_workflow_object_ref_id_domain_key" from table: "workflow_proposals"
CREATE UNIQUE INDEX "workflowproposal_workflow_object_ref_id_domain_key" ON "workflow_proposals" ("workflow_object_ref_id", "domain_key") WHERE ((state)::text = ANY (ARRAY[('DRAFT'::character varying)::text, ('SUBMITTED'::character varying)::text]));
-- reverse: modify "discussions" table
ALTER TABLE "discussions" ALTER COLUMN "external_id" SET NOT NULL;
