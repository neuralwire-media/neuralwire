## Bug Description & Symptom
<!-- Describe the defect, incorrect behavior, or error observed. -->

## Root Cause Analysis
<!-- Detail the exact underlying cause discovered in code, data flow, or timing. -->

## Proposed Fix
<!-- Explain the minimal, correct fix implemented to resolve the root cause. -->

## Scope Lock & Blast Radius
- [ ] **Strict Scope Boundary**: Fix is limited strictly to the bug's root cause.
- [ ] **No Symptom Patching**: The root data flow was addressed rather than defensive null-checks.
- [ ] **Regression Test Added**: Automated test case verifying the fix and preventing regression.

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

### Live Localhost End-to-End (E2E) Verification (Mandatory)
- [ ] Server booted cleanly on localhost (`go run ./cmd/server` / preview) with 0 startup crashes/panics.
- [ ] Live HTTP requests / UI workflows tested end-to-end against localhost matching implementation plan.
- [ ] Server logs verified clean with 0 unhandled errors or unexpected 5xx responses.

## Local Verification Evidence
<!-- Provide the actual terminal execution output summary showing 0 errors across test/lint suites and localhost E2E tests. -->
```text
<!-- Paste local test/build/e2e verification output here -->
```
