# SCF - END-07 - Host Intrusion Detection and Prevention Systems (HIDS / HIPS)
Mechanisms exist to utilize Host-based Intrusion Detection / Prevention Systems (HIDS / HIPS), or similar technologies, to monitor for and protect against anomalous host activity, including lateral movement across the network.
## Mapped framework controls
### SOC 2
- [CC6.8](../soc2/cc68.md)

## Evidence request list


## Control questions
Does the organization utilize Host-based Intrusion Detection / Prevention Systems (HIDS / HIPS), or similar technologies, to monitor for and protect against anomalous host activity, including lateral movement across the network?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to utilize Host-based Intrusion Detection / Prevention Systems (HIDS / HIPS), or similar technologies, to monitor for and protect against anomalous host activity, including lateral movement across the network.

### Performed internally
Endpoint Security (END) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Asset management is informally assigned as an additional duty to existing IT/cybersecurity personnel.
- IT/cybersecurity personnel use an informal process to design, build and maintain secure configurations for test, development, staging and production environments, including the implementation of appropriate cybersecurity & data privacy controls.
- Anti-malware technologies are decentralized but are deployed on all technology assets that can run Anti-malware software.
- Data management is decentralized.

### Planned and tracked
Endpoint Security (END) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Endpoint security management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for endpoint security management.
- Anti-malware technologies are decentralized but are deployed on all technology assets that can run anti-malware software.
- Physical controls, administrative processes and technologies focus on protecting High Value Assets (HVAs), including environments where sensitive/regulated data is stored, transmitted and processed.
- Technologies are configured to protect data with the strength and integrity commensurate with the classification or sensitivity of the information and mostly conform to industry-recognized standards for hardening (e.g., DISA STIGs, CIS Benchmarks or OEM security guides), including cryptographic protections for sensitive/regulated data.

### Well defined
Endpoint Security (END) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Configuration management is centralized for all operating systems, applications, servers and other configurable technologies.
- Technologies are configured to protect data with the strength and integrity commensurate with the classification or sensitivity of the information and conform to industry-recognized standards for hardening (e.g., DISA STIGs, CIS Benchmarks or OEM security guides), including test, development, staging and production environments.
- Configurations conform to industry-recognized standards for hardening (e.g., DISA STIGs, CIS Benchmarks or OEM security guides) for test, development, staging and production environments.
- An Identity & Access Management (IAM) function, or similar function, centrally manages permissions and implements “least privileges” practices for the management of user, group and system accounts, including privileged accounts.
- An IT Asset Management (ITAM) function, or similar function, categorizes endpoint devices according to the data the asset stores, transmits and/ or processes and applies the appropriate technology controls to protect the asset and data that conform to industry-recognized standards for hardening (e.g., DISA STIGs, CIS Benchmarks or OEM security guides).
- A Security Operations Center (SOC), or similar function, centrally manages anti-malware and anti-phishing technologies, in accordance with industry-recognized practices for Prevention, Detection & Response (PDR) activities.
- A Security Incident Event Manager (SIEM), or similar automated tool, is tuned to detect and respond to anomalous behavior that could indicate account compromise or other malicious activities.
- The Human Resources (HR) department ensures that every user accessing a system that processes, stores, or transmits sensitive/regulated data is cleared and regularly trained in proper data handling practices.
- Unauthorized configuration changes are responded to in accordance with an Incident Response Plan (IRP) to determine if the any unauthorized configuration is malicious in nature.

### Quantitatively controlled
Endpoint Security (END) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to utilize Host-based Intrusion Detection / Prevention Systems (HIDS / HIPS), or similar technologies, to monitor for and protect against anomalous host activity, including lateral movement across the network.
