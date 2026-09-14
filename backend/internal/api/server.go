// Package api exposes the Neuralwire REST API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"neuralwire/backend/internal/auth"
	"neuralwire/backend/internal/cache"
	"neuralwire/backend/internal/fetcher"
	"neuralwire/backend/internal/metrics"
	"neuralwire/backend/internal/models"
	"neuralwire/backend/internal/ratelimit"
	"neuralwire/backend/internal/repository"
	"neuralwire/backend/internal/scheduler"
)

// FeedFetcher runs one RSS fetch cycle and reports its progress. It is
// satisfied by *fetcher.Fetcher and by test doubles.
type FeedFetcher interface {
	FetchAll(ctx context.Context) (fetcher.FetchStats, error)
	Progress() fetcher.FetchProgress
}

// Server wires repositories and middleware into an http.Handler.
type Server struct {
	newsRepo     *repository.NewsRepository
	categoryRepo *repository.CategoryRepository
	settingsRepo *repository.SettingsRepository
	allowOrigins []string
	auth         *auth.Manager
	adminUser    string
	adminPass    string
	fetcher      FeedFetcher
	logger       *log.Logger
	slog         *slog.Logger

	// fetchMu guards the in-flight manual fetch cancel function so the admin
	// cancel endpoint can abort a running cycle without data races.
	fetchMu     sync.Mutex
	fetchCancel context.CancelFunc
	// viewLimiter is a per-IP rate limit for POST /api/news/{id}/view to
	// prevent view-count abuse. Nil disables limiting.
	viewLimiter *ratelimit.Limiter
	// loginLimiter is a per-IP rate limit for POST /api/admin/login to slow
	// brute-force password attempts. Nil disables limiting.
	loginLimiter *ratelimit.Limiter
	// globalLimiter is a per-IP rate limit for every request (anti-scan /
	// anti-bot). Nil disables limiting.
	globalLimiter *ratelimit.Limiter
	// trustProxy enables trusting X-Forwarded-For for client IP resolution.
	trustProxy bool
	// trendingCache memoizes trending results per window for a few minutes so
	// the heavy GROUP BY query is not re-run on every request. Nil disables.
	trendingCache *cache.Cache
	// disableCompression turns off gzip/brotli response compression.
	disableCompression bool
	// metrics collects runtime counters exposed via GET /api/metrics.
	metrics *metrics.Metrics
	// staticDir is the frontend build directory served at "/" when present.
	staticDir string
	// uploadDir is where admin-uploaded images are stored, served at
	// /uploads/.
	uploadDir string
	// scheduler controls the auto fetch/publish loop (STY-57/61). Nil
	// disables the start/stop API.
	scheduler *scheduler.Scheduler
	// backupDir is where admin-created database backups are stored.
	backupDir string
	// backupRetain is how many backups to keep (0 = no pruning).
	backupRetain int
}

