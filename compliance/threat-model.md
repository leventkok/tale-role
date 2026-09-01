# Threat model (F0)

Educational baseline. Expand before any production traffic.

## Assets

- User emails and session tokens
- JWT signing secret
- Postgres + Mongo connection strings
- Game transcripts (may contain player secrets)
- LLM prompts and adapter files
- Electron product licenses

## Trust boundaries

Browser / Electron (untrusted) → Next.js server → masterfabric-go → Postgres / Mongo / Redis → LLM gateway.

## Abuse cases

| ID | Case | Mitigation |
| --- | --- | --- |
| T1 | Session theft | HttpOnly Secure cookies; no tokens in localStorage |
| T2 | Secret in git | gitleaks CI; `.env` gitignored |
| T3 | LLM invents HP/dice | Go engine is the only writer |
| T4 | Admin identity leak | `system_admin` stripped from public presence and Storyteller context |
| T5 | Prompt injection via player text | Retrieved/player text is data; AuthZ in code |
| T6 | Cross-lobby bleed | Room-scoped channels + org RBAC |

Prompt is not a policy. See [cursor-security MANIFEST](https://github.com/gurkanfikretgunak/cursor-security/blob/main/MANIFEST.md).
