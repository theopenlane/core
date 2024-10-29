# NIST 800-53v5 - IA-4 - Identifier Management
- Receiving authorization from \[ Assignment: personnel or roles \] to assign an individual, group, role, service, or device identifier;
- Selecting an identifier that identifies an individual, group, role, service, or device;
- Assigning the identifier to the intended individual, group, role, service, or device; and
- Preventing reuse of identifiers for \[ Assignment: time period \].
## Guidance
Common device identifiers include Media Access Control (MAC) addresses, Internet Protocol (IP) addresses, or device-unique token identifiers. The management of individual identifiers is not applicable to shared system accounts. Typically, individual identifiers are the usernames of the system accounts assigned to those individuals. In such instances, the account management activities of [AC-2](#ac-2) use account names provided by [IA-4](#ia-4) . Identifier management also addresses individual identifiers not necessarily associated with system accounts. Preventing the reuse of identifiers implies preventing the assignment of previously used individual, group, role, service, or device identifiers to different individuals, groups, roles, services, or devices.
## Mapped SCF controls
- [IAC-01.2 - Authenticate, Authorize and Audit (AAA)](../scf/iac-012-authenticate,authorizeandauditaaa.md)
- [IAC-09 - Identifier Management (User Names)](../scf/iac-09-identifiermanagementusernames.md)