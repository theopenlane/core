-- +goose Up
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "platform_architecture_diagrams" character varying NULL, ADD COLUMN "platform_data_flow_diagrams" character varying NULL, ADD COLUMN "platform_trust_boundary_diagrams" character varying NULL, ADD CONSTRAINT "files_platforms_architecture_diagrams" FOREIGN KEY ("platform_architecture_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_platforms_data_flow_diagrams" FOREIGN KEY ("platform_data_flow_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_platforms_trust_boundary_diagrams" FOREIGN KEY ("platform_trust_boundary_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_platforms_trust_boundary_diagrams", DROP CONSTRAINT "files_platforms_data_flow_diagrams", DROP CONSTRAINT "files_platforms_architecture_diagrams", DROP COLUMN "platform_trust_boundary_diagrams", DROP COLUMN "platform_data_flow_diagrams", DROP COLUMN "platform_architecture_diagrams";
