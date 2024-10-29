# NIST 800-53v5 - AU-7 - Audit Record Reduction and Report Generation
- Supports on-demand audit record review, analysis, and reporting requirements and after-the-fact investigations of incidents; and
- Does not alter the original content or time ordering of audit records.
## Guidance
Audit record reduction is a process that manipulates collected audit log information and organizes it into a summary format that is more meaningful to analysts. Audit record reduction and report generation capabilities do not always emanate from the same system or from the same organizational entities that conduct audit logging activities. The audit record reduction capability includes modern data mining techniques with advanced data filters to identify anomalous behavior in audit records. The report generation capability provided by the system can generate customizable reports. Time ordering of audit records can be an issue if the granularity of the timestamp in the record is insufficient.
## Mapped SCF controls
- [MON-06 - Monitoring Reporting](../scf/mon-06-monitoringreporting.md)