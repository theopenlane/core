# SCF - BCD-04 - Contingency Plan Testing & Exercises
Mechanisms exist to conduct tests and/or exercises to evaluate the contingency plan's effectiveness and the organization's readiness to execute the plan.
## Mapped framework controls
### ISO 27002
- [A.5.29](../iso27002/a-5.md#a529)
- [A.5.30](../iso27002/a-5.md#a530)

### NIST 800-53
- [CP-4](../nist80053/cp-4.md)

### SOC 2
- [A1.3-POF1](../soc2/a13-pof1.md)
- [A1.3-POF2](../soc2/a13-pof2.md)
- [A1.3](../soc2/a13.md)
- [CC7.5-POF6](../soc2/cc75-pof6.md)
- [CC7.5](../soc2/cc75.md)

## Evidence request list
E-BCM-06
E-BCM-07

## Control questions
Does the organization conduct tests and/or exercises to evaluate the contingency plan's effectiveness and the organization's readiness to execute the plan?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to conduct tests and/ or exercises to evaluate the contingency plan's effectiveness and the organization's readiness to execute the plan.

### Performed internally
Business Continuity & Disaster Recovery (BCD) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- IT personnel work with business stakeholders to identify business-critical systems and services, including internal teams and third-party service providers.
- IT personnel develop limited Disaster Recovery Plans (DRP) to recover business-critical systems and services.
- Business stakeholders develop limited Business Continuity Plans (BCPs) to ensure business-critical functions are sustainable both during and after an incident within Recovery Time Objectives (RTOs).
- Backups are performed ad-hoc and focus on business-critical systems.
- Limited technologies exist to support near real-time network infrastructure failover (e.g., redundant ISPs, redundant power, etc.).

### Planned and tracked
Business Continuity & Disaster Recovery (BCD) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Business Continuity / Disaster Recovery (BC/DR) management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for BC/DR management.
- BC/DR roles are formally assigned as an additional duty to existing IT/cybersecurity personnel.
- Recovery Time Objectives (RTOs) identify business-critical systems and services, which are given priority of service in alternate processing and storage sites.
- IT personnel develop Disaster Recovery Plans (DRP) to recover business-critical systems and services.
- Data/process owners conduct a Business Impact Analysis (BIA) at least annually, or after any major technology or process change, to identify assets critical to the business in need of protection, as well as single points of failure.
- IT/cybersecurity personnel work with business stakeholders to develop Business Continuity Plans (BCPs) to ensure business functions are sustainable both during and after an incident within Recovery Time Objectives (RTOs).
- IT personnel use a backup methodology (e.g., grandfather, father & s on rotation) to store backups in a secondary location, separate from the primary storage site.
- IT personnel configure business-critical systems to transfer backup data to the alternate storage site at a rate that is capable of meeting Recovery Time Objectives (RTOs) and Recovery Point Objectives (RPOs).
-  on at least an annual basis, IT/cybersecurity personnel conduct tabletop exercises to validate disaster recovery and contingency plans, in conjunction with stakeholders and any required vendors.

### Well defined
Business Continuity & Disaster Recovery (BCD) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- A formal Business Continuity & Disaster Recovery (BC/DR) program exists with defined roles and responsibilities to restore functionality in the event of a catastrophe, emergency, or significant disruptive incident that is handled in accordance with a Continuity of Operations Plan (COOP).
- BC/DR personnel work with business stakeholders to identify business-critical systems, services, internal teams and third-party service providers.
- Specific criteria are defined to initiate BC/DR activities that facilitate business continuity operations capable of meeting applicable RTOs and/or RPOs.- Application/system/process owners conduct a Business Impact Analysis (BIA) at least annually, or after any major technology or process change, to identify assets critical to the business in need of protection, as well as single points of failure.
- Recovery Time Objectives (RTOs) are defined for business-critical systems and services.
- Recovery Point Objectives (RPOs) are defined and technologies exist to conduct transaction-level recovery, in accordance with RPOs.

- Controls are assigned to sensitive/regulated assets to comply with specific BC/DR requirements to facilitate recovery operations in accordance with RTOs and RPOs.
- IT personnel work with business stakeholders to develop Disaster Recovery Plans (DRP) to recover business-critical systems and services within RPOs.
- Business stakeholders work with IT personnel to develop Business Continuity Plans (BCPs) to ensure business functions are sustainable both during and after an incident within RTOs.
- The data backup function is formally assigned with defined roles and responsibilities.
- BC/DR personnel have pre-established methods to communicate the status of recovery activities and progress in restoring operational capabilities to designated internal and external stakeholders.
- The integrity of backups and other restoration assets are verified prior to using them for restoration.

### Quantitatively controlled
Business Continuity & Disaster Recovery (BCD) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to conduct tests and/ or exercises to evaluate the contingency plan's effectiveness and the organization's readiness to execute the plan.
