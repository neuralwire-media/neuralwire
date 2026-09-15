# Neuralwire — Agentic Development Guidelines & Invariants

This file defines the persistent rules, verification protocols, and coding standards for all AI coding agents working on the Neuralwire codebase.

---

## 1. Zero-Failure Pre-Commit Verification Protocol (MANDATORY)

Before creating any `git commit`, pushing to remote, or proposing a Pull Request, **ALL** verification steps below MUST be executed locally and pass with Exit Code 0.

### 1.1 Backend Verification Suite
Navigate to `backend/` and execute:
```bash
# 1. Format code and verify 0 unformatted files exist
gofmt -w . && gofmt -l .

# 2. Go static analysis
go vet ./...

# 3. Complete test suite with cache disabled
go test -count=1 -v ./...
```
*Criteria*:
- `gofmt -l .` output MUST be completely empty.
- `go vet` and `go test` MUST exit with 0.

### 1.2 Frontend Verification Suite
Navigate to `frontend/` and execute:
```bash
# 1. Code style formatting and linting (Prettier + ESLint)
npm run format && npm run lint

# 2. Typechecking
npm run check

# 3. Production build
npm run build
```
*Criteria*:
- `npm run lint` MUST pass with "All matched files use Prettier code style!" and 0 ESLint errors.
- `npm run check` and `npm run build` MUST exit with 0.

### 1.3 Mandatory Live Localhost End-to-End (E2E) Verification Protocol
Before proposing a commit or Pull Request, the agent/developer MUST execute live localhost smoke testing:
1. **Boot Live Services on Localhost**:
   - Start the Go backend server (and frontend preview if applicable) on localhost with a test configuration.
2. **Execute Real End-to-End Tests**:
   - Test all user scenarios, modified endpoints, and feature flows against the running localhost server via real HTTP requests (`curl` or API client).
   - Validate live HTTP status codes (200 OK / 201 Created / 204 No Content for success paths; explicit validated status codes for negative test cases), response payloads, database state changes, and server logs.
3. **Confirm Zero Runtime Defects & Full HTTP Health**:
   - Verify that the running server produces 0 panics, 0 unhandled runtime errors, 0 unhandled route misses (unexpected 404), 0 unintended authorization failures (unexpected 401/403), 0 payload/validation errors (unexpected 400/422), and 0 server/gateway errors (500/502/503/504) during the entire E2E test run.
   - For error simulation or negative flows, strictly verify that the returned HTTP status code and error schema match expected specifications.
4. **Graceful Cleanup**:
   - Gracefully terminate the test server process after verification completes.

### 1.4 Truthfulness & Real Execution Invariant
- **NEVER** claim or fabricate verification results without actually running the shell commands.
- **NEVER** bypass or suppress compiler, linter, or typechecker errors using quick-fix escapes (e.g. `@ts-ignore`, `any`, `eslint-disable`, empty `catch` blocks).

---

## 2. Architecture & Project Conventions

### 2.1 Backend (Go)
- **Framework**: Standard library `net/http` with Go 1.22+ routing patterns (`GET /api/...`, `POST /api/...`).
- **Database**: SQLite with `database/sql` and custom migrations in `backend/internal/database/migrate.go`.
- **Security Invariants**:
  - Always use `netutil.SafeHTTPClient` or `netutil.SafeDialContext` for outgoing requests (SSRF defense).
  - Wrap JSON/multipart request bodies with `http.MaxBytesReader` to prevent memory DoS.
  - Enforce standard security headers (COOP, CORP, HSTS, X-Content-Type-Options, etc.) in `api/security.go`.

### 2.2 Frontend (SvelteKit)
- **Framework**: SvelteKit with Svelte 5 runes (`$state`, `$derived`, `$effect`, `$props`).
- **Styling**: TailwindCSS with cyber/editorial dark theme.
- **Build Adapter**: `@sveltejs/adapter-static` for static frontend SPA generation.

---

## 3. Developer & Git Guardrails

### 3.1 Feature Branching Strategy (MANDATORY)
1. **Base & Target Branch**:
   - `main` is the single source of truth for stable production code.
   - All tasks (features, bug fixes, refactors, docs, security fixes, CI) MUST branch off the latest `main`.
2. **Branch Naming Standard (Kebab-Case)**:
   - Features: `feat/<feature-name>` (e.g. `feat/interactive-cluster-ui`)
   - Bug fixes: `fix/<bug-name>` (e.g. `fix/legacy-schema-migration`)
   - Documentation & Rules: `docs/<topic>` (e.g. `docs/git-branching-strategy`)
   - Refactoring: `refactor/<component>` (e.g. `refactor/news-repository`)
   - Security Hardening: `security/<scope>` (e.g. `security/ssrf-safedialer`)
   - CI & Tooling: `ci/<pipeline>` (e.g. `ci/github-multi-template`)
3. **Workflow Lifecycle**:
   1. `git checkout main && git pull origin main`
   2. `git checkout -b <type>/<kebab-case-name>`
   3. Develop & execute mandatory local verification suite + live localhost E2E smoke test.
   4. Obtain explicit user confirmation before committing.
   5. `git push origin <type>/<kebab-case-name>`
   6. Open Pull Request with target `base: main` using the appropriate PR template.
   7. Confirm all GitHub Actions CI checks pass with 100% green status.

### 3.2 Permissions & Scope Lock Invariants
1. **Explicit Permission Required**:
   - NEVER execute `git commit`, `git push`, or create a Pull Request without explicit confirmation from the user.
2. **Scope Lock Invariant**:
   - Limit code modifications strictly to the task requested. Avoid speculative refactoring or style churn outside the feature scope.

