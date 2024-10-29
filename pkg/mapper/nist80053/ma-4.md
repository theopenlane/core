# NIST 800-53v5 - MA-4 - Nonlocal Maintenance
- Approve and monitor nonlocal maintenance and diagnostic activities;
- Allow the use of nonlocal maintenance and diagnostic tools only as consistent with organizational policy and documented in the security plan for the system;
- Employ strong authentication in the establishment of nonlocal maintenance and diagnostic sessions;
- Maintain records for nonlocal maintenance and diagnostic activities; and
- Terminate session and network connections when nonlocal maintenance is completed.
## Guidance
Nonlocal maintenance and diagnostic activities are conducted by individuals who communicate through either an external or internal network. Local maintenance and diagnostic activities are carried out by individuals who are physically present at the system location and not communicating across a network connection. Authentication techniques used to establish nonlocal maintenance and diagnostic sessions reflect the network access requirements in [IA-2](#ia-2) . Strong authentication requires authenticators that are resistant to replay attacks and employ multi-factor authentication. Strong authenticators include PKI where certificates are stored on a token protected by a password, passphrase, or biometric. Enforcing requirements in [MA-4](#ma-4) is accomplished, in part, by other controls. [SP 800-63B](#e59c5a7c-8b1f-49ca-8de0-6ee0882180ce) provides additional guidance on strong authentication and authenticators.
## Mapped SCF controls
- [MNT-05 - Remote Maintenance](../scf/mnt-05-remotemaintenance.md)
- [MNT-05.1 - Auditing Remote Maintenance](../scf/mnt-051-auditingremotemaintenance.md)
- [MNT-05.2 - Remote Maintenance Notifications](../scf/mnt-052-remotemaintenancenotifications.md)