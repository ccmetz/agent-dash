## Agent skills

### Issue tracker

Issues and PRDs live in GitHub Issues. See `docs/agents/issue-tracker.md`.

### Triage labels

Triage labels use the default mattpocock/skills vocabulary. See `docs/agents/triage-labels.md`.

### Domain docs

This is a single-context repo. See `docs/agents/domain.md`.

## Implementation

### JavaScript package manager

Use `pnpm` for all JavaScript dependency and script commands. Do not use `npm` or `npx` in this repo.

### Code Checks

- `pnpm --dir frontend lint`
- `pnpm --dir frontend format:check`
- `pnpm --dir frontend typecheck`
- `pnpm test` (runs all tests in backend and frontend)
- `cd backend && go fmt ./...`
