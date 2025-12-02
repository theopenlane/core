-- +goose Up
-- Add the seed data for the custom_type_enums table
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
    (
        '01JF4KNDK9F7R1JY1JX4Z5T0NZ',
        'task',
        'Evidence',
        'Tasks created to collect or verify compliance evidence, such as screenshots, reports, or configurations supporting a control',
        '#2CCBAB',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4KNDKAM8RQ5V1N1D5J8W8X',
        'task',
        'Risk Review',
        'Tasks used to assess identified risks, document mitigation steps, and track review outcomes or risk acceptance decisions',
        '#0EA5E9',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4KNDKB4F6TF1EJX8MP1PW6',
        'task',
        'Policy Review',
        'Tasks for reviewing and approving policies or procedures to ensure they remain accurate, effective, and up to date',
        '#10B981',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4KNDKC2S4Y5V1XW1K8HTWM',
        'task',
        'Control Implementation',
        'Tasks that track implementation activities required to satisfy a control, such as configuring security settings or deploying monitoring tools',
        '#F59E0B',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
      (
        '01JF4KNDKC2S4Y5V1XW1K8HTWM',
        'task',
        'Other',
        'Fallback category for other tasks that do not have a defined type',
        '#EF4444',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    )
    ON CONFLICT DO NOTHING;

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
    (
        '01JF4M3B7T6H3E9V6KCN1FXG8E',
        'control',
        'Preventative',
        'Controls designed to proactively stop unwanted or unauthorized actions before they occur—for example, access restrictions or input validation',
        '#8B5CF6',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4M3B7V8M2R3B8EYK9XK7HC',
        'control',
        'Detective',
        'Controls that identify and alert on events or conditions after they happen, such as monitoring, logging, or anomaly detection',
        '#EC4899',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4M3B7X1Q6Q5T9HYJ7QAH4C',
        'control',
        'Corrective',
        'Controls intended to remediate or recover from identified issues, including incident response and system restoration activities',
        '#14B8A6',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4M3B7Y5M4Z8T3CQN7RSP8V',
        'control',
        'Deterrent',
        'Controls that discourage or reduce the likelihood of undesired behavior through policy, awareness, or visible enforcement measures',
        '#6366F1',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ) ON CONFLICT DO NOTHING;

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
    (
        '01JF4N1R3Z5J8Q0F6YVK9W2RTS',
        'internal_policy',
        'Compliance',
        'Policies defining how the organization meets legal, regulatory, and contractual obligations such as SOC 2, ISO 27001, and GDPR',
        '#84CC16',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4N1R40V1R6E3X9DT1C2WMP',
        'internal_policy',
        'Operational',
        'Policies outlining standards for managing day-to-day operations, change management, and third-party relationships',
        '#F97316',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4N1R42K8M7A9T2ZB4J8NXY',
        'internal_policy',
        'Health and Safety',
        'Policies for ensuring a safe workplace, emergency preparedness, and employee well-being',
        '#64748B',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4N1R43C6D2Y7V5QJ1B9KRL',
        'internal_policy',
        'Security',
        'Policies governing data protection, access control, encryption, and system hardening',
        '#2CCBAB',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4N1R44P5H4C8E2RZ7N1LKM',
        'internal_policy',
        'Privacy',
        'Policies describing how personal and customer data are collected, stored, shared, and deleted in accordance with privacy laws',
        '#0EA5E9',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4N1R45Q9J6T1U3VY5B7XDP',
        'internal_policy',
        'HR / Personnel',
        'Policies that define employee lifecycle management, including hiring, onboarding, and termination with a focus on security awareness',
        '#10B981',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4N1R46S8P3M9R2QW1V6TLY',
        'internal_policy',
        'Incident Response',
        'Policies defining the process for identifying, reporting, and responding to security incidents or operational disruptions',
        '#F59E0B',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4N1R47X2B9L4T5M8Q7J2FD',
        'internal_policy',
        'Business Continuity',
        'Policies outlining recovery and continuity strategies to maintain operations during disasters or major outages',
        '#EF4444',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ) ON CONFLICT DO NOTHING;

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
    (
        '01JF4P2A1T9Z7M8E2VQH4K5YJB',
        'program',
        'Framework',
        'Programs aligned to external standards or frameworks such as SOC 2, ISO 27001, NIST 800-53, or HIPAA, defining required controls and objectives',
        '#8B5CF6',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4P2A1V4H3R9N5BXD1Q6FZT',
        'program',
        'Gap Analysis',
        'Programs created to identify and document gaps between current security or compliance posture and target framework requirements',
        '#EC4899',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4P2A1X8N2L7K3DWF9C4QSG',
        'program',
        'Risk Assessment',
        'Programs that evaluate and prioritize organizational risks, establishing likelihood, impact, and mitigation strategies',
        '#14B8A6',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4P2A1Y5R6C8V4NQT2X7HML',
        'program',
        'Other',
        'Custom or ad-hoc programs not tied to a specific framework or risk assessment, used for internal or experimental initiatives',
        '#6366F1',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ) ON CONFLICT DO NOTHING;

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
    (
        '01JF4Q6M2T7Z4R1P8NXC3B9FDJ',
        'risk',
        'Strategic',
        'Risks that impact long-term goals, mission alignment, or organizational direction—such as market changes, competition, or poor strategic decisions',
        '#84CC16',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4Q6M2V9D5T3E6QYR7W1LKP',
        'risk',
        'Operational',
        'Risks arising from day-to-day business processes, systems, and people—including process failures, human error, or system outages',
        '#F97316',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4Q6M2X1K9J2H5RZ8V3YCGN',
        'risk',
        'Financial',
        'Risks related to financial stability, including budgeting errors, cost overruns, fraud, or economic volatility',
        '#64748B',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4Q6M2Y3P4L8N7BHT2S5QWD',
        'risk',
        'External',
        'Risks originating outside the organization, such as supply chain disruption, natural disasters, geopolitical events, or regulatory changes',
        '#2CCBAB',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
     (
        '01JF4QZZ8K3N6P2T7VY9C1R5XM',
        'risk',
        'Compliance',
        'Risks of failing to comply with laws, regulations, contractual requirements, or standards (e.g., fines, penalties, audit findings, or enforced remediation)',
        '#0EA5E9',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4QZZ9D4S7M3X6QY2B8H5TN',
        'risk',
        'Reputational',
        'Risks to brand trust and stakeholder confidence arising from incidents, outages, privacy breaches, or negative publicity',
        '#10B981',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ) ON CONFLICT DO NOTHING;

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
        '01JF4S5K9A2M7P6Q8VY1B3R4DX',
        'risk',
        'category',
        'Human Resources',
        'Risks related to personnel management, staffing, training, insider threats, or turnover',
        '#F59E0B',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4S5K9B4Q8R3T6XW2N9V7HJ',
        'risk',
        'category',
        'Operations',
        'Risks arising from process inefficiencies, service disruptions, or vendor dependencies',
        '#EF4444',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4S5K9C9T2E5M4QZ7P1Y8WK',
        'risk',
        'category',
        'Information Technology',
        'Risks linked to systems, networks, applications, or infrastructure failures',
        '#8B5CF6',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4S5K9D5L3B8R9VQ6H2C4YP',
        'risk',
        'category',
        'Legal & Compliance',
        'Risks from noncompliance with laws, regulations, or contractual obligations',
        '#EC4899',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4S5K9E7N1W2X3R8T6C9BHF',
        'risk',
        'category',
        'Finance',
        'Risks tied to budgeting, cash flow, or accounting accuracy',
        '#14B8A6',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4S5K9F2Q6H4V5N3X8Y7ZJR',
        'risk',
        'category',
        'Physical Security',
        'Risks from unauthorized physical access, theft, or facility damage',
        '#6366F1',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4S5K9G3V9B2M6R4P7T8HQC',
        'risk',
        'category',
        'Supply Chain / Third Party',
        'Risks stemming from vendor performance, dependency, or supply chain disruption',
        '#84CC16',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ),
    (
        '01JF4S5K9H1L5W7T9Q2E3Y4ZKD',
        'risk',
        'category',
        'Strategic / Governance',
        'Risks connected to leadership decisions, organizational structure, or misalignment with business goals',
        '#F97316',
        true,
        NOW(),
        NOW(),
        'system',
        'system'
    ) ON CONFLICT DO NOTHING;

-- +goose Down
-- Remove the seed data for the custom_type_enums table
DELETE FROM custom_type_enums
WHERE id IN (
  '01JF4KNDK9F7R1JY1JX4Z5T0NZ', -- Evidence
  '01JF4KNDKAM8RQ5V1N1D5J8W8X', -- Risk Review
  '01JF4KNDKB4F6TF1EJX8MP1PW6', -- Policy Review
  '01JF4KNDKC2S4Y5V1XW1K8HTWM'  -- Control Implementation
);

DELETE FROM custom_type_enums
WHERE id IN (
  '01JF4M3B7T6H3E9V6KCN1FXG8E', -- Preventative
  '01JF4M3B7V8M2R3B8EYK9XK7HC', -- Detective
  '01JF4M3B7X1Q6Q5T9HYJ7QAH4C', -- Corrective
  '01JF4M3B7Y5M4Z8T3CQN7RSP8V'  -- Deterrent
);

DELETE FROM custom_type_enums
WHERE id IN (
  '01JF4N1R3Z5J8Q0F6YVK9W2RTS', -- Compliance
  '01JF4N1R40V1R6E3X9DT1C2WMP', -- Operational
  '01JF4N1R42K8M7A9T2ZB4J8NXY', -- Health and Safety
  '01JF4N1R43C6D2Y7V5QJ1B9KRL', -- Security
  '01JF4N1R44P5H4C8E2RZ7N1LKM', -- Privacy
  '01JF4N1R45Q9J6T1U3VY5B7XDP', -- HR / Personnel
  '01JF4N1R46S8P3M9R2QW1V6TLY', -- Incident Response
  '01JF4N1R47X2B9L4T5M8Q7J2FD'  -- Business Continuity
);

DELETE FROM custom_type_enums
WHERE id IN (
  '01JF4P2A1T9Z7M8E2VQH4K5YJB', -- Framework
  '01JF4P2A1V4H3R9N5BXD1Q6FZT', -- Gap Analysis
  '01JF4P2A1X8N2L7K3DWF9C4QSG', -- Risk Assessment
  '01JF4P2A1Y5R6C8V4NQT2X7HML'  -- Other
);

DELETE FROM custom_type_enums
WHERE id IN (
  '01JF4Q6M2T7Z4R1P8NXC3B9FDJ', -- Strategic
  '01JF4Q6M2V9D5T3E6QYR7W1LKP', -- Operational
  '01JF4Q6M2X1K9J2H5RZ8V3YCGN', -- Financial
  '01JF4Q6M2Y3P4L8N7BHT2S5QWD'  -- External
  '01JF4QZZ8K3N6P2T7VY9C1R5XM', -- Compliance
  '01JF4QZZ9D4S7M3X6QY2B8H5TN'  -- Reputational
);

DELETE FROM custom_type_enums
WHERE id IN (
  '01JF4S5K9A2M7P6Q8VY1B3R4DX', -- Human Resources
  '01JF4S5K9B4Q8R3T6XW2N9V7HJ', -- Operations
  '01JF4S5K9C9T2E5M4QZ7P1Y8WK', -- Information Technology
  '01JF4S5K9D5L3B8R9VQ6H2C4YP', -- Legal & Compliance
  '01JF4S5K9E7N1W2X3R8T6C9BHF', -- Finance
  '01JF4S5K9F2Q6H4V5N3X8Y7ZJR', -- Physical Security
  '01JF4S5K9G3V9B2M6R4P7T8HQC', -- Supply Chain / Third Party
  '01JF4S5K9H1L5W7T9Q2E3Y4ZKD'  -- Strategic / Governance
);
