# Security Policy

User safety and the protection of user credentials are the top priority for the **p-manager** project. We take all vulnerability reports seriously and greatly value the assistance of security researchers.

---

## Supported versions

Security fixes are released only for the current release and the main development branch.

| Version | Security Support |
| :--- | :--- |
| `v1.x.x` (Latest release) | Supported |
| `main` branch | Supported |
| Older releases | Not supported (please upgrade) |

---

## How to safely report a vulnerability

**PLEASE DO NOT CREATE PUBLIC GITHUB ISSUES TO REPORT VULNERABILITIES.**

Publicly disclosing the issue before the patch is released could compromise the security of other users' passwords.

### Private communication methods:

1. **Private report on GitHub (Recommended):**
   Go to the **[Security -> Vulnerability reporting](https://github.com/stepan41k/p-manager/security/advisories/new)** tab in this repository and click the **"Report a vulnerability"** button. Only the project maintainers will see this report.

2. **Прямая связь:**
   If the reporting functionality is unavailable, please contact the author directly using the contact details in the [@stepan41k](https://github.com/stepan41k) profile. 

---

## What to include in the report

Please include as many details as possible in your message:

1. Vulnerability description (e.g., issue with cryptographic IV generation, master key leakage in logs, insecure data transmission to S3).
2. Steps to reproduce or PoC (Proof of Concept).
3. Potential impact on users.
4. `p-manager` version and your OS.

---

## Our response process

* **Acknowledgment:** We will aim to respond and acknowledge receipt of the report within **24–48 hours**.
* **Fix:** Once the issue is verified and confirmed, we will prepare a private patch.
* **Release:** We will release an updated version of `p-manager` and publish a Security Advisory crediting the researcher (unless you prefer to remain anonymous).

---

## Safety recommendations for users

* Always use a strong master password that is not used anywhere else.
* Never share your S3 storage access keys (Selectel/AWS) with third parties.
* Regularly update `p-manager` to the latest version.