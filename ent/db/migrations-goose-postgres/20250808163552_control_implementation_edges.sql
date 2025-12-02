-- +goose Up
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "evidence_control_implementations" character varying NULL, ADD COLUMN "internal_policy_control_implementations" character varying NULL, ADD CONSTRAINT "control_implementations_evidences_control_implementations" FOREIGN KEY ("evidence_control_implementations") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "control_implementations_intern_78a7d74302db6f99776c0594111f170b" FOREIGN KEY ("internal_policy_control_implementations") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "control_implementation_tasks" table
CREATE TABLE "control_implementation_tasks" ("control_implementation_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "task_id"), CONSTRAINT "control_implementation_tasks_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_implementation_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "control_implementation_tasks" table
DROP TABLE "control_implementation_tasks";
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP CONSTRAINT "control_implementations_intern_78a7d74302db6f99776c0594111f170b", DROP CONSTRAINT "control_implementations_evidences_control_implementations", DROP COLUMN "internal_policy_control_implementations", DROP COLUMN "evidence_control_implementations";
