# Contributing

Tale Role is developed in public. **GitHub is the only transfer path.**

## Workflow

1. Branch from `main`: `feat/…`, `fix/…`, `ci/…`, `docs/…`, `security/…`
2. Conventional commits (`feat: add join-link verification`)
3. Open a pull request using the template (summary, test plan, security note)
4. Wait for CI (typecheck, tests, gitleaks, audit)
5. Merge only when CI is green

Do not push to `main`. Do not commit secrets. Draft PRs are welcome before a feature is finished.

## Style

Match [cursor-security](https://github.com/gurkanfikretgunak/cursor-security) PR titles and bodies.
