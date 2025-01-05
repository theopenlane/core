# SCF - IAC-21 - Least Privilege
Mechanisms exist to utilize the concept of least privilege, allowing only authorized access to processes necessary to accomplish assigned tasks in accordance with organizational business functions.
## Mapped framework controls
### ISO 27002
- [A.5.15](../iso27002/a-5.md#a515)
- [A.5.18](../iso27002/a-5.md#a518)
- [A.8.12](../iso27002/a-8.md#a812)
- [A.8.3](../iso27002/a-8.md#a83)

### NIST 800-53
- [AC-6](../nist80053/ac-6.md)

### SOC 2
- [CC5.2-POF3](../soc2/cc52-pof3.md)
- [CC6.1-POF12](../soc2/cc61-pof12.md)
- [CC6.1-POF13](../soc2/cc61-pof13.md)
- [CC6.1-POF7](../soc2/cc61-pof7.md)
- [CC6.1](../soc2/cc61.md)

## Evidence request list


## Control questions
Does the organization utilize the concept of least privilege, allowing only authorized access to processes necessary to accomplish assigned tasks in accordance with organizational business functions?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to utilize the concept of least privilege, allowing only authorized access to processes necessary to accomplish assigned tasks in accordance with organizational business functions.

### Performed internally
Identification & Authentication (IAC) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Identity & Access Management (IAM) is decentralized where Active Directory (AD), or a similar technology, may be used to centrally manage identities and permissions, but asset/process owners are authorized to operate a decentralized access control program for their specific systems, applications and services.
- IAM controls are primarily administrative in nature (e.g., policies & standards) to manage accounts and permissions.
- IT personnel identify and implement IAM cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements.

### Planned and tracked
Identification & Authentication (IAC) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Logical Access Control (LAC) is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for logical access control.
- IT personnel:
o	Implement and maintain an Identity & Access Management (IAM) capability for all users to implement “least privileges” Role Based Access Control (RBAC) practices for the management of user, group and system accounts, including privileged accounts.
o	Govern IAM technologies via RBAC to prohibit privileged access by non-organizational users, unless there is an explicit support contract for privileged IT support services.
- Active Directory (AD), or a similar technology, is primarily used to centrally manage identities and permissions with RBAC. Due to technical or business limitations, asset/process owners are empowered to operate a decentralized access control program for their specific systems, applications and services that cannot be integrated into AD.
- IAM controls are primarily administrative in nature (e.g., policies & standards) to manage accounts and permissions.

### Well defined
Identification & Authentication (IAC) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- An Identity & Access Management (IAM) function, or similar function, centrally manages permissions and implements “least privileges” Role Based Access Control (RBAC) practices for the management of user, group and system accounts, including privileged accounts.
- The Human Resources (HR) department governs personnel management operations and notifies IAM personnel of personnel role changes for RBAC-based provisioning and deprovisioning actions.
- An IT Asset Management (ITAM) function, or similar function, categorizes endpoint devices according to the data the asset stores, transmits and/ or processes and applies the appropriate technology controls to protect the asset and data that conform to industry-recognized standards for hardening (e.g., DISA STIGs, CIS Benchmarks or OEM security guides).
- An IT infrastructure team, or similar function, ensures that statutory, regulatory and contractual cybersecurity & data privacy obligations are addressed to ensure secure configurations are designed, built and maintained.
- Active Directory (AD), or a similar technology, is used to centrally manage identities and permissions. Only by exception due to a technical or business limitation are solutions authorized to operate a decentralized access control program for systems, applications and services.

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to utilize the concept of least privilege, allowing only authorized access to processes necessary to accomplish assigned tasks in accordance with organizational business functions.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to utilize the concept of least privilege, allowing only authorized access to processes necessary to accomplish assigned tasks in accordance with organizational business functions.
