# Neuralwire — AI News & Editorial Platform

<img width="1632" height="1343" alt="Neuralwire Editorial Platform" src="https://github.com/user-attachments/assets/124469d1-a851-41d6-b86c-57424047a3b7" />

> An editorial news portal for artificial intelligence, neural networks, and the future of computation. Bridging the gap between silicon and humanity.

**Live site:** [https://neuralwire.info](https://neuralwire.info)

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Features](#features)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Development Workflow](#development-workflow)
- [CI/CD](#cicd)
- [Deployment & Operations](#deployment--operations)
- [Security](#security)
- [Project Structure](#project-structure)
- [API Reference](#api-reference)
- [License](#license)

---

## Overview

Neuralwire is a **semi-automated news curation platform**. It ingests RSS feeds from curated industry and research sources, extracts readable article content, generates AI summaries and category classifications, scores news value, and presents the results in a modern editorial frontend.

The system uses a **curator model**: AI assists with summarization, categorization, and value scoring, but **human admins make all publishing decisions**. Nothing is auto-published by default — every article is reviewed as a draft before going live. Original full text is **never stored or republished**, respecting source copyright while directing readers to original publishers.

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Browser (visitor)                    │
│        https://neuralwire.info (Cloudflare CDN / WAF)   │
└──────────────────────────┬──────────────────────────────┘
                           │ HTTPS (Port 443)
┌──────────────────────────▼──────────────────────────────┐
│                    Production VPS                       │
│  ┌───────────────────────────────────────────────────┐  │
│  │             Caddy (Reverse Proxy / TLS)           │  │
│  └───────────────────────┬───────────────────────────┘  │
│                          │ Reverse proxy (127.0.0.1:8080)│
│  ┌───────────────────────▼───────────────────────────┐  │
│  │            Docker (single container)              │  │
│  │  ┌──────────────────────────┐  ┌───────────────┐  │  │
│  │  │   Go Backend (port 8080) │  │   SvelteKit   │  │  │
│  │  │   REST API + SPA server  │  │ adapter-static│  │  │
│  │  └────────────┬─────────────┘  └───────▲───────┘  │  │
│  │               │ serves /api/* & static │          │  │
│  │  ┌────────────▼─────────────┐          │          │  │
│  │  │  SQLite (volume mount)   │          │          │  │
│  │  │  /app/data/neuralwire.db │          │          │  │
│  │  └──────────────────────────┘          │          │  │
│  └────────────────────────────────────────┼──────────┘  │
└───────────────────────────────────────────┼─────────────┘
                                            │
                       ┌────────────────────▼───────────┐
                       │   Curated RSS / Atom Feeds     │
                       │   + OpenAI / Gemini / DeepSeek │
                       └────────────────────────────────┘
```

**Key design decisions:**
- **Unified Single Container**: The SvelteKit frontend is built with `@sveltejs/adapter-static` and **served directly by the Go backend** (same origin). No separate frontend node server is required in production.
- **Dynamic SEO & LCP Optimization**: The Go server intercepts HTML requests to dynamically inject `<head>` fetch preloads, image preconnects, LCP hero image preload tags, and server-rendered SEO metadata.
- **Embedded Zero-CGO Database**: SQLite via `modernc.org/sqlite` provides ACID compliance, fast in-process query execution, and zero CGO dependencies.

---

## Tech Stack

### Backend
- **Go 1.25+** — high-performance, type-safe standard library REST API (`net/http`, Go 1.22+ routing patterns)
- **SQLite** (`modernc.org/sqlite`) — pure-Go, zero-CGO, single-file database with WAL mode and custom migrations
- **go-readability** (`codeberg.org/readeck/go-readability/v2`) — article content extraction
- **gofeed** (`github.com/mmcdole/gofeed`) — robust RSS/Atom feed parsing
- **Structured Logging** (`log/slog`) — JSON/text structured logging with contextual metadata

### Frontend
- **SvelteKit 2** & **Svelte 5** — modern compiler-based UI using Svelte 5 Runes (`$state`, `$derived`, `$props`, `$effect`)
- **Tailwind CSS 4** — utility-first styling with cyber/editorial dark theme
- **@sveltejs/adapter-static** — prerendered static client SPA
- **TypeScript** & **Vite** — end-to-end type safety and rapid HMR

### Infrastructure & DevOps
- **VPS + Docker Compose** — containerized deployment running non-root `neuralwire` user
- **Caddy** — high-performance reverse proxy and automated TLS management
- **Cloudflare** — CDN caching, DDoS mitigation, DNS, and WAF protection
- **GitHub Actions** — multi-stage CI/CD (`backend-ci.yml`, `frontend-ci.yml`, `deploy.yml`) with automated post-deploy container health retry probes

---

## Features

### Content Pipeline & Ingestion
- **25 Default Curated Sources**: Pre-seeded across 5 core categories (AI, Tools, Research, Industry, Machine Learning).
- **Dynamic Source Management**: Add, edit, enable/disable, delete, or test feed URLs on demand via the admin panel.
- **Polite Crawling & SSRF Protection**: 1–2s randomized politeness delay between requests and safe dialer IP verification (`netutil.SafeHTTPClient`).
- **Hybrid Content Extraction**: Full-content extraction via Readability with fallback to RSS excerpts and low-quality content filters (`SCRAPE_MIN_CONTENT_CHARS`).
- **CDN Image Upgrading**: Automatically detects and upgrades low-res thumbnails to high-resolution variants (Contentful, imgix, Cloudinary, Unsplash, Google, WordPress).

### Curator Moderation Model
- **Drafts First**: Fetched articles are stored as drafts and never published automatically unless explicitly configured.
- **Level 2 AI Value Scoring**: Weighted rating ($0.6 \times \text{AI score} + 0.4 \times \text{Heuristic score}$) assigning HIGH/MEDIUM/LOW badges, confidence metrics, impact/novelty/quality sub-scores, and reasoning.
- **Configurable Thresholds**: Admins can adjust score boundaries (`low_max`, `medium_min`, `medium_max`, `high_min`) dynamically in `app_settings`.
- **Live Fetch Control**: Real-time progress monitoring (`/api/admin/fetch/progress`) and immediate cancellation (`/api/admin/fetch/cancel`).
- **Bulk Operations**: Bulk publish, bulk reject, bulk delete, and primary/featured article toggling.

### Automation & Autopublish Scheduler
- **Optional Timer Scheduler**: Configurable cycle interval (minimum 5 minutes) running in background.
- **Granular Auto-Post Filters**: Filter by category whitelist, value labels (`["medium", "high"]`), and max posts per cycle.
- **Independent Controls**: Separate start/stop state management (`/api/admin/autopublish/start` and `/stop`).

### Public Editorial Portal
- **Home Feed & Categories**: Multi-category browsing with client-side reactive "Load More" pagination and quick "Collapse Feed" FAB.
- **Trending / Most-Read**: Deduplicated view tracking per browser (`nw_viewer_id` + 6h IP cooldown) returning top-read stories.
- **Related Articles**: TF-IDF weighted similarity matching by keyword overlap, category, and source.
- **Backend Full-Text Search**: Instant debounced search across all published articles with multi-token AND matching.
- **Dynamic SEO & Feeds**: Backend-generated `/sitemap.xml` and `/robots.txt` with Open Graph and Twitter Card tags.

### Operations, Analytics & Security
- **Analytics Dashboard**: Real-time stats on page views, published counts, draft pipelines, and top-performing articles.
- **Automated Database Backups**: Periodic gzip snapshots with retention pruning and on-demand download via `/api/admin/backup`.
- **Prometheus Metrics & Health Checks**: `/api/metrics` (request counts, latencies, AI calls, fetch stats) and `/api/healthz` (DB ping).
- **Security Hardening**: HMAC bearer tokens, CSRF origin verification on mutations, strict rate limiting (login, views, global), and security headers (CSP, X-Frame-Options, nosniff, COOP, CORP).

---

## Getting Started

### Prerequisites
- **Go 1.25+**
- **Node.js 24+** and **npm**
- **Docker & Docker Compose** (for containerized setup)

---

### Local Development Setup

#### 1. Clone the repository
```bash
git clone git@github.com:neuralwire-media/neuralwire.git
cd neuralwire
```

#### 2. Backend Setup
```bash
cd backend
cp .env.example .env        # Configure your local settings
go mod download
go run ./cmd/server         # Starts server on http://localhost:8080
```

#### 3. Frontend Setup
```bash
cd frontend
cp .env.example .env.local  # Set PUBLIC_API_URL=http://localhost:8080/api
npm install
npm run dev                 # Starts Vite dev server on http://localhost:5173
```

#### 4. Access Local Services
- **Public Portal**: `http://localhost:5173`
- **Admin Panel**: `http://localhost:5173/admin`
- **API Endpoints**: `http://localhost:8080/api`

---

### Running via Docker Compose (Single Container)

To run the unified production build locally:
```bash
# Build and boot container
docker compose up -d --build

# View logs
docker compose logs -f
```
The application will be accessible at `http://localhost:8080`.

---

## Environment Variables

### Backend Configuration (`backend/.env`)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port the server listens on |
| `APP_ENV` | `development` | Set to `production` to enforce security hardening |
| `TRUST_PROXY` | `false` | Set `true` behind reverse proxies (Caddy/Nginx) for accurate IP resolution |
| `DB_PATH` | `data/neuralwire.db` | SQLite database file location |
| `STATIC_DIR` | `../frontend/build` | Directory of built static frontend files (served at `/`) |
| `UPLOAD_DIR` | `./data/uploads` | Directory for admin image uploads (served at `/uploads/`) |
| `BACKUP_DIR` | `./data/backups` | Directory for automated SQLite gzip snapshots |
| `BACKUP_RETENTION` | `7` | Number of backup snapshots to retain |
| `BACKUP_INTERVAL_HOURS` | `24` | Automated backup schedule interval in hours (0 = disabled) |
| `ADMIN_USERNAME` | `admin` | Admin username for `/api/admin/login` |
| `ADMIN_PASSWORD` | `admin123` | Admin password (**must be changed in production**) |
| `ADMIN_TOKEN_SECRET` | *(dev secret)* | HMAC key for signing bearer tokens (**must be changed in production**) |
| `AI_SUMMARY_API_KEY` | *(empty)* | OpenAI-compatible API key for summarization & scoring |
| `AI_SUMMARY_PROVIDER` | `openai` | Preset: `openai`, `gemini`, `openrouter`, `groq`, `ollama` |
| `AI_SUMMARY_BASE_URL` | `https://api.openai.com/v1` | OpenAI-compatible base URL |
| `AI_SUMMARY_MODEL` | `gpt-4o-mini` | AI model used for summaries and value scoring |
| `AI_IMAGE_GENERATION_ENABLED` | *(auto)* | Enable/disable DALL-E style image generation (`true`/`false`) |
| `CORS_ALLOW_ORIGIN` | `http://localhost:5173,http://127.0.0.1:5173` | Comma-separated allowed frontend origins |
| `GLOBAL_RATE_LIMIT` | `120` | Max requests per IP per minute across all endpoints (0 = off) |
| `LOGIN_RATE_LIMIT` | `5` | Max login attempts per IP per minute (0 = off) |
| `VIEW_RATE_LIMIT` | `30` | Max view tracking requests per IP per minute (0 = off) |
| `TRENDING_CACHE_TTL_SECONDS` | `300` | In-memory cache TTL for trending query in seconds |
| `SCRAPE_MAX_PER_SOURCE` | `5` | Maximum newest articles scraped per source per cycle |
| `SCRAPE_MAX_INSERT_PER_SOURCE` | `5` | Hard cap on new drafts stored per source per cycle |
| `SCRAPE_TIMEOUT_SECONDS` | `15` | Per-article scraper timeout |
| `SCRAPE_MIN_CONTENT_CHARS` | `500` | Minimum article character count required to keep draft |
| `SCRAPE_DELAY_MIN_SECONDS` | `1` | Minimum politeness delay before external requests |
| `SCRAPE_DELAY_MAX_SECONDS` | `2` | Maximum politeness delay before external requests |
| `HTTP_COMPRESSION_ENABLED` | `true` | Enable gzip/brotli payload compression |
| `LOG_LEVEL` | `info` | Minimum log level: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `text` | Structured log output format: `text` or `json` |

### Frontend Configuration (`frontend/.env.local`)

| Variable | Description |
|---|---|
| `PUBLIC_API_URL` | API base URL (dev: `http://localhost:8080/api`; **production: leave unset** to use relative `/api`) |
| `PUBLIC_SITE_URL` | Canonical site origin (e.g. `https://neuralwire.info`) used for sitemaps, OpenGraph, and Twitter cards |

---

## Development Workflow

Neuralwire enforces the **Feature Branching Strategy (GitHub Flow)** off `main`:

```
main (protected production branch, deploys on merge)
  │
  ├── branch off main: feat/*, fix/*, docs/*, refactor/*, security/*, ci/*
  │     (develop, run local verification suite, test localhost E2E)
  │
  └── open Pull Request -> main (CI must pass -> review -> merge -> deploy)
```

### 1. Mandatory Local Verification Protocol
Before proposing any commit or PR, both suites MUST pass with Exit Code 0:

```bash
# Backend Verification (gofmt check, go vet, test suite)
cd backend && make verify

# Frontend Verification (prettier format, eslint, svelte-check, static build)
cd frontend && npm run verify
```

### 2. Live Localhost Smoke Test
1. Boot the server locally: `cd backend && go run ./cmd/server`.
2. Perform smoke testing against `http://localhost:8080/api/healthz`, `/api/categories`, and `/api/news`.
3. Confirm 0 panics, 0 unexpected 4xx/5xx HTTP errors, and clean shutdown.

### 3. Cryptographic GPG Commit Signing (Strict Invariant)
All commits MUST be cryptographically signed with GPG (`commit.gpgsign=true`):
```bash
git commit -S -m "type(scope): clear description"
git log -n 1 --show-signature
```

---

## CI/CD

The repository includes automated GitHub Actions workflows in `.github/workflows/`:

- **`backend-ci.yml`**:
  - Concurrency group `backend-ci-${{ github.ref }}` (`cancel-in-progress: true`).
  - Runs `go vet`, `go build`, `gofmt -l .`, race test suite (`go test -race`), and security audit via `govulncheck` (`GOTOOLCHAIN=auto`).
  - Uploads compiled `neuralwire-server` binary artifact.
- **`frontend-ci.yml`**:
  - Concurrency group `frontend-ci-${{ github.ref }}` (`cancel-in-progress: true`).
  - Runs `npm ci`, Prettier + ESLint checks, `svelte-check` typecheck, static production build, and `npm audit --audit-level=high`.
- **`deploy.yml`**:
  - Serialized concurrency group `deploy-vps` (`cancel-in-progress: false`).
  - Triggers on semantic version tag push (`v*.*.*`) or manual dispatch (`workflow_dispatch`) → SSHs into production VPS → checks out target ref and builds container with Docker Compose.
  - **Automated Post-Deploy Health Check**: Executes a 30-second polling retry loop (5s interval) validating container state (`docker compose ps`) and `http://127.0.0.1:8080/api/healthz`. Dumps 100 lines of container logs and exits with code 1 if health check fails.

---

## Deployment & Operations

### Production Architecture (VPS + Docker + Caddy)

The production deployment runs on a Linux VPS behind Caddy and Cloudflare:

```bash
# Manual VPS deployment / restart
cd /home/deploy/neuralwire
git pull origin main
docker compose up -d --build
```

### Volume Persistence & File Permissions
Persistent data is mounted at `/app/data` (mapped to volume `neuralwire-data`). The runtime container runs as non-root user `neuralwire` (`UID 10001`). The entrypoint script (`docker-entrypoint.sh`) handles permission ownership automatically on startup before dropping privileges.

### Release Tagging & Rollback
Releases are tagged using semantic versioning:
```bash
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

**Rollback Procedures:**
1. **Via Git Revert (Standard)**: Revert bad commit on a `fix/` branch, open PR, merge, and let CI/CD deploy.
2. **Via Docker Compose on VPS (Emergency)**: Check out the previous release tag on the VPS and run `docker compose up -d --build`.

---

## Security

- **Constant-Time Auth**: Bearer tokens are signed with HMAC-SHA256 and validated with `crypto/subtle.ConstantTimeCompare`.
- **CSRF Defense**: Origin header validation enforced on all state-changing admin mutation endpoints.
- **SSRF Defense**: Outbound scrapers and feed testers route through `netutil.SafeHTTPClient` which blocks private, loopback, link-local, and multicast IP ranges.
- **DoS & Memory Protection**: Request bodies are restricted via `http.MaxBytesReader`.
- **Hardened Production Gate**: Server refuses to boot with default credentials or development secrets when `APP_ENV=production`.
- **Security Headers**: Injects `Content-Security-Policy`, `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Cross-Origin-Opener-Policy`, `Cross-Origin-Resource-Policy`, `Referrer-Policy`, and `Permissions-Policy`.

---

## Project Structure

```
.
├── .github/workflows/        # CI/CD pipelines (backend-ci, frontend-ci, deploy)
├── backend/
│   ├── cmd/server/           # Go main entrypoint
│   ├── internal/
│   │   ├── ai/               # AI summarizer, categorizer, news-value scoring
│   │   ├── api/              # HTTP server, handlers, middleware, routes, preloading
│   │   ├── auth/             # HMAC token issue & validation
│   │   ├── cache/            # TTL in-memory cache for trending queries
│   │   ├── config/           # Environment variable loading & defaults
│   │   ├── database/         # SQLite initialization, migrations, seeds
│   │   ├── fetcher/          # RSS ingestion pipeline & politeness throttler
│   │   ├── metrics/          # Prometheus runtime collectors
│   │   ├── models/           # Domain data models & request/response schemas
│   │   ├── netutil/          # SSRF-safe dialer & HTTP clients
│   │   ├── ratelimit/        # Per-IP sliding-window rate limiters
│   │   ├── repository/       # Database queries (News, Categories, Sources, Settings)
│   │   ├── scheduler/        # Autopublish cron timer & background worker
│   │   ├── scoring/          # Heuristic & AI weighted news-value scorer
│   │   ├── scraper/          # Readability content extraction & image upgrade
│   │   └── slug/             # SEO URL slug generator
│   ├── Makefile              # Single-command verification & tasks
│   └── .env.example          # Backend configuration template
├── frontend/
│   ├── src/
│   │   ├── lib/              # API client, TypeScript models, Svelte components
│   │   └── routes/           # SvelteKit pages (Portal, [slug], Search, Admin suite)
│   ├── static/               # Favicon and static assets
│   ├── package.json          # Frontend dependencies & verification scripts
│   └── .env.example          # Frontend configuration template
├── Dockerfile                # Multi-stage production container build
├── docker-compose.yml        # Container orchestration specification
├── docker-entrypoint.sh      # Volume permission initialization & non-root drop
├── AGENTS.md                 # Persistent agent guidelines & verification protocols
├── CONTRIBUTING.md           # Developer contribution guidelines
└── LICENSE                   # MIT License
```

---

## API Reference

All endpoints return JSON responses unless otherwise noted. List endpoints follow the standard envelope `{ "data": [...], "pagination": {...} }`.

### Public Endpoints

| Method | Path | Description | Auth |
|---|---|---|---|
| `GET` | `/api/health` | Liveness health check (DB ping) | Public |
| `GET` | `/api/healthz` | Health check alias for load balancers & CI probes | Public |
| `GET` | `/api/metrics` | Prometheus metrics counters | Public |
| `GET` | `/sitemap.xml` | Dynamic XML sitemap of all published articles | Public |
| `GET` | `/robots.txt` | Dynamic robots exclusion file | Public |
| `GET` | `/api/news` | List published articles (`?category=`, `?q=`, `?page=`, `?page_size=`) | Public |
| `GET` | `/api/news/{id}` | Single published article by ID or slug | Public |
| `GET` | `/api/news/trending` | Most-read published articles (`?window=day\|week\|all`, `?limit=`) | Public |
| `GET` | `/api/news/{id}/related` | Related articles ranked by keyword & category similarity | Public |
| `POST` | `/api/news/{id}/view` | Record an article read (`{ "viewer_key": "..." }`) | Public |
| `GET` | `/api/categories` | List all available categories | Public |
| `POST` | `/api/admin/login` | Authenticate admin credentials and receive bearer token | Public |

### Admin Moderation Endpoints

All admin endpoints below require `Authorization: Bearer <token>` and enforce CSRF origin protection on mutations.

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/admin/news` | List articles across any status (`?status=draft\|published\|rejected`, `?value_label=`) |
| `GET` | `/api/admin/news/{id}` | Get full article details including extracted content |
| `POST` | `/api/admin/news` | Create a new article manually |
| `PUT` | `/api/admin/news/{id}` | Update article title, summary, category, image URL, or status |
| `POST` | `/api/admin/news/{id}/publish` | Publish a draft article |
| `POST` | `/api/admin/news/{id}/reject` | Move a draft or article to rejected archive |
| `POST` | `/api/admin/news/{id}/set-primary` | Set article as primary/featured story |
| `POST` | `/api/admin/news/bulk` | Bulk action (`publish`, `reject`, `delete`) on an array of IDs |
| `DELETE` | `/api/admin/news/{id}` | Permanently delete an article |
| `DELETE` | `/api/admin/news` | Bulk delete all articles with a specific status (`?status=rejected`) |
| `POST` | `/api/admin/fetch` | Manually trigger an RSS ingestion cycle across all active sources |
| `GET` | `/api/admin/fetch/progress` | Live progress of running fetch cycle (`{ "running": true, "percent": 50 }`) |
| `POST` | `/api/admin/fetch/cancel` | Abort an in-flight fetch cycle |
| `GET` | `/api/admin/settings` | Get value scoring thresholds (`low_max`, `medium_min`, `medium_max`, `high_min`) |
| `PUT` | `/api/admin/settings` | Update value scoring thresholds in database |
| `GET` | `/api/admin/autopublish` | Get autopublish scheduler configuration and running status |
| `PUT` | `/api/admin/autopublish` | Update autopublish scheduler settings |
| `POST` | `/api/admin/autopublish/start` | Start the automated background ingestion & publishing scheduler |
| `POST` | `/api/admin/autopublish/stop` | Stop the automated background scheduler |
| `POST` | `/api/admin/upload-image` | Upload custom article cover image (JPEG, PNG, WebP, GIF $\le$ 5MB) |
| `GET` | `/api/admin/backup` | Download an immediate gzip snapshot of the SQLite database |
| `GET` | `/api/admin/sources` | List all configured RSS feeds with active status and category |
| `POST` | `/api/admin/sources` | Add a new RSS feed source |
| `PUT` | `/api/admin/sources/{id}` | Update an existing RSS feed source |
| `PATCH` | `/api/admin/sources/{id}/toggle` | Toggle RSS source enabled/disabled state |
| `DELETE` | `/api/admin/sources/{id}` | Delete an RSS feed source |
| `POST` | `/api/admin/sources/test` | Validate and preview an RSS feed URL before adding |
| `GET` | `/api/admin/analytics` | Fetch analytics summary (views, categories, drafts, published, top articles) |

---

## License

MIT License — see [LICENSE](LICENSE).

© 2026 NEURALWIRE MEDIA. All rights reserved for curated content; source code is licensed under MIT.
