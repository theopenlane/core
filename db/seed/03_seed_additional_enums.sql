-- +goose Up
-- Add additional seed data for the custom_type_enums table
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
        '01JF7YD9Z6WQ4TK3J8M2R1HCPFV',
        'task',
        'kind',
        'Gap',
        'Tasks for capturing issues or deficiencies discovered during a gap assessment and tracking the remediation work needed to close them',
        '#F97316',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W1V7SJW6Y8H0M6PY1S5ZN',
        'procedure',
        'kind',
        'Compliance',
        'Procedures that describe how the organization plans, executes, and documents activities to meet legal, regulatory, and contractual obligations (e.g., SOC 2, ISO 27001, GDPR).',
        '#84CC16',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W1V7X8F3V4K2TQJ9R6WCG',
        'procedure',
        'kind',
        'Operational',
        'Procedures for managing day-to-day operations, including change management, configuration management, deployment workflows, and third-party operational processes.',
        '#F97316',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W1V81G2E9Q7B4M1Z8CKPF',
        'procedure',
        'kind',
        'Health and Safety',
        'Procedures for ensuring a safe working environment, emergency response steps, and other health and safety activities that support employee well-being.',
        '#64748B',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W1V84B9T6D3P2R7V5MQHY',
        'procedure',
        'kind',
        'Security',
        'Procedures that define step-by-step security activities such as access provisioning and deprovisioning, vulnerability management, patching, key rotation, and system hardening.',
        '#2CCBAB',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W1V86Q4E1J5Z8K3C7TNWX',
        'procedure',
        'kind',
        'Privacy',
        'Procedures describing how personal and customer data is handled in practice, including data subject requests, data retention and deletion workflows, and privacy-impact operations.',
        '#0EA5E9',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W1V88P6M2R4V1J9Y0BDQS',
        'procedure',
        'kind',
        'HR / Personnel',
        'Procedures for employee lifecycle management such as hiring, onboarding, access setup, role changes, and offboarding with a focus on security and compliance requirements.',
        '#10B981',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W1V8BP3K7T9F2D6W4HNYM',
        'procedure',
        'kind',
        'Incident Response',
        'Procedures that define how incidents are detected, triaged, contained, remediated, and reviewed, including communication runbooks and severity classification steps.',
        '#F59E0B',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W1V8DB7Z1C5Q3T8P6KRGX',
        'procedure',
        'kind',
        'Business Continuity',
        'Procedures outlining how to maintain or restore critical operations during disasters or major outages, including disaster recovery and system restoration runbooks.',
        '#EF4444',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
     (
        '01JF7W8Z7Z8JH3X1M9Q2D5S4VT',
        'procedure',
        'kind',
        'Access Control',
        'Procedures defining how access is provisioned, modified, reviewed, and revoked for systems, applications, and data.',
        '#6366F1',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W8Z80K4F7P2R6M1C9NYBE',
        'procedure',
        'kind',
        'Change Management',
        'Procedures outlining how changes are proposed, reviewed, approved, implemented, and verified across infrastructure and applications.',
        '#A855F7',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W8Z82T1Q9H4K3V6Z7LPMD',
        'procedure',
        'kind',
        'Configuration Management',
        'Procedures describing how system configurations are defined, baseline standards applied, monitored, and maintained.',
        '#14B8A6',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W8Z84R2M8C5X1T9B3QFGK',
        'procedure',
        'kind',
        'Vendor Management',
        'Procedures defining how vendors are assessed, approved, monitored, and reviewed for security, reliability, and compliance risks.',
        '#EC4899',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W8Z86W3Z1D7Q4K2F5TNVP',
        'procedure',
        'kind',
        'Asset Management',
        'Procedures for tracking, classifying, maintaining, and reviewing the organization’s information assets.',
        '#4ADE80',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W8Z88P9C6R3M2T1X4HDSQ',
        'procedure',
        'kind',
        'Backup & Recovery',
        'Procedures detailing how backups are performed, verified, stored, and restored to ensure data resilience and business continuity.',
        '#F43F5E',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W8Z8BP4J2F6W9N3K7CRMH',
        'procedure',
        'kind',
        'Monitoring & Logging',
        'Procedures that define how logs, alerts, and monitoring data are collected, reviewed, triaged, and escalated.',
        '#0EA5A0',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF7W8Z8DQ7T5Z2C4R1V8NHWK',
        'procedure',
        'kind',
        'Vulnerability Management',
        'Procedures covering vulnerability scanning, assessment, prioritization, remediation, and validation workflows.',
        '#FACC15',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    );


-- +goose Down
-- Remove the seed data for the custom_type_enums table
DELETE FROM custom_type_enums
WHERE id IN (
    '01JF7YD9Z6WQ4TK3J8M2R1HCPFV', -- Gap (task)
    '01JF7W1V7SJW6Y8H0M6PY1S5ZN', -- Compliance (procedure)
    '01JF7W1V7X8F3V4K2TQJ9R6WCG', -- Operational (procedure)
    '01JF7W1V81G2E9Q7B4M1Z8CKPF', -- Health and Safety (procedure)
    '01JF7W1V84B9T6D3P2R7V5MQHY', -- Security (procedure)
    '01JF7W1V86Q4E1J5Z8K3C7TNWX', -- Privacy (procedure)
    '01JF7W1V88P6M2R4V1J9Y0BDQS', -- HR / Personnel (procedure)
    '01JF7W1V8BP3K7T9F2D6W4HNYM', -- Incident Response (procedure)
    '01JF7W1V8DB7Z1C5Q3T8P6KRGX'  -- Business Continuity (procedure)
    '01JF7W8Z7Z8JH3X1M9Q2D5S4VT', -- Access Control (procedure)
    '01JF7W8Z80K4F7P2R6M1C9NYBE', -- Change Management (procedure)
    '01JF7W8Z82T1Q9H4K3V6Z7LPMD', -- Configuration Management (procedure)
    '01JF7W8Z84R2M8C5X1T9B3QFGK', -- Vendor Management (procedure)
    '01JF7W8Z86W3Z1D7Q4K2F5TNVP', -- Asset Management (procedure)
    '01JF7W8Z88P9C6R3M2T1X4HDSQ', -- Backup & Recovery (procedure)
    '01JF7W8Z8BP4J2F6W9N3K7CRMH', -- Monitoring & Logging (procedure)
    '01JF7W8Z8DQ7T5Z2C4R1V8NHWK'  -- Vulnerability Management
);
