# NIST 800-53v5 - SI-7 - Software, Firmware, and Information Integrity
- Employ integrity verification tools to detect unauthorized changes to the following software, firmware, and information: \[ Assignment: organization-defined software, firmware, and information \] ; and
- Take the following actions when unauthorized changes to the software, firmware, and information are detected: \[ Assignment: organization-defined actions \].
## Guidance
Unauthorized changes to software, firmware, and information can occur due to errors or malicious activity. Software includes operating systems (with key internal components, such as kernels or drivers), middleware, and applications. Firmware interfaces include Unified Extensible Firmware Interface (UEFI) and Basic Input/Output System (BIOS). Information includes personally identifiable information and metadata that contains security and privacy attributes associated with information. Integrity-checking mechanisms—including parity checks, cyclical redundancy checks, cryptographic hashes, and associated tools—can automatically monitor the integrity of systems and hosted applications.
## Mapped SCF controls
- [END-06 - Endpoint File Integrity Monitoring (FIM)](../scf/end-06-endpointfileintegritymonitoringfim.md)
- [NET-12 - Safeguarding Data Over Open Networks](../scf/net-12-safeguardingdataoveropennetworks.md)
- [TDA-18 - Input Data Validation](../scf/tda-18-inputdatavalidation.md)