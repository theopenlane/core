# NIST 800-53v5 - CP-9 - System Backup
- Conduct backups of user-level information contained in \[ Assignment: system components \] {{ insert: param, cp-09_odp.02 }};
- Conduct backups of system-level information contained in the system \[ Assignment: frequency \];
- Conduct backups of system documentation, including security- and privacy-related documentation \[ Assignment: frequency \] ; and
- Protect the confidentiality, integrity, and availability of backup information.
-
## Guidance
System-level information includes system state information, operating system software, middleware, application software, and licenses. User-level information includes information other than system-level information. Mechanisms employed to protect the integrity of system backups include digital signatures and cryptographic hashes. Protection of system backup information while in transit is addressed by [MP-5](#mp-5) and [SC-8](#sc-8) . System backups reflect the requirements in contingency plans as well as other organizational requirements for backing up information. Organizations may be subject to laws, executive orders, directives, regulations, or policies with requirements regarding specific categories of information (e.g., personal health information). Organizational personnel consult with the senior agency official for privacy and legal counsel regarding such requirements.
## Mapped SCF controls
- [BCD-11 - Data Backups](../scf/bcd-11-databackups.md)