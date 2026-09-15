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

### 1.3 Truthfulness & Real Execution Invariant
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

1. **Explicit Permission Required**:
   - NEVER execute `git commit`, `git push`, or create a Pull Request without explicit confirmation from the user.
2. **Target Branch Awareness**:
   - Active development branch is `development`. Base target for Pull Requests is `main`.
3. **Scope Lock**:
   - Limit code modifications strictly to the task requested. Avoid speculative refactoring or style churn outside the feature scope.
