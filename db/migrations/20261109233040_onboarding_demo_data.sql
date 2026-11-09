-- Modify "onboardings" table
ALTER TABLE "onboardings" ADD COLUMN "seed_demo_data" boolean NULL, ADD COLUMN "demo_dataset" character varying NULL DEFAULT 'soc2';
