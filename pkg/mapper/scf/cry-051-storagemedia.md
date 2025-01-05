# SCF - CRY-05.1 - Storage Media
Cryptographic mechanisms exist to protect the confidentiality and integrity of sensitive/regulated data residing on storage media.
## Mapped framework controls
### SOC 2
- [CC6.7-POF3](../soc2/cc67-pof3.md)

## Evidence request list


## Control questions
Are cryptographic mechanisms utilized to protect the confidentiality and integrity of sensitive/regulated data residing on storage media?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to Cryptographic protect the confidentiality and integrity of sensitive/regulated data residing on storage media.

### Performed internally
Cryptographic Protections (CRY) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Data classification and handling criteria govern requirements to encrypt sensitive/regulated data during transmission and in storage.
- Network communications containing sensitive/regulated data are protected using a cryptographic mechanism to prevent unauthorized disclosure of information while in transit (e.g., SSH, TLS, VPN, etc.).
- Wireless access is protected via secure authentication and encryption.
- Mobile devices containing sensitive/regulated data use a cryptographic mechanism to prevent the unauthorized disclosure of information at rest (e.g., whole drive encryption).

### Planned and tracked
Cryptographic Protections (CRY) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Cryptographic management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for cryptographic management.
- Data classification and handling criteria govern requirements to encrypt sensitive/regulated data during transmission and in storage.
- Decentralized technologies implement cryptographic mechanisms on endpoints to control how sensitive/regulated data is encrypted during transmission and in storage.
- Systems, applications and services that store, process or transmit sensitive/regulated data use cryptographic mechanisms to prevent unauthorized disclosure of information as an alternate to physical safeguards.

### Well defined
Cryptographic Protections (CRY) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- The Chief Information Security Officer (CISO), or similar function with technical competence to address cybersecurity concerns, analyzes the organization's business strategy to determine prioritized and authoritative guidance for cryptographic protections.
- The CISO, or similar function, develops a security-focused Concept of Operations (CONOPS) that documents management, operational and technical measures for cryptographic protections.
- A Governance, Risk & Compliance (GRC) function, or similar function, provides governance oversight for the implementation of applicable statutory, regulatory and contractual cybersecurity & data privacy controls to protect the confidentiality, integrity, availability and safety of the organization's applications, systems, services and data using cryptographic protections.
- A steering committee is formally established to provide executive oversight of the cybersecurity & data privacy program, including cryptographic protections.
- Data classification and handling criteria govern requirements to encrypt sensitive/regulated data during transmission and in storage.
- Centrally-managed technologies implement cryptographic mechanisms on endpoints to control how sensitive/regulated data is encrypted during transmission and in storage.
- Certificate management is centrally-managed and the use of certificates is monitored.
- Cryptographic keys are proactively managed to protect the Confidentiality, Integrity and Availability (CIA) of cryptographic capabilities.
- Systems, applications and services that store, process or transmit sensitive/regulated data use cryptographic mechanisms to prevent unauthorized disclosure of information as an alternate to physical safeguards.
- An IT infrastructure team, or similar function:
o	Implements Public Key Infrastructure (PKI) key management controls to protect the confidentiality, integrity and availability of keys.
o	Implements and maintains an internal PKI infrastructure or obtains PKI services from a reputable PKI service provider.
- IT/cybersecurity personnel perform an annual review of deployed cryptographic cipher suites and protocols to identify and replace weak cryptographic cipher suites and protocols.

### Quantitatively controlled
Cryptographic Protections (CRY) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to Cryptographic protect the confidentiality and integrity of sensitive/regulated data residing on storage media.
