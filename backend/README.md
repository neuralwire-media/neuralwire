# Neuralwire Backend

Go REST API backend for the **Neuralwire** AI news website (global audience, English by default).

## Features

- **Standard Library Go 1.25+ REST API** on port `8080` (Go 1.22+ routing patterns, zero heavy web frameworks).
- **SQLite Storage**: Pure-Go driver (`modernc.org/sqlite`), zero CGO required, running in WAL mode with robust database migrations.
- **25 Default Curated RSS Sources**: Pre-seeded across 5 core categories (AI, Tools, Research, Industry, Machine Learning).
- **Dynamic RSS Source Management**: Add, update, toggle, delete, and test feed URLs in real time via `/api/admin/sources`.
- **Curator Model**: The scraper reads full article text only as material for the AI summary and value scoring, then discards it. Original full text is **never stored or republished**, keeping the site compliant with fair use and directing traffic to original publishers.
- **AI Summaries & Categorization**: Compatible with any OpenAI-compliant API (OpenAI, Gemini, DeepSeek, Groq, Ollama) with robust deterministic fallbacks.
- **Level 2 AI Value Scoring**: Every draft is rated 0–100 by a weighted blend of AI judgment ($60\%$) and deterministic heuristics ($40\%$), labelled HIGH/MEDIUM/LOW with sub-score breakdowns, confidence metrics, and advisory reasons. Scoring is advisory only and never auto-publishes without explicit scheduler configuration.
- **Admin-Configurable Scoring Thresholds**: Persisted in database `app_settings` (defaults: LOW <60, MEDIUM 60–79, HIGH $\ge$80) via `GET/PUT /api/admin/settings`.
- **Automated Ingestion & Autopublish Scheduler**: Optional timer scheduler configured via `GET/PUT /api/admin/autopublish` with category whitelists, score label filters, max posts per cycle, and independent start/stop controllers (`/api/admin/autopublish/start` and `/stop`).
- **Live Fetch Control**: Real-time progress tracking (`GET /api/admin/fetch/progress`) and in-flight fetch cancellation (`POST /api/admin/fetch/cancel`).
- **Cover Image Upgrades & Admin Uploads**: Low-res thumbnails from known CDNs (Contentful, imgix, Cloudinary, Unsplash, Google, WordPress) are automatically upgraded to high-resolution variants. Admins can also upload custom cover images (`POST /api/admin/upload-image`, $\le 5$ MiB) stored under `UPLOAD_DIR` and served at `/uploads/`.
- **Trending / Most-Read Ranking**: Public `POST /api/news/{id}/view` records reads (deduplicated per visitor via `viewer_key` + 6-hour cooldown) and `GET /api/news/trending?window=day|week|all&limit=N` returns top articles with view counts.
- **Dynamic Sitemap & Robots.txt**: `/sitemap.xml` and `/robots.txt` are served dynamically from published database records.
- **Analytics Endpoint**: `GET /api/admin/analytics` aggregates total views, published counts, draft pipelines, category distributions, and top articles.
- **Automated Database Backups**: Gzip snapshot on boot and every `BACKUP_INTERVAL_HOURS` (default 24h) under `BACKUP_DIR` with `BACKUP_RETENTION` (default 7). Admins can also download on-demand snapshots via `GET /api/admin/backup`.
- **Security & Hardening**:
  - Constant-time HMAC-SHA256 bearer token authentication.
  - CSRF origin validation on all state-changing admin mutations.
  - SSRF protection via `netutil.SafeHTTPClient` blocking loopback, private, link-local, and multicast IP ranges.
  - Security headers on every response: CSP, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Cross-Origin-Opener-Policy`, `Cross-Origin-Resource-Policy`, `Referrer-Policy`, and `Permissions-Policy`.
  - Rate limiting: Login brute-force limiter, view abuse limiter, and global anti-scan limiter.
  - Memory DoS prevention: Request bodies limited via `http.MaxBytesReader`.
- **Structured Logging & Observability**: Go `log/slog` structured logging (JSON or text) with log-level filtering. Prometheus metrics exposed at `/api/metrics` (requests, status codes, latencies, AI calls, fetch stats) and `/api/healthz` health checks.
- **Response Compression & ETag Revalidation**: Gzip/brotli compression for text/JSON payloads and ETag conditional revalidation (304 Not Modified).

---

## Requirements

- **Go 1.25+** (newer toolchains are downloaded automatically when using `GOTOOLCHAIN=auto`).

---

## Quick Start

```bash
cd backend
cp .env.example .env

