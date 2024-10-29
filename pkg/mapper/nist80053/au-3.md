# NIST 800-53v5 - AU-3 - Content of Audit Records
- What type of event occurred;
- When the event occurred;
- Where the event occurred;
- Source of the event;
- Outcome of the event; and
- Identity of any individuals, subjects, or objects/entities associated with the event.
## Guidance
Audit record content that may be necessary to support the auditing function includes event descriptions (item a), time stamps (item b), source and destination addresses (item c), user or process identifiers (items d and f), success or fail indications (item e), and filenames involved (items a, c, e, and f) . Event outcomes include indicators of event success or failure and event-specific results, such as the system security and privacy posture after the event occurred. Organizations consider how audit records can reveal information about individuals that may give rise to privacy risks and how best to mitigate such risks. For example, there is the potential to reveal personally identifiable information in the audit trail, especially if the trail records inputs or is based on patterns or time of usage.
## Mapped SCF controls
- [MON-03 - Content of Event Logs](../scf/mon-03-contentofeventlogs.md)