// ServerOptions configures the API server.
type ServerOptions struct {
	NewsRepo     *repository.NewsRepository
	CategoryRepo *repository.CategoryRepository
	SettingsRepo *repository.SettingsRepository
	AllowOrigins []string
	Auth         *auth.Manager
	AdminUser    string
	AdminPass    string
	// Fetcher drives the manual POST /api/admin/fetch endpoint. When nil the
	// endpoint responds 503.
	Fetcher FeedFetcher
	// ViewRateLimit is the max view-count requests per IP per window
	// (default 30 per minute when >0). <=0 disables the limiter.
	ViewRateLimit  int
	ViewRateWindow time.Duration
	// TrendingCacheTTL caches trending results for this duration (default 5m
	// when >0). <=0 disables trending caching.
	TrendingCacheTTL time.Duration
	// TrustProxy enables trusting X-Forwarded-For for client IP resolution.
	// Enable only behind a trusted reverse proxy.
	TrustProxy bool
	// LoginRateLimit is the max login attempts per IP per window
	// (default 5 per minute when >0). <=0 disables login limiting.
	LoginRateLimit  int
	LoginRateWindow time.Duration
	// GlobalRateLimit is the max requests per IP per window applied to every
	// request (default 120 per minute when >0). <=0 disables global limiting.
	GlobalRateLimit  int
	GlobalRateWindow time.Duration
	// DisableCompression turns off gzip/brotli response compression.
	DisableCompression bool
	Logger             *log.Logger
	// Slog is the structured logger used for request/error logging. When nil,
	// a discard logger is used so request logging is a no-op.
	Slog *slog.Logger
	// Metrics collects runtime counters exposed via GET /api/metrics. When
	// nil, a fresh collector is used so the endpoint always responds.
	Metrics *metrics.Metrics
	// StaticDir is the frontend build directory served at "/" (SPA fallback
	// to index.html). Empty disables static serving.
	StaticDir string
	// UploadDir is where admin-uploaded images are stored. Empty disables
	// the upload endpoint and /uploads/ serving.
	UploadDir string
	// Scheduler is the auto fetch/publish controller. When non-nil, the
	// /api/admin/autopublish/start and /stop endpoints toggle it.
	Scheduler *scheduler.Scheduler
	// BackupDir is where database backups are stored. Empty disables the
	// backup endpoint.
	BackupDir string
	// BackupRetain keeps the newest N backups and prunes the rest.
	BackupRetain int
}