# Build and start server
make run
# or directly:
go run ./cmd/server
```

The server listens on `http://localhost:8080` and initializes `data/neuralwire.db` automatically on first run.

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `APP_ENV` | `development` | Runtime environment (`production` enforces security hardening) |
| `TRUST_PROXY` | `false` | Trust `X-Forwarded-For` header (enable only behind trusted reverse proxies like Caddy/Nginx) |
| `DB_PATH` | `data/neuralwire.db` | SQLite database file location |
| `USER_AGENT` | `Mozilla/5.0 (compatible; NeuralwireBot/1.0-dev; +https://neuralwire.example)` | User-Agent header sent on outbound requests |
| `STATIC_DIR` | `../frontend/build` | Static frontend directory served at `/` (SPA fallback to `index.html`) |
| `UPLOAD_DIR` | `./data/uploads` | Storage directory for admin-uploaded images (served at `/uploads/`) |
| `BACKUP_DIR` | `./data/backups` | Storage directory for automated database gzip snapshots |
| `BACKUP_RETENTION` | `7` | Number of backup snapshots retained |
| `BACKUP_INTERVAL_HOURS` | `24` | Backup frequency in hours (0 disables automated backup) |
| `CORS_ALLOW_ORIGIN` | `http://localhost:5173,http://127.0.0.1:5173` | Comma-separated allowed CORS origins |
| `ADMIN_USERNAME` | `admin` | Admin username |
| `ADMIN_PASSWORD` | `admin123` | Admin password (**must be changed in production**) |
| `ADMIN_TOKEN_SECRET` | *(dev secret)* | HMAC key for admin bearer tokens (**must be changed in production**) |
| `AI_SUMMARY_API_KEY` | *(empty)* | OpenAI-compatible API key for summarization & scoring |
| `AI_SUMMARY_PROVIDER` | `openai` | Provider preset: `openai`, `gemini`, `openrouter`, `groq`, `ollama` |
| `AI_SUMMARY_BASE_URL` | `https://api.openai.com/v1` | OpenAI-compatible base URL |
| `AI_SUMMARY_MODEL` | `gpt-4o-mini` | AI model used for summaries and value scoring |
| `AI_IMAGE_GENERATION_ENABLED` | *(auto)* | Enable/disable AI cover image generation |
| `SCRAPE_MAX_PER_SOURCE` | `5` | Maximum newest articles scraped per source per cycle |
| `SCRAPE_MAX_INSERT_PER_SOURCE` | `5` | Maximum new drafts stored per source per cycle |
| `SCRAPE_TIMEOUT_SECONDS` | `15` | Per-article scraper timeout |
| `SCRAPE_MIN_CONTENT_CHARS` | `500` | Minimum article character count required to keep draft |
| `SCRAPE_DELAY_MIN_SECONDS` | `1` | Minimum politeness delay before external requests |
| `SCRAPE_DELAY_MAX_SECONDS` | `2` | Maximum politeness delay before external requests |
| `VIEW_RATE_LIMIT` | `30` | Max view tracking requests per IP per minute (0 = disabled) |
| `TRENDING_CACHE_TTL_SECONDS` | `300` | Memory cache TTL for trending query in seconds |
| `LOGIN_RATE_LIMIT` | `5` | Max login attempts per IP per minute (0 = disabled) |
| `GLOBAL_RATE_LIMIT` | `120` | Max requests per IP per minute for all endpoints (0 = disabled) |
| `HTTP_COMPRESSION_ENABLED` | `true` | Enable gzip/brotli response compression |
| `LOG_LEVEL` | `info` | Minimum log level: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `text` | Structured log output format: `text` or `json` |

---

## How Ingestion Works (Curator Model)

By default, there is **no background scheduler running** unless explicitly configured — fetching happens when an authenticated admin calls `POST /api/admin/fetch`:

1. **Polite Crawling**: Waits a random 1–2 second delay before outbound requests to avoid rate limits on source servers.
2. **Readability Scraping**: Extracts clean article body text using `codeberg.org/readeck/go-readability/v2`, stripping ads, navigations, and widgets.
3. **Curator Extraction**: Scraped text is used **only** to generate summaries, classifications, and value scores. Original full text is **discarded** and never stored.
4. **Quality Gate**: Articles with content shorter than `SCRAPE_MIN_CONTENT_CHARS` (500 chars) are dropped to filter out stub articles.
5. **Budgets & Caps**: Maximum 5 articles scraped and maximum 5 drafts inserted per source per cycle.
6. **Image Enhancement**: First usable article image is extracted and CDN URLs are upgraded to high-resolution variants.
7. **Value Scoring**: Evaluated via `scoring.ScoreService` ($0.6 \times \text{AI score} + 0.4 \times \text{Heuristic score}$) and assigned HIGH/MEDIUM/LOW badges based on configurable thresholds.

