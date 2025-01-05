# SCF - CHG-05 - Stakeholder Notification of Changes
Mechanisms exist to ensure stakeholders are made aware of and understand the impact of proposed changes.
## Mapped framework controls
### NIST 800-53
- [CM-9](../nist80053/cm-9.md)

### SOC 2
- [CC2.2-POF13](../soc2/cc22-pof13.md)
- [CC8.1](../soc2/cc81.md)

## Evidence request list


## Control questions
Does the organization ensure stakeholders are made aware of and understand the impact of proposed changes?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to ensure stakeholders are made aware of and understand the impact of proposed changes.

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

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to ensure stakeholders are made aware of and understand the impact of proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to ensure stakeholders are made aware of and understand the impact of proposed changes.
