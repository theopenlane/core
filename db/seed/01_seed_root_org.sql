-- +goose Up
-- Add the seed data for the organizations table
INSERT INTO
    organizations (
        id,
        name,
        display_name,
        description,
        created_at,
        updated_at,
        created_by,
        updated_by
    )
VALUES (
        '01101101011010010111010001100010',
        'openlane system',
        'openlane',
        'the openlane system organization',
        '1970-01-01 00:00:00',
        '1970-01-01 00:00:00',
        'system',
        'system'
    ) ON CONFLICT DO NOTHING;

-- Add the seed data for the organizations settings table
INSERT INTO
    organization_settings (
        id,
        organization_id,
        created_at,
        updated_at,
        created_by,
        updated_by
    )
VALUES (
        '01101101011001010110111101110111',
        '01101101011010010111010001100010',
        '1970-01-01 00:00:00',
        '1970-01-01 00:00:00',
        'system',
        'system'
    ) ON CONFLICT DO NOTHING;

-- +goose Down
-- Remove the seed data for the organizations settings table
DELETE FROM organization_settings
WHERE
    id = '01101101011001010110111101110111';

-- Remove the seed data for the organizations table
DELETE FROM organizations
WHERE
    id = '01101101011010010111010001100010';