// NewServer builds a Server.
func NewServer(opts ServerOptions) *Server {
	if len(opts.AllowOrigins) == 0 {
		opts.AllowOrigins = []string{"http://localhost:5173"}
	}
	if opts.Auth == nil {
		opts.Auth = auth.NewManager("", 0)
	}
	if opts.Logger == nil {
		opts.Logger = log.Default()
	}
	if opts.Slog == nil {
		opts.Slog = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if opts.Metrics == nil {
		opts.Metrics = metrics.New()
	}
	if opts.ViewRateWindow <= 0 {
		opts.ViewRateWindow = time.Minute
	}
	srv := &Server{
		newsRepo:           opts.NewsRepo,
		categoryRepo:       opts.CategoryRepo,
		settingsRepo:       opts.SettingsRepo,
		allowOrigins:       opts.AllowOrigins,
		auth:               opts.Auth,
		adminUser:          opts.AdminUser,
		adminPass:          opts.AdminPass,
		fetcher:            opts.Fetcher,
		logger:             opts.Logger,
		slog:               opts.Slog,
		trustProxy:         opts.TrustProxy,
		disableCompression: opts.DisableCompression,
		metrics:            opts.Metrics,
		staticDir:          opts.StaticDir,
		uploadDir:          opts.UploadDir,
		scheduler:          opts.Scheduler,
		backupDir:          opts.BackupDir,
		backupRetain:       opts.BackupRetain,
	}
	if opts.ViewRateLimit > 0 {
		srv.viewLimiter = ratelimit.New(opts.ViewRateLimit, opts.ViewRateWindow)
		srv.viewLimiter.Start()
	}
	if opts.LoginRateLimit > 0 {
		if opts.LoginRateWindow <= 0 {
			opts.LoginRateWindow = time.Minute
		}
		srv.loginLimiter = ratelimit.New(opts.LoginRateLimit, opts.LoginRateWindow)
		srv.loginLimiter.Start()
	}
	if opts.GlobalRateLimit > 0 {
		if opts.GlobalRateWindow <= 0 {
			opts.GlobalRateWindow = time.Minute
		}
		srv.globalLimiter = ratelimit.New(opts.GlobalRateLimit, opts.GlobalRateWindow)
		srv.globalLimiter.Start()
	}
	if opts.TrendingCacheTTL > 0 {
		srv.trendingCache = cache.New(opts.TrendingCacheTTL)
		srv.trendingCache.Start()
	}
	return srv
}

// Handler assembles the route table and middleware stack. All routes under
// /api/admin/ require a bearer token except POST /api/admin/login.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/healthz", s.handleHealth)
	mux.HandleFunc("GET /api/metrics", s.handleMetrics)
	mux.HandleFunc("GET /sitemap.xml", s.handleSitemap)
	mux.HandleFunc("GET /robots.txt", s.handleRobotsTXT)
	mux.HandleFunc("GET /api/news", s.handleListNews)
	mux.HandleFunc("GET /api/news/{id}", s.handleGetNews)
	mux.HandleFunc("GET /api/news/trending", s.handleTrendingNews)
	mux.HandleFunc("GET /api/news/{id}/related", s.handleRelatedNews)
	mux.HandleFunc("POST /api/news/{id}/view", s.handleRecordView)
	mux.HandleFunc("GET /api/categories", s.handleListCategories)

	// Admin login is public; every other /api/admin/* route is protected.
	mux.HandleFunc("POST /api/admin/login", s.handleLogin)

	admin := http.NewServeMux()
	admin.HandleFunc("GET /api/admin/news", s.handleAdminListNews)
	admin.HandleFunc("GET /api/admin/news/{id}", s.handleAdminGetNews)
	admin.HandleFunc("POST /api/admin/news", s.handleCreateNews)
	admin.HandleFunc("PUT /api/admin/news/{id}", s.handleUpdateNews)
	admin.HandleFunc("POST /api/admin/news/{id}/publish", s.handlePublishNews)
	admin.HandleFunc("POST /api/admin/news/{id}/reject", s.handleRejectNews)
	admin.HandleFunc("DELETE /api/admin/news/{id}", s.handleDeleteNews)
	admin.HandleFunc("DELETE /api/admin/news", s.handleDeleteNewsByStatus)
	mux.Handle("POST /api/admin/fetch", s.requireAuth(s.csrfProtect(http.HandlerFunc(s.handleFetchNews))))
	mux.Handle("POST /api/admin/fetch/cancel", s.requireAuth(s.csrfProtect(http.HandlerFunc(s.handleCancelFetch))))
	mux.Handle("GET /api/admin/fetch/progress", s.requireAuth(http.HandlerFunc(s.handleFetchProgress)))
	mux.Handle("GET /api/admin/settings", s.requireAuth(http.HandlerFunc(s.handleGetSettings)))
	mux.Handle("PUT /api/admin/settings", s.requireAuth(s.csrfProtect(http.HandlerFunc(s.handleUpdateSettings))))
	mux.Handle("GET /api/admin/autopublish", s.requireAuth(http.HandlerFunc(s.handleGetAutoPublish)))
	mux.Handle("PUT /api/admin/autopublish", s.requireAuth(s.csrfProtect(http.HandlerFunc(s.handleUpdateAutoPublish))))
	mux.Handle("POST /api/admin/autopublish/start", s.requireAuth(s.csrfProtect(http.HandlerFunc(s.handleStartAutoPublish))))
	mux.Handle("POST /api/admin/autopublish/stop", s.requireAuth(s.csrfProtect(http.HandlerFunc(s.handleStopAutoPublish))))
	mux.Handle("POST /api/admin/upload-image", s.requireAuth(s.csrfProtect(http.HandlerFunc(s.handleUploadImage))))
	mux.Handle("GET /api/admin/backup", s.requireAuth(http.HandlerFunc(s.handleBackup)))
	mux.Handle("/api/admin/", s.requireAuth(s.csrfProtect(admin)))

	// Serve admin-uploaded images under /uploads/ when an upload directory
	// is configured.
	if s.uploadDir != "" {
		mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.uploadDir))))
	}

	// Serve the built frontend when a static directory is configured. The
	// SPA fallback returns index.html for valid frontend routes so client
	// side routing works (e.g. /some-article-slug). Unknown or non-existent
	// routes return HTTP 404 to avoid Soft 404 indexation.
	if s.staticDir != "" {
		if info, err := os.Stat(s.staticDir); err == nil && info.IsDir() {
			fileServer := http.FileServer(http.Dir(s.staticDir))
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				path := r.URL.Path
				if path == "/" || path == "/index.html" {
					s.serveIndexHTML(w, r)
					return
				}
				// If the asset exists on disk, serve it directly.
				candidate := http.Dir(s.staticDir)
				if _, err := candidate.Open(strings.TrimPrefix(path, "/")); err != nil {
					cleanPath := strings.Trim(r.URL.Path, "/")
					if s.isValidFrontendRoute(cleanPath) {
						s.serveIndexHTML(w, r)
						return
					}
					s.serveNotFound(w, r)
					return
				}
				fileServer.ServeHTTP(w, r)
			})
		} else {
			s.slog.Warn("api: static dir not found; frontend not served",
				"static_dir", s.staticDir,
				"error", err,
			)
		}
	}

	// Middleware order: recover -> rateLimit -> securityHeaders -> log -> cors
	// -> cacheControl -> etag -> compress. ETag runs outside compression so the
	// hash covers the compressed bytes the client actually received; cache
	// headers are set last so conditional revalidation can serve 304s.
	return s.recover(s.rateLimit(s.securityHeaders(s.log(s.cors(
		s.cacheControl(s.etag(s.compress(mux))),
	)))))
}

