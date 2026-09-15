## Refactor Motivation & Context
<!-- Describe why this refactoring is needed (e.g. anti-duplication, performance tuning, architectural alignment). -->

## Summary of Code Changes
<!-- Outline what was reorganized, consolidated, or optimized. -->

## Scope Lock & Non-Breaking Invariants
- [ ] **Zero Behavioral Drift**: Public contracts, API endpoints, and response formats are preserved.
- [ ] **Scope Contained**: No speculative features or unrelated cleanups added.
- [ ] **No Error Weakening**: Compiler, linter, or type checks were not suppressed or relaxed.

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

## Local Verification Evidence
<!-- Provide the actual terminal execution output summary showing 0 errors across test/lint suites. -->
```text
<!-- Paste local test/build verification output here -->
```
