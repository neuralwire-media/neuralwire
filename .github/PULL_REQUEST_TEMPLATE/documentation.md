## Documentation / Rules Summary
<!-- Provide a clear, concise overview of what documentation, specification, developer rules, or guidelines are being added or updated. -->

## Motivation & Context
<!-- Explain why this documentation update or rule change is needed (e.g. clarify workflow, document new APIs, agent guardrails, CI/CD procedures). -->

## Impact & Scope
- [ ] **Accurate & Verified**: All code examples, command sequences, and paths have been verified against the current codebase.
- [ ] **Consistent Style**: Follows existing documentation formatting and tone standards.
- [ ] **No Unrelated Churn**: Changes are strictly confined to the requested documentation/rules scope.

## Mandatory Zero-Failure Pre-Commit Verification (AGENTS.md)
*All commands below must be executed locally and pass with Exit Code 0 prior to creating the PR:*

### Backend Suite (if backend code changed)
- [ ] `gofmt -w . && gofmt -l .` (Output must be completely empty)
- [ ] `go vet ./...` (Static analysis clean)
- [ ] `go test -count=1 -v ./...` (All tests pass with cache disabled)

### Frontend Suite (if frontend code changed)
- [ ] `npm run format && npm run lint` (Prettier code style & ESLint clean)
- [ ] `npm run check` (SvelteKit TypeScript validation clean)
- [ ] `npm run build` (Production build succeeds with static adapter)

### Live Localhost End-to-End (E2E) Verification (if runtime code changed)
- [ ] Server booted cleanly on localhost (`go run ./cmd/server` / preview) with 0 startup crashes/panics.
- [ ] Live HTTP requests / UI workflows tested end-to-end against localhost matching implementation plan.
- [ ] Server logs & responses verified clean with 0 unexpected HTTP errors (no unintended 4xx/5xx responses) and proper status codes (2xx/explicit error contracts).

## Local Verification Evidence
<!-- Provide the actual terminal execution output summary showing 0 errors across test/lint suites and localhost E2E tests. -->
```text
<!-- Paste local test/build/e2e verification output here -->
```
