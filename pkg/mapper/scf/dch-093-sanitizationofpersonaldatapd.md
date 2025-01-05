# SCF - DCH-09.3 - Sanitization of Personal Data (PD)
Mechanisms exist to facilitate the sanitization of Personal Data (PD).
## Mapped framework controls
### GDPR
- [Art 5.1](../gdpr/art5.md#Article-51)

### ISO 27002
- [A.8.10](../iso27002/a-8.md#a810)

### NIST 800-53
- [MP-6](../nist80053/mp-6.md)

### SOC 2
- [P4.3](../soc2/p43.md)

## Evidence request list


## Control questions
Does the organization facilitate the sanitization of Personal Data (PD)?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to facilitate the sanitization of Personal Data (PD).

### Performed internally
SP-CMM1 is N/A, since a structured process is required to facilitate the sanitization of Personal Data (PD).

### Planned and tracked
Data Classification & Handling (DCH) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Data management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for data management.
- Data protection controls are primarily administrative and preventative in nature (e.g., policies & standards) to classify, protect and dispose of systems and data, including storage media.
- A data classification process exists to identify categories of data and specific protection requirements.
- A data retention process exists and is a manual process to govern.
- Data/process owners:
o	Document where sensitive/regulated data is stored, transmitted and processed to identify data repositories and data flows.
o	Create and maintain Data Flow Diagrams (DFDs) and network diagrams.
o	Are expected to take the initiative to work with Data Protection Officers (DPOs) to ensure applicable statutory, regulatory and contractual obligations are properly addressed, including the storage, transmission and processing of sensitive/regulated data
- A manual data retention process exists.
- Content filtering blocks users from performing ad hoc file transfers through unapproved file transfer services (e.g., Box, Dropbox, Google Drive, etc.).
- Mobile Device Management (MDM) software is used to restrict and protect the data that resides on mobile devices.
- Physical controls, administrative processes and technologies focus on protecting High Value Assets (HVAs), including environments where sensitive/regulated data is stored, transmitted and processed.
- Administrative means (e.g., policies and standards) dictate:
o	Geolocation requirements for sensitive/regulated data types, including the transfer of data to third-countries or international organizations.
o	Requirements for minimizing data collection to what is necessary for business purposes.
o	Requirements for limiting the use of sensitive/regulated data in testing, training and research.

### Well defined
Data Classification & Handling (DCH) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- A Governance, Risk & Compliance (GRC) function, or similar function, assists users in making information sharing decisions to ensure data is appropriately protected, regardless of where or how it is stored, processed and/ or transmitted.
- A data classification process exists to identify categories of data and specific protection requirements.
- A data retention process exists to protect archived data in accordance with applicable statutory, regulatory and contractual obligations.
- Data/process owners:
o	Are expected to take the initiative to work with Data Protection Officers (DPOs) to ensure applicable statutory, regulatory and contractual obligations are properly addressed, including the storage, transmission and processing of sensitive/regulated data.
o	Maintain decentralized inventory logs of all sensitive/regulated media and update sensitive/regulated media inventories at least annually.
o	Create and maintain Data Flow Diagrams (DFDs) and network diagrams.
o	Document where sensitive/regulated data is stored, transmitted and processed in order to document data repositories and data flows.
- A Data Protection Impact Assessment (DPIA) is used to help ensure the protection of sensitive/regulated data processed, stored or transmitted on internal or external systems, in order to implement cybersecurity & data privacy controls in accordance with applicable statutory, regulatory and contractual obligations.
- Human Resources (HR), documents formal “rules of behavior” as an employment requirement that stipulates acceptable and unacceptable practices pertaining to sensitive/regulated data handling.
- Data Loss Prevention (DLP), or similar content filtering capabilities, blocks users from performing ad hoc file transfers through unapproved file transfer services (e.g., Box, Dropbox, Google Drive, etc.).
- Mobile Device Management (MDM) software is used to restrict and protect the data that resides on mobile devices.
- Administrative processes and technologies:
o	Identify data classification types to ensure adequate cybersecurity & data privacy controls are in place to protect organizational information and individual data privacy.
o	Identify and document the location of information on which the information resides.
o	Restrict and govern the transfer of data to third-countries or international organizations.
o	Limit the disclosure of data to authorized parties.
o	Mark media in accordance with data protection requirements so that personnel are alerted to distribution limitations, handling caveats and applicable security requirements.
o	Prohibit “rogue instances” where unapproved third parties are engaged to store, process or transmit data, including budget reviews and firewall connection authorizations.
o	Protect and control digital and non-digital media during transport outside of controlled areas using appropriate security measures.
o	Govern the use of personal devices (e.g., Bring Your Own Device (BYOD)) as part of acceptable and unacceptable behaviors.
o	Dictate requirements for minimizing data collection to what is necessary for business purposes.
o	Dictate requirements for limiting the use of sensitive/regulated data in testing, training and research.

### Quantitatively controlled
Data Classification & Handling (DCH) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to facilitate the sanitization of Personal Data (PD).
