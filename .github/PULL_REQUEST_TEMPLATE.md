## Summary

-

## Test plan

- [ ] `npm test` (workspace packages)
- [ ] No secrets, `.env`, or credentials in the diff
- [ ] i18n: no new hardcoded user-facing strings (or both `en` and `tr` catalogs updated)

## Security

- [ ] AuthZ stays in code (prompt is not policy)
- [ ] Player payloads cannot include invisible admin spectators
- [ ] `.env.example` uses placeholders only
