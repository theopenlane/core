# SCF - CHG-03 - Security Impact Analysis for Changes
Mechanisms exist to analyze proposed changes for potential security impacts, prior to the implementation of the change.
## Mapped framework controls
### NIST 800-53
- [CM-4](../nist80053/cm-4.md)

### SOC 2
- [CC3.4-POF4](../soc2/cc34-pof4.md)
- [CC3.4](../soc2/cc34.md)
- [CC8.1-POF10](../soc2/cc81-pof10.md)
- [CC8.1-POF3](../soc2/cc81-pof3.md)

## Evidence request list


## Control questions
Does the organization analyze proposed changes for potential security impacts, prior to the implementation of the change?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to analyze proposed changes for potential security impacts, prior to the implementation of the change.

### Performed internally
Change Management (CHG) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- IT personnel use an informal process to:
o	Govern changes to systems, applications and services to ensure their stability, reliability and predictability.
o	Notify stakeholders about proposed changes.
- Logical Access Control (LAC) limits the ability of non-administrators from making unauthorized configuration changes to systems, applications and services.
- Requests for Change (RFC) are submitted to IT personnel.
- prior to changes being made, RFCs are informally reviewed for cybersecurity & data privacy ramifications.
- Whenever possible, IT personnel test changes to business-critical systems/services/applications on a similarly configured IT environment as that of Production, prior to widespread production release of the change.

### Planned and tracked
Change Management (CHG) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Change management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for change management.
- Changes are tracked through a centralized technology solution to submit, review, approve and assign Requests for Change (RFC).
- A Change Advisory Board (CAB), or similar function, exists to govern changes to systems, applications and services to ensure their stability, reliability and predictability.
- A CAB, or similar function, reviews RFCs for cybersecurity & data privacy ramifications.
- A CAB, or similar function, notifies stakeholders to ensure awareness of the impact of proposed changes.
- Logical Access Control (LAC) limits the ability of non-administrators from making unauthorized configuration changes to systems, applications and services.
- Cybersecurity controls are tested after a change is implemented to ensure cybersecurity controls are operating properly.

### Well defined
Change Management (CHG) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- An IT Asset Management (ITAM) function, or similar function, ensures compliance with requirements for asset management.
- ITAM leverages a Configuration Management Database (CMDB), or similar tool, as the authoritative source of IT assets.
- Logical Access Control (LAC) is governed to limit the ability of non-administrators from making configuration changes to systems, applications and services.
- A formal Change Management (CM) program ensures that no unauthorized changes are made, that all changes are documented, that services are not disrupted and that resources are used efficiently.
- The CM function has formally defined roles and associated responsibilities.
- Changes are tracked through a centralized technology solution to submit, review, approve and assign Requests for Change (RFC).
- A Change Advisory Board (CAB), or similar function:
o	Exists to govern changes to systems, applications and services to ensure their stability, reliability and predictability.
o	Reviews RFC for cybersecurity & data privacy ramifications.
o	Notifies stakeholders to ensure awareness of the impact of proposed changes.
- IT personnel use dedicated development/test/staging environments to deploy and evaluate changes, wherever technically possible.
- Up on implementing the RFC, the technician implementing a change tests to ensure anti-malware, logging and other cybersecurity & data privacy controls are still implemented and operating properly.
- Results from testing changes are documented.

### Quantitatively controlled
Change Management (CHG) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to analyze proposed changes for potential security impacts, prior to the implementation of the change.
