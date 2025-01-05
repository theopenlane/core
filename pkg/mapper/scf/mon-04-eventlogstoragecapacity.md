# SCF - MON-04 - Event Log Storage Capacity
Mechanisms exist to allocate and proactively manage sufficient event log storage capacity to reduce the likelihood of such capacity being exceeded.
## Mapped framework controls
### NIST 800-53
- [AU-4](../nist80053/au-4.md)

## Evidence request list


## Control questions
Does the organization allocate and proactively manage sufficient event log storage capacity to reduce the likelihood of such capacity being exceeded?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to allocate and proactively manage sufficient event log storage capacity to reduce the likelihood of such capacity being exceeded.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to allocate and proactively manage sufficient event log storage capacity to reduce the likelihood of such capacity being exceeded.

### Planned and tracked
Continuous Monitoring (MON) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Situational awareness management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- Secure baseline configurations generate logs that contain sufficient information to establish necessary details of activity and allow for forensics analysis.
- IT/cybersecurity personnel:
o	Identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for situational awareness management.
o	Configure alerts for critical or sensitive data that is stored, transmitted and processed on assets.
o	Use a structured process to review and analyze logs.
- A log aggregator, or similar automated tool, provides an event log report generation capability to aid in detecting and assessing anomalous activities on business-critical systems.

### Well defined
Continuous Monitoring (MON) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- An IT Asset Management (ITAM) function, or similar function:
o	Governs asset management that ensures compliance with requirements for asset management.
o	Leverages a Configuration Management Database (CMDB), or similar tool, as the authoritative source of IT assets.
- A Security Incident Event Manager (SIEM), or similar automated tool:
o	Centrally collects logs and is protected according to the manufacturer’s security guidelines to protect the integrity of the event logs with cryptographic mechanisms.
o	Monitors the organization for Indicators of Compromise (IoC) and provides 24x7x365 near real-time alerting capability.
o	Is configured to alert incident response personnel of detected suspicious events such that incident responders can look to terminate suspicious events.
- Both inbound and outbound network traffic is monitored for unauthorized activities to identify prohibited activities and assist incident handlers with identifying potentially compromised systems.

### Quantitatively controlled
Continuous Monitoring (MON) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to allocate and proactively manage sufficient event log storage capacity to reduce the likelihood of such capacity being exceeded.
