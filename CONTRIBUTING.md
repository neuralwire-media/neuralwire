# Contributing to Neuralwire

Thanks for your interest in contributing to Neuralwire! We welcome
contributions that improve the platform while keeping the curator model
intact: AI assists, humans decide what gets published.

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before participating.

## Ways to Contribute

- **Report bugs** — open an issue with clear steps to reproduce.
- **Suggest features** — open an issue describing the problem you're solving.
- **Fix / implement** — fork the repo, create a branch, open a pull request.

## Development Setup

### Prerequisites
- Go 1.25+
- Node.js 24+
- npm

### Backend
```bash
cd backend
cp .env.example .env   # configure env (defaults work for local dev)
go mod download
go run ./cmd/server    # starts on :8080
```

### Frontend
```bash
cd frontend
cp .env.example .env.local   # set PUBLIC_API_URL=http://localhost:8080/api
npm install
npm run dev                  # starts on :5173
```

## Branch & PR Workflow

Neuralwire follows the **Feature Branching Strategy (GitHub Flow)** off `main`:

```
main (protected production branch, deploys on merge)
  │
  ├── branch off main: feat/*, fix/*, docs/*, refactor/*, security/*, ci/*
  │     (develop locally, run verification suite, test localhost E2E)
  │
  └── open Pull Request -> main (CI green + review -> merge -> deploy)
```

### 1. Branch Naming Standard (Kebab-Case)
- Features: `feat/<feature-name>` (e.g., `feat/interactive-cluster-ui`)
- Bug fixes: `fix/<bug-name>` (e.g., `fix/legacy-schema-migration`)
- Documentation: `docs/<topic>` (e.g., `docs/git-branching-strategy`)
- Refactoring: `refactor/<component>` (e.g., `refactor/news-repository`)
- Security Hardening: `security/<scope>` (e.g., `security/ssrf-safedialer`)
- CI & Tooling: `ci/<pipeline>` (e.g., `ci/github-multi-template`)

### 2. Development & Verification Lifecycle
1. Pull latest `main`: `git checkout main && git pull origin main`
2. Create task branch: `git checkout -b <type>/<kebab-case-name>`
3. Implement changes with minimal diff and zero error suppression.
4. Execute mandatory local verification:
   - Backend: `gofmt -w . && gofmt -l . && go vet ./... && go test -count=1 ./...`
   - Frontend: `npm run format && npm run lint && npm run check && npm run build`
   - Live Localhost Smoke Test: Boot server locally and test endpoints end-to-end.
5. Push branch and open a **Pull Request** targeting `main` using the appropriate template in `.github/PULL_REQUEST_TEMPLATE/`.
6. Confirm all GitHub Actions CI checks pass with 100% green status before merge.

> `main` is protected: direct pushes and force-pushes are blocked. All code enters `main` via PRs with passing CI.

## Coding Standards

### Backend (Go)
- Run `gofmt` before committing (`gofmt -l .` must output nothing).
- Run `go vet ./...` and `go test ./...` locally.
- Keep dependencies minimal; prefer the standard library.

### Frontend (SvelteKit)
- Run `npm run check` (svelte-check) and `npm run lint` (prettier + eslint).
- Keep TypeScript types strict.
- Do not modify `backend/` in frontend PRs and vice versa.

## Commit Messages

Write concise, descriptive commit messages that explain the *why*:

```
type(scope): short summary

Body explaining context, especially non-obvious decisions.
```

Examples:
- `feat(api): add admin image upload endpoint`
- `fix(ui): article share buttons open correct share URLs`
- `perf(db): add index on news(url)`
- `docs: add contributing guide`

## Project Structure

```
backend/    Go REST API + scheduler
frontend/   SvelteKit app (adapter-static)
.github/    CI workflows
```

See the root [README](README.md) for full details.

## Questions?

Open an issue or reach out via the contact in our [Security Policy](SECURITY.md).
