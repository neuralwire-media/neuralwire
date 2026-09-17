package ai

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImageGeneratorDisabledSkipsGeneration(t *testing.T) {
	// Even with an API key + base URL, disabled generator must never call
	// the upstream image endpoint and must return a dynamic OG fallback.
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	g := NewImageGenerator("test-key", srv.URL, false, slog.New(slog.NewTextHandler(io.Discard, nil)))
	got := g.Generate(context.Background(), "Some Article Title", "ai")
	if called {
		t.Error("disabled image generator should not call upstream")
	}
	if got == "" || !strings.HasPrefix(got, "https://og.neuralwire.info/api/og?") {
		t.Errorf("expected dynamic OG fallback URL, got %q", got)
	}
}

func TestImageGeneratorUnsupportedSkipsFutureCalls(t *testing.T) {
	// First call hits 404 (unsupported) -> marks generator unsupported;
	// second call must skip upstream entirely.
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	g := NewImageGenerator("test-key", srv.URL, true, slog.New(slog.NewTextHandler(io.Discard, nil)))
	g.Generate(context.Background(), "Title One", "ai")
	g.Generate(context.Background(), "Title Two", "ai")
	if calls != 1 {
		t.Errorf("upstream image calls = %d, want 1 (second call skipped)", calls)
	}
}

func TestImageGeneratorSuccessReturnsURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"url":"https://img.example.com/cover.png"}]}`))
	}))
	defer srv.Close()

	g := NewImageGenerator("test-key", srv.URL, true, slog.New(slog.NewTextHandler(io.Discard, nil)))
	got := g.Generate(context.Background(), "Title", "ai")
	if got != "https://img.example.com/cover.png" {
		t.Errorf("Generate = %q, want generated image URL", got)
	}
}

func TestGetCuratedTechImageNeverPanics(t *testing.T) {
	categories := []string{"ai", "tools", "research", "machine-learning", "industry", "default", "unknown-cat", ""}
	titles := []string{
		"",
		"A",
		"Simple Title",
		"Very long title with lots of characters and punctuation !@#$%^&*()_+-=[]{}|;':,.<>/?`~",
		"Unicode: 🤖 神经网络 深度学习 Frontier Model",
		strings.Repeat("overflow-test-string-pattern-", 100),
	}

	for _, cat := range categories {
		for _, title := range titles {
			url := GetCuratedTechImage(cat, title)
			if url == "" || !strings.HasPrefix(url, "https://og.neuralwire.info/api/og?") {
				t.Errorf("GetCuratedTechImage(%q, %q) returned invalid url: %q", cat, title, url)
			}
		}
	}
}

func TestGetDynamicOGURLParams(t *testing.T) {
	url := GetDynamicOGURL("Research", "DeepSeek V3 Model", "ArXiv", 92)
	if !strings.Contains(url, "title=DeepSeek+V3+Model") {
		t.Errorf("missing title param in url: %s", url)
	}
	if !strings.Contains(url, "category=Research") {
		t.Errorf("missing category param in url: %s", url)
	}
	if !strings.Contains(url, "source=ArXiv") {
		t.Errorf("missing source param in url: %s", url)
	}
	if !strings.Contains(url, "score=92") {
		t.Errorf("missing score param in url: %s", url)
	}

	urlWithRead := GetDynamicOGURLWithReadTime("Research", "DeepSeek V3 Model", "ArXiv", "3 min read", 92)
	if !strings.Contains(urlWithRead, "read_time=3+min+read") {
		t.Errorf("missing read_time param in url: %s", urlWithRead)
	}
}
