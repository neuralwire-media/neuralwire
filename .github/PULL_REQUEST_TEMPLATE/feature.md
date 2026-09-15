## Feature Summary
<!-- Provide a clear, concise overview of the new feature or enhancement. -->

## Architectural & Design Decisions
<!-- Explain the technical approach, data flow changes, new components, and rationale. -->

## Scope Lock & Blast Radius
- [ ] **Strict Scope Boundary**: Changes are strictly confined to the requested feature.
- [ ] **No Unrelated Churn**: No unrelated files, styles, or dependency upgrades included.
- [ ] **Backwards Compatibility**: Existing public API contracts and data models remain compatible.

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

## Visual Changes / UI Mockups (if applicable)
<!-- Attach screenshots, screen recordings, or before/after comparisons if UI changed. -->