---

## Autopublish & Ingestion Scheduler

The background scheduler can be configured via `GET/PUT /api/admin/autopublish`:

```json
{
  "enabled": true,
  "auto_post_enabled": true,
  "interval_minutes": 360,
  "categories": ["ai", "machine-learning"],
  "min_score_labels": ["medium", "high"],
  "max_posts_per_cycle": 5
}
```

- **Scheduler Control Endpoints**:
  - `POST /api/admin/autopublish/start` — Starts the timer scheduler.
  - `POST /api/admin/autopublish/stop` — Stops the timer scheduler.
  - `GET /api/admin/autopublish` — Returns configuration and live running status.

---

## API Reference

All endpoints return JSON responses unless specified otherwise.

### Public Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/health` | Health check (DB ping) |
| `GET` | `/api/healthz` | Health check alias |
| `GET` | `/api/metrics` | Prometheus metrics counters |
| `GET` | `/sitemap.xml` | Dynamic XML sitemap of all published articles |
| `GET` | `/robots.txt` | Dynamic robots exclusion rules |
| `GET` | `/api/news` | List published articles (`?category=`, `?q=`, `?page=`, `?page_size=`) |
| `GET` | `/api/news/{id}` | Single published article by ID or slug |
| `GET` | `/api/news/trending` | Most-read published articles (`?window=day\|week\|all`, `?limit=5`) |
| `GET` | `/api/news/{id}/related` | Related articles ranked by similarity |
| `POST` | `/api/news/{id}/view` | Record an article read (`{ "viewer_key": "..." }`) |
| `GET` | `/api/categories` | List all available categories |
| `POST` | `/api/admin/login` | Login with username and password, returns bearer token |

### Admin Moderation Endpoints

All admin endpoints require `Authorization: Bearer <token>` and enforce CSRF protection on mutations.

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/admin/news` | List articles across any status (`?status=draft\|published\|rejected`, `?value_label=`) |
| `GET` | `/api/admin/news/{id}` | Get full article details |
| `POST` | `/api/admin/news` | Create a new article manually |
| `PUT` | `/api/admin/news/{id}` | Update article details |
| `POST` | `/api/admin/news/{id}/publish` | Publish a draft article |
| `POST` | `/api/admin/news/{id}/reject` | Move an article to rejected archive |
| `POST` | `/api/admin/news/{id}/set-primary` | Toggle article as primary/featured story |
| `POST` | `/api/admin/news/bulk` | Bulk action (`publish`, `reject`, `delete`) on an array of IDs |
| `DELETE` | `/api/admin/news/{id}` | Delete an article permanently |
| `DELETE` | `/api/admin/news` | Bulk delete all articles with a specific status (`?status=rejected`) |
| `POST` | `/api/admin/fetch` | Manually trigger an RSS ingestion cycle |
| `GET` | `/api/admin/fetch/progress` | Live progress of running fetch cycle |
| `POST` | `/api/admin/fetch/cancel` | Abort an in-flight fetch cycle |
| `GET` | `/api/admin/settings` | Get current value scoring thresholds |
| `PUT` | `/api/admin/settings` | Update value scoring thresholds |
| `GET` | `/api/admin/autopublish` | Get autopublish configuration and running state |
| `PUT` | `/api/admin/autopublish` | Update autopublish configuration |
| `POST` | `/api/admin/autopublish/start` | Start the autopublish background scheduler |
| `POST` | `/api/admin/autopublish/stop` | Stop the autopublish background scheduler |
| `POST` | `/api/admin/upload-image` | Upload custom cover image (multipart form, $\le 5$ MiB) |
| `GET` | `/api/admin/backup` | Download SQLite database gzip snapshot |
| `GET` | `/api/admin/sources` | List all configured RSS feeds |
| `POST` | `/api/admin/sources` | Add a new RSS feed source |
| `PUT` | `/api/admin/sources/{id}` | Update an existing RSS feed source |
| `PATCH` | `/api/admin/sources/{id}/toggle` | Toggle RSS source enabled/disabled state |
| `DELETE` | `/api/admin/sources/{id}` | Delete an RSS feed source |
| `POST` | `/api/admin/sources/test` | Test and preview an RSS feed URL |
| `GET` | `/api/admin/analytics` | Summary analytics for dashboard |

---

## Verification & Development

To verify the backend code before committing:

```bash
# Single-command verification shortcut:
make verify

# Or manual step-by-step sequence:
# 1. Format check
gofmt -w . && gofmt -l .

# 2. Static analysis
go vet ./...

# 3. Test suite
go test -count=1 -race -v ./...

# 4. Vulnerability audit
GOTOOLCHAIN=auto go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```
