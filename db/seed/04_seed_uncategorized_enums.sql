-- +goose Up
-- insert new uncategorized enums and remove old "Other" enums
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
    (
        'ADJ4H1WXTEBFDR3RHFJGTQCWTJ',
        'task',
        'kind',
        'Uncategorized',
        'Tasks that do not fit any defined type or where a more specific task type is not required.',
        '#9CA3AF',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        '0YZD3BF3BVSK218M9B5NZKG3K7',
        'control',
        'kind',
        'Uncategorized',
        'Controls that do not fit any defined type or where a more specific control type is not required.',
        '#9CA3AF',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        '43YVERHXG5AYMHPMZZEPMXX8AM',
        'internal_policy',
        'kind',
        'Uncategorized',
        'Policies that do not fit any defined type or where a more specific policy type is not required.',
        '#9CA3AF',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        'QM9VVMBVTK3NMGVJRY4DD8WT68',
        'program',
        'kind',
        'Uncategorized',
        'Programs that do not fit any defined type or where a more specific program type is not required.',
        '#9CA3AF',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        '9FCJXH7CSYPDE81T4G91D3XRTN',
        'risk',
        'kind',
        'Uncategorized',
        'Risks that do not fit any defined type or where a more specific risk type is not required.',
        '#9CA3AF',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        'KWE9AM9KC43HC3KNRQ74TWMFJ3',
        'risk',
        'category',
        'Uncategorized',
        'Risks that do not fit any defined category or where a more specific category is not required.',
        '#9CA3AF',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        '97V9YB9VJNKXPH03HGZDPS3MAC',
        'procedure',
        'kind',
        'Uncategorized',
        'Procedures whose kind does not fit any defined procedure kind or where a more specific kind is not required.',
        '#9CA3AF',
        true, NOW(), NOW(), 'system', 'system'
    )
ON CONFLICT DO NOTHING;

-- 2) Remove all existing "Other" enums (any object_type/field)
DELETE FROM custom_type_enums
WHERE name = 'Other';


-- +goose Down
-- Restore all 'Other' enums and remove all 'Uncategorized' ones

-- 1) Remove the new Uncategorized rows
DELETE FROM custom_type_enums
WHERE id IN (
    'ADJ4H1WXTEBFDR3RHFJGTQCWTJ',
    '0YZD3BF3BVSK218M9B5NZKG3K7',
    '43YVERHXG5AYMHPMZZEPMXX8AM',
    'QM9VVMBVTK3NMGVJRY4DD8WT68',
    '9FCJXH7CSYPDE81T4G91D3XRTN',
    'KWE9AM9KC43HC3KNRQ74TWMFJ3',
    'E3NVV62HSKHW59RX7AD3JW7RFN',
    '97V9YB9VJNKXPH03HGZDPS3MAC'
);

-- 2) Reinsert all original "Other" rows from seeds
INSERT INTO
    custom_type_enums (
        id,
        object_type,
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
    --- This row previously the same ULID as "Control Implementation" in older seed file
    (
        '01JF4KNDKC2S4Y5V1XW1K8HTWM',
        'task',
        'Other',
        'Fallback category for other tasks that do not have a defined type',
        '#EF4444',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        '01JF4P2A1Y5R6C8V4NQT2X7HML',
        'program',
        'Other',
        'Custom or ad-hoc programs not tied to a specific framework or risk assessment, used for internal or experimental initiatives',
        '#6366F1',
        true, NOW(), NOW(), 'system', 'system'
    )
ON CONFLICT DO NOTHING;
