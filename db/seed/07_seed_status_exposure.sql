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
    -- Vulnerability statuses
    ('01HVK6Z8YJ7M8X3T3M7W8X1A01', 'vulnerability', 'status', 'Open', 'Newly identified and not yet reviewed.', '#6B7280', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1A02', 'vulnerability', 'status', 'Triaged', 'Reviewed and prioritized for action.', '#3B82F6', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1A03', 'vulnerability', 'status', 'In Progress', 'Actively being worked on.', '#F59E0B', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1A04', 'vulnerability', 'status', 'Remediated', 'Fix has been implemented but not yet verified.', '#10B981', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1A05', 'vulnerability', 'status', 'Resolved', 'Issue has been verified as fixed.', '#22C55E', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1A06', 'vulnerability', 'status', 'Closed', 'No further action required.', '#64748B', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1A07', 'vulnerability', 'status', 'Archived', 'Deprecated or no longer relevant.', '#94A3B8', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1A08', 'vulnerability', 'status', 'False Positive', 'Reviewed and determined not to be a valid issue.', '#F43F5E', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1A09', 'vulnerability', 'status', 'Duplicate', 'Duplicate of another tracked issue.', '#8B5CF6', true, NOW(), NOW(), 'system', 'system'),

    -- Finding statuses
    ('01HVK6Z8YJ7M8X3T3M7W8X1B01', 'finding', 'status', 'Open', 'Newly identified and not yet reviewed.', '#6B7280', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1B02', 'finding', 'status', 'Triaged', 'Reviewed and prioritized for action.', '#3B82F6', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1B03', 'finding', 'status', 'In Progress', 'Actively being worked on.', '#F59E0B', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1B04', 'finding', 'status', 'Remediated', 'Fix has been implemented but not yet verified.', '#10B981', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1B05', 'finding', 'status', 'Resolved', 'Issue has been verified as fixed.', '#22C55E', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1B06', 'finding', 'status', 'Closed', 'No further action required.', '#64748B', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1B07', 'finding', 'status', 'Archived', 'Deprecated or no longer relevant.', '#94A3B8', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1B08', 'finding', 'status', 'False Positive', 'Reviewed and determined not to be a valid issue.', '#F43F5E', true, NOW(), NOW(), 'system', 'system'),
    ('01HVK6Z8YJ7M8X3T3M7W8X1B09', 'finding', 'status', 'Duplicate', 'Duplicate of another tracked issue.', '#8B5CF6', true, NOW(), NOW(), 'system', 'system')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM custom_type_enums
WHERE id IN (
    '01HVK6Z8YJ7M8X3T3M7W8X1A01',
    '01HVK6Z8YJ7M8X3T3M7W8X1A02',
    '01HVK6Z8YJ7M8X3T3M7W8X1A03',
    '01HVK6Z8YJ7M8X3T3M7W8X1A04',
    '01HVK6Z8YJ7M8X3T3M7W8X1A05',
    '01HVK6Z8YJ7M8X3T3M7W8X1A06',
    '01HVK6Z8YJ7M8X3T3M7W8X1A07',
    '01HVK6Z8YJ7M8X3T3M7W8X1A08',
    '01HVK6Z8YJ7M8X3T3M7W8X1A09',
    '01HVK6Z8YJ7M8X3T3M7W8X1B01',
    '01HVK6Z8YJ7M8X3T3M7W8X1B02',
    '01HVK6Z8YJ7M8X3T3M7W8X1B03',
    '01HVK6Z8YJ7M8X3T3M7W8X1B04',
    '01HVK6Z8YJ7M8X3T3M7W8X1B05',
    '01HVK6Z8YJ7M8X3T3M7W8X1B06',
    '01HVK6Z8YJ7M8X3T3M7W8X1B07',
    '01HVK6Z8YJ7M8X3T3M7W8X1B08',
    '01HVK6Z8YJ7M8X3T3M7W8X1B09'
);