// --- response helpers -----------------------------------------------------

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.slog.Error("api: encode response", "error", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

// isValidFrontendRoute checks whether the requested path corresponds to a
// known frontend route (e.g. root, about, search, copyright, admin, category,
// or a published article slug). If not, the request should receive a real 404
// instead of index.html to prevent Soft 404 indexing.
func (s *Server) isValidFrontendRoute(cleanPath string) bool {
	switch cleanPath {
	case "", "about", "copyright", "search":
		return true
	}

	if cleanPath == "admin" || strings.HasPrefix(cleanPath, "admin/") {
		return true
	}

	if strings.HasPrefix(cleanPath, "category/") {
		catSlug := strings.TrimPrefix(cleanPath, "category/")
		if catSlug == "" || strings.Contains(catSlug, "/") {
			return false
		}
		if s.categoryRepo != nil {
			exists, err := s.categoryRepo.ExistsBySlug(catSlug)
			if err != nil || !exists {
				return false
			}
		}
		return true
	}

	// Dynamic article route: /{slug}
	if !strings.Contains(cleanPath, "/") {
		if s.newsRepo != nil {
			news, err := s.newsRepo.GetBySlug(cleanPath)
			if err != nil || news == nil || news.Status != models.StatusPublished {
				return false
			}
		}
		return true
	}

	return false
}

// serveNotFound responds with HTTP 404 and a clean error page or JSON.
func (s *Server) serveNotFound(w http.ResponseWriter, r *http.Request) {
	cleanPath := strings.Trim(r.URL.Path, "/")
	if strings.HasPrefix(cleanPath, "api/") || cleanPath == "api" {
		s.writeError(w, http.StatusNotFound, "endpoint not found")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="robots" content="noindex, nofollow">
  <title>404 Not Found — Neuralwire</title>
  <style>
    body { background: #0A0E17; color: #F8FAFC; font-family: ui-sans-serif, system-ui, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; }
    .container { text-align: center; padding: 2rem; max-width: 28rem; }
    h1 { font-size: 3.75rem; font-weight: 700; color: #22D3EE; margin: 0 0 0.5rem; font-family: monospace; }
    h2 { font-size: 1.25rem; font-weight: 500; margin: 0 0 1rem; color: #E2E8F0; }
    p { color: #94A3B8; font-size: 0.875rem; margin-bottom: 2rem; line-height: 1.5; }
    a { display: inline-block; background: rgba(34,211,238,0.1); color: #22D3EE; border: 1px solid rgba(34,211,238,0.3); padding: 0.625rem 1.25rem; border-radius: 0.5rem; text-decoration: none; font-size: 0.875rem; font-weight: 500; transition: all 0.2s; }
    a:hover { background: rgba(34,211,238,0.2); border-color: #22D3EE; }
  </style>
</head>
<body>
  <div class="container">
    <h1>404</h1>
    <h2>Page Not Found</h2>
    <p>The page or article you are looking for does not exist, has been removed, or is not yet published.</p>
    <a href="/">Return to Homepage</a>
  </div>
</body>
</html>`))
}

// serveIndexHTML serves the SPA entrypoint index.html, dynamically injecting
// API fetch preloads, image preconnects, and an LCP cover image preload link in <head>.
func (s *Server) serveIndexHTML(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(s.staticDir, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	cleanPath := strings.Trim(r.URL.Path, "/")
	var extraPreloads strings.Builder
	extraPreloads.WriteString("\t\t<link rel=\"preload\" as=\"fetch\" href=\"/api/categories\" crossorigin>\n")
	if cleanPath == "" || cleanPath == "index.html" {
		extraPreloads.WriteString("\t\t<link rel=\"preload\" as=\"fetch\" href=\"/api/news?page_size=15\" crossorigin>\n")
		extraPreloads.WriteString("\t\t<link rel=\"preload\" as=\"fetch\" href=\"/api/news/trending?window=week&limit=5\" crossorigin>\n")
	}

	var preloadImage string
	if s.newsRepo != nil {
		if cleanPath == "" || cleanPath == "index.html" {
			// Homepage: preload latest published article cover image
			articles, _, err := s.newsRepo.ListPublished("", "", 1, 1)
			if err == nil && len(articles) > 0 && articles[0].ImageURL != "" {
				preloadImage = articles[0].ImageURL
			}
		} else if !strings.HasPrefix(cleanPath, "category/") && cleanPath != "about" && cleanPath != "copyright" && cleanPath != "search" {
			// Single article page: preload that article's cover image and inject article SEO tags
			article, err := s.newsRepo.GetBySlug(cleanPath)
			if err == nil && article != nil {
				if article.ImageURL != "" {
					preloadImage = article.ImageURL
				}
				articleTitle := html.EscapeString(article.Title) + " | Neuralwire"
				articleDesc := html.EscapeString(article.Summary)
				content = bytes.Replace(content, []byte("<title>Neuralwire | AI News, Neural Networks &amp; Future Computation</title>"), []byte("<title>"+articleTitle+"</title>"), 1)
				content = bytes.Replace(content, []byte(`content="Curated intelligence on frontier AI research, neural networks, machine learning, and computational industry."`), []byte(`content="`+articleDesc+`"`), -1)
			}
		}
	}

	if preloadImage != "" {
		if u, err := url.Parse(preloadImage); err == nil && u.Scheme != "" && u.Host != "" {
			imageOrigin := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			extraPreloads.WriteString(fmt.Sprintf("\t\t<link rel=\"preconnect\" href=\"%s\">\n", html.EscapeString(imageOrigin)))
		}
		if strings.Contains(preloadImage, "platform.theverge.com") && strings.Contains(preloadImage, "w=") {
			re := regexp.MustCompile(`([?&]w=)\d+`)
			u400 := re.ReplaceAllString(preloadImage, "${1}400")
			u800 := re.ReplaceAllString(preloadImage, "${1}800")
			u1200 := re.ReplaceAllString(preloadImage, "${1}1200")
			srcSet := fmt.Sprintf("%s 400w, %s 800w, %s 1200w", html.EscapeString(u400), html.EscapeString(u800), html.EscapeString(u1200))
			extraPreloads.WriteString(fmt.Sprintf("\t\t<link rel=\"preload\" as=\"image\" href=\"%s\" imagesrcset=\"%s\" imagesizes=\"(max-width: 640px) 400px, (max-width: 1024px) 800px, 1200px\" fetchpriority=\"high\">\n", html.EscapeString(preloadImage), srcSet))
		} else {
			extraPreloads.WriteString(fmt.Sprintf("\t\t<link rel=\"preload\" as=\"image\" href=\"%s\" fetchpriority=\"high\">\n", html.EscapeString(preloadImage)))
		}
	}

	if extraPreloads.Len() > 0 {
		content = bytes.Replace(content, []byte("</head>"), []byte(extraPreloads.String()+"\t</head>"), 1)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}
