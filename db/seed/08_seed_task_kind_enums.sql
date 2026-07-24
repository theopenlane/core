-- +goose Up
INSERT INTO
    custom_type_enums (
        id,
        object_type,
        field,
        name,
        description,
        color,
        system_owned,
        created_at,
        updated_at,
        created_by,
        updated_by
    )
VALUES
    ('01KY3SPAPA16RPWMS5R3KJCFN9', 'task', 'kind', 'Operational', 'Tasks related to day-to-day operational activities.', '#6392FF', true, NOW(), NOW(), 'system', 'system'),
    ('01KY3SPAPA16RPWMS5R5EEQ4H6', 'task', 'kind', 'Registry', 'Tasks related to maintaining and completing entries in the registry.', '#FF842C', true, NOW(), NOW(), 'system', 'system'),
    ('01KY3SPAPA16RPWMS5R87Z2VWG', 'task', 'kind', 'Trust Center', 'Tasks related to configuring and maintaining the trust center.', '#8B5CF6', true, NOW(), NOW(), 'system', 'system')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM custom_type_enums
WHERE id IN (
    '01KY3SPAPA16RPWMS5R3KJCFN9',
    '01KY3SPAPA16RPWMS5R5EEQ4H6',
    '01KY3SPAPA16RPWMS5R87Z2VWG'
);
