-- +goose Up
-- Add the seed data for the organizations table
INSERT INTO
    organizations (
        id,
        mapping_id,
        name,
        display_name,
        description,
        created_at,
        updated_at,
        created_by_id,
        updated_by_id
    )
VALUES (
        '01101101011010010111010001100010',
        '01101111011100000110010101101110',
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
        mapping_id,
        created_at,
        updated_at,
        created_by_id,
        updated_by_id
    )
VALUES (
        '01101101011001010110111101110111',
        '01101101011010010111010001100010',
        '01101100011000010110111001100101',
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