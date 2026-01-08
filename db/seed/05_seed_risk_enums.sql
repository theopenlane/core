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
       (
        '9FCJXH8BSYPDE81T4G91D3XRTK',
        'risk',
        'kind',
        'Technical',
        'Risks that pertain to technical aspects such as software, hardware, or network vulnerabilities.',
        '#3B82F6',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        '9FCJXH9DSYPDE81T4G91D3XRTL',
        'risk',
        'category',
        'Security',
        'Risks related to information security, including threats to data confidentiality, integrity, and availability.',
        '#10B981',
        true, NOW(), NOW(), 'system', 'system'
    ),
    (
        '9FCJXHAESYPDE81T4G91D3XRTM',
        'risk',
        'category',
        'Infrastructure',
        'Risks associated with physical and virtual infrastructure components, such as servers, data centers, and network devices.',
        '#F59E0B',
        true, NOW(), NOW(), 'system', 'system'
    )
ON CONFLICT DO NOTHING;


-- +goose Down
DELETE FROM custom_type_enums
WHERE id IN (
    '9FCJXH8BSYPDE81T4G91D3XRTK',
    '9FCJXH9DSYPDE81T4G91D3XRTL',
    '9FCJXHAESYPDE81T4G91D3XRTM'
);
