# Security

Report vulnerabilities privately. Do not file public issues for secrets or exploitable bugs.

## Contact

Open a **private** advisory on GitHub Security if available, or email the maintainers listed on the repo.

## Baseline

- Secrets stay in a CI/host secret store, never in git
- AuthZ is enforced in code (masterfabric-go RBAC), not in LLM prompts
- Player-facing APIs must not leak `system_admin` spectators
- LLM prompts are PII-redacted before send
- Aligns with [cursor-security](https://github.com/gurkanfikretgunak/cursor-security) MANIFEST principles and [masterfabric-go](https://github.com/gurkanfikretgunak/masterfabric-go) hardening notes
