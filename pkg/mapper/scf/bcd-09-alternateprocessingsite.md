# SCF - BCD-09 - Alternate Processing Site
Mechanisms exist to establish an alternate processing site that provides security measures equivalent to that of the primary site.
## Mapped framework controls
### ISO 27002
- [A.8.14](../iso27002/a-8.md#a814)

### NIST 800-53
- [CP-7](../nist80053/cp-7.md)

### SOC 2
- [A1.2-POF10](../soc2/a12-pof10.md)
- [A1.2](../soc2/a12.md)

## Evidence request list


## Control questions
Does the organization establish an alternate processing site that provides security measures equivalent to that of the primary site?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to establish an alternate processing site that provides security measures equivalent to that of the primary site.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to establish an alternate processing site that provides security measures equivalent to that of the primary site.

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
- IT personnel configure business-critical systems to transfer backup data to the alternate site at a rate that is capable of meeting Recovery Time Objectives (RTOs) and Recovery Point Objectives (RPOs).
- the organization acquires space to serve as the alternate site that is in a different geographic zone from the inaccessible facility (e.g., dedicated facility or cloud instance).Roles and responsibilities are formally assigned to restore the primary site in the event of a catastrophe, emergency, or similar-type disruptive incident in accordance with the Continuity of Operations (COOP) plan.
- IT personnel backup copies of business-critical software, licenses/keys and other security-related information to the alternate site.
- IT personnel configure business-critical systems to be able to failover to an alternate location than the primary system which can be activated without loss of information or disruption to operations.
- A dedicated alternate site is identified, documented, equipped to match the processing capabilities of the primary site and not used for existing production activities .
- IT personnel maintain network connectivity for data communications from the alternate site toother business critical locations in order to support business processes.
- IT personnel maintain redundant network connectivity from the alternate site to the business locations providing the ability for data communications to support business processes.
- IT personnel maintain technologies compatible with existing network and infrastructure configuration. the organization either owns the facility or contracts with a third-party provider for off-site storage.

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
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to establish an alternate processing site that provides security measures equivalent to that of the primary site.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to establish an alternate processing site that provides security measures equivalent to that of the primary site.
