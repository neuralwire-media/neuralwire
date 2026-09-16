# Neuralwire Frontend

SvelteKit frontend for the **Neuralwire** AI news website (global audience, English by default). Renders the public editorial portal and the admin moderation suite.

---

## Architecture & Tech Stack

- **Framework**: **SvelteKit 2** running **Svelte 5** in Runes mode (`$state`, `$derived`, `$props`, `$effect`).
- **Styling**: **Tailwind CSS 4** (`@tailwindcss/vite`) with custom editorial dark theme.
- **Build Adapter**: **`@sveltejs/adapter-static`** generates a prerendered static client SPA served directly by the Go backend in production.
- **Type Safety**: Full TypeScript integration checked via `svelte-check`.
- **Bundler**: Vite with instant HMR and optimized asset hashing.

---

## Core Capabilities

### 1. Public Editorial Portal

- **Homepage News Grid (`/`)**:
  - Hero featured article with automatic LCP preloading.
  - Category navigation tabs (AI, Machine Learning, Research, Tools, Industry).
  - Reactive **"Load More"** pagination (displays 15 articles initially, appends 15 more per click).
  - Floating **"Collapse Feed"** FAB to reset feed view and smoothly scroll back to top.
  - **Trending / Most-Read Carousel (`TrendingNews.svelte`)**: Highlights top-5 most read articles for the week.
- **Article Detail Page (`/[slug]`)**:
  - Displays curated AI digest summary and metadata.
  - Links out to the original publisher ("Read Full Story") — preserving source copyright and fair use.
  - Automatic fire-and-forget view recording (`POST /api/news/{id}/view`) with browser-level deduplication (`nw_viewer_id`).
  - Related stories section using TF-IDF weighted similarity matching.
- **Category Feeds (`/category/[slug]`)**:
  - Dedicated pages for all 5 core categories with responsive 5-column grid layout (`2xl:grid-cols-5`).
- **Backend-Powered Search (`/search`)**:
  - Real-time search with ~300ms debounce querying backend full-text search with multi-word AND matching.
- **Static & Legal Pages**:
  - `/about`: Editorial platform overview and curation principles.
  - `/copyright`: Fair use statement and publisher attribution notice.

### 2. Admin Management Suite (`/admin`)

- **Dashboard Overview (`/admin`)**:
  - Real-time pipeline counters (draft, published, rejected).
  - Manual fetch trigger with live progress bar and `[CANCEL FETCH]` capability.
  - Autopublish scheduler configuration (intervals, category filters, score threshold filters) with independent Start/Stop controls.
  - Value scoring threshold manager (`score_low_max`, `score_medium_min`, `score_medium_max`, `score_high_min`).
  - On-demand database backup download.
- **Drafts Moderation (`/admin/drafts`)**:
  - Advisory AI value score badges (`HIGH`, `MEDIUM`, `LOW`).
  - Detailed sub-score breakdown (Impact, Novelty, Quality, Heuristics, Confidence, and AI reasoning).
  - Single-item and bulk actions: Bulk Publish, Bulk Reject, and Bulk Delete.
- **Published Management (`/admin/published`)**:
  - Search and filter live articles, view individual read counts, and toggle primary/featured status.
- **Rejected Archive (`/admin/rejected`)**:
  - Archive of rejected drafts with restore and permanent delete options.
- **RSS Sources Manager (`/admin/sources`)**:
  - Full CRUD interface for RSS feeds.
  - Live feed tester to validate and preview RSS URLs before saving.
  - Instant enable/disable toggle.
- **Analytics Dashboard (`/admin/analytics`)**:
  - Comprehensive metrics: total page views, category view shares, publication velocity, and top-read articles.
- **Article Preview & Editor (`/admin/preview/[id]`)**:
  - Metadata and content editor with custom cover image upload (`POST /api/admin/upload-image`) and instant live preview.

---

## Environment Variables

Configure in `frontend/.env.local`:

| Variable          | Description                                                                                                              |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `PUBLIC_API_URL`  | Base URL for backend API (development: `http://localhost:8080/api`; **production: leave unset** to use relative `/api`)  |
| `PUBLIC_SITE_URL` | Canonical site URL (e.g. `https://neuralwire.info`) used for generating canonical tags, OpenGraph, and Twitter Card URLs |

---

## Development & Verification

### Running the Dev Server

```bash
cd frontend
cp .env.example .env.local

# Start development server on http://localhost:5173
npm run dev

# Or open browser automatically
npm run dev -- --open
```

### Production Build & Preview

```bash
# Compile static SPA build into build/ directory
npm run build

# Preview the static build locally
npm run preview
```

### Mandatory Verification Suite

Before pushing changes or creating a PR, run the single-command verification suite:

```bash
# Runs format check, ESLint, svelte-check, and static production build
npm run verify
```

---

## Project Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── api.ts              # REST API client & fetch helpers
│   │   ├── components/         # Shared UI components (Trending, Progress, Image)
│   │   ├── stores.ts           # Svelte stores & reactive state
│   │   └── types.ts            # TypeScript interfaces & models
│   └── routes/
│       ├── +error.svelte       # Custom 404/500 error page
│       ├── +layout.svelte      # Root layout, navigation header, and footer
│       ├── +page.svelte        # Homepage news feed & portal
│       ├── [slug]/             # Article detail page
│       ├── category/[slug]/    # Category news feed
│       ├── search/             # Debounced search page
│       ├── about/              # About editorial platform
│       ├── copyright/          # Copyright & fair-use notice
│       └── admin/              # Admin moderation suite
│           ├── +layout.svelte  # Admin shell & navigation
│           ├── +page.svelte    # Admin dashboard & scheduler controls
│           ├── login/          # Bearer auth login page
│           ├── drafts/         # Draft review & value scoring
│           ├── published/      # Published articles management
│           ├── rejected/       # Rejected articles archive
│           ├── sources/        # RSS feed source manager
│           ├── analytics/      # Analytics dashboard
│           └── preview/[id]/   # Article editor & image upload
├── static/                     # Favicon, robots.txt, and static assets
├── package.json                # NPM dependencies & verification scripts
├── svelte.config.js            # SvelteKit configuration (adapter-static)
├── tsconfig.json               # TypeScript configuration
└── vite.config.ts              # Vite & Tailwind CSS 4 configuration
```
