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
   ('01KGC28TKC64S0EMFJ4YT6YX1X', 'trust_center_subprocessor', 'kind', 'Email Provider', 'Provides email services for communication and marketing purposes.', '#3B82F6', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YX2A', 'trust_center_subprocessor', 'kind', 'Infrastructure Provider', 'Provides foundational IT infrastructure services such as servers, networking, and computing resources.', '#6366F1', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YX3B', 'trust_center_subprocessor', 'kind', 'Cloud Storage', 'Provides cloud-based data storage solutions.', '#10B981', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YX4C', 'trust_center_subprocessor', 'kind', 'Content Delivery Network', 'Accelerate content delivery by distributing it across global servers.', '#F59E42', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YX5D', 'trust_center_subprocessor', 'kind', 'Customer Support', 'Provides customer support and helpdesk services.', '#F43F5E', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YX6E', 'trust_center_subprocessor', 'kind', 'Payment Processor', 'Handle payment processing and financial transactions.', '#FBBF24', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YX7F', 'trust_center_subprocessor', 'kind', 'Analytics Provider', 'Provides analytics and data insights services.', '#8B5CF6', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YX8G', 'trust_center_subprocessor', 'kind', 'Error Monitoring', 'Monitor and report application errors and performance issues.', '#EC4899', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YX9H', 'trust_center_subprocessor', 'kind', 'Authentication Provider', 'Manage user authentication and identity services.', '#06B6D4', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YXI1', 'trust_center_subprocessor', 'kind', 'Customer Relationship Management', 'Provides CRM solutions for managing customer data and interactions.', '#84CC16', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YXJ2', 'trust_center_subprocessor', 'kind', 'Marketing Automation', 'Subprocessors that automate marketing campaigns and customer engagement.', '#EAB308', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YXK3', 'trust_center_subprocessor', 'kind', 'Communication Tools', 'Messaging, video, or other communication services.', '#0EA5E9', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YXL4', 'trust_center_subprocessor', 'kind', 'Payroll Provider', 'Manage payroll and employee compensation services.', '#F472B6', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YXM5', 'trust_center_subprocessor', 'kind', 'eSignature Provider', 'Provides electronic signature and document management services.', '#A3E635', true, NOW(), NOW(), 'system', 'system'),
    ('01KGC28TKC64S0EMFJ4YT6YXN6', 'trust_center_subprocessor', 'kind', 'Compliance & Privacy Management', 'Subprocessors that help manage compliance and privacy requirements.', '#F87171', true, NOW(), NOW(), 'system', 'system')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM custom_type_enums
WHERE id IN (
    '01KGC28TKC64S0EMFJ4YT6YX1X',
    '01KGC28TKC64S0EMFJ4YT6YX2A',
    '01KGC28TKC64S0EMFJ4YT6YX3B',
    '01KGC28TKC64S0EMFJ4YT6YX4C',
    '01KGC28TKC64S0EMFJ4YT6YX5D',
    '01KGC28TKC64S0EMFJ4YT6YX6E',
    '01KGC28TKC64S0EMFJ4YT6YX7F',
    '01KGC28TKC64S0EMFJ4YT6YX8G',
    '01KGC28TKC64S0EMFJ4YT6YX9H',
    '01KGC28TKC64S0EMFJ4YT6YXI1',
    '01KGC28TKC64S0EMFJ4YT6YXJ2',
    '01KGC28TKC64S0EMFJ4YT6YXK3',
    '01KGC28TKC64S0EMFJ4YT6YXL4',
    '01KGC28TKC64S0EMFJ4YT6YXM5',
    '01KGC28TKC64S0EMFJ4YT6YXN6'
);