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

## Local Verification Evidence
<!-- Provide the actual terminal execution output summary showing 0 errors across test/lint suites. -->
```text
<!-- Paste local test/build verification output here -->
```
