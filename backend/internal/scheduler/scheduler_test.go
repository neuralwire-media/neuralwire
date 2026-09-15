package scheduler

import (
	"context"
	"testing"
	"time"

	"neuralwire/backend/internal/database"
	"neuralwire/backend/internal/fetcher"
	"neuralwire/backend/internal/models"
	"neuralwire/backend/internal/repository"
)

func TestLabelAtLeast(t *testing.T) {
	cases := []struct {
		got, want string
		wantOK    bool
	}{
		{"high", "high", true},
		{"high", "medium", true},
		{"high", "low", true},
		{"medium", "medium", true},
		{"medium", "low", true},
		{"medium", "high", false},
		{"low", "low", true},
		{"low", "medium", false},
		{"", "low", false},
		{"unknown", "low", false},
		{"LOW", "low", true}, // case-insensitive
		{"HIGH", "medium", true},
	}
	for _, c := range cases {
		if got := labelAtLeast(c.got, c.want); got != c.wantOK {
			t.Errorf("labelAtLeast(%q, %q) = %v, want %v", c.got, c.want, got, c.wantOK)
		}
	}
}

func TestFilterMatch(t *testing.T) {
	cfg := models.AutoPublishConfig{
		Categories:    []string{"ai"},
		MinScoreLabel: "high",
	}
	news := models.News{Category: "AI", ValueLabel: "HIGH"}
	if !filterMatch(news, cfg) {
		t.Error("expected AI + HIGH to match category ai + min high")
	}
	news.Category = "Industry"
	if filterMatch(news, cfg) {
		t.Error("expected Industry to be filtered out by category whitelist")
	}
	news.Category = "AI"
	news.ValueLabel = "MEDIUM"
	if filterMatch(news, cfg) {
		t.Error("expected MEDIUM to be filtered out by min high")
	}
	// Empty whitelist = all categories.
	cfg.Categories = []string{}
	news.Category = "Tools"
	news.ValueLabel = "HIGH"
	if !filterMatch(news, cfg) {
		t.Error("expected empty category whitelist to allow any category")
	}
}

func TestFilterMatchMultiLabel(t *testing.T) {
	cfg := models.AutoPublishConfig{
		MinScoreLabels: []string{"medium", "high"},
	}
	news := models.News{ValueLabel: "HIGH"}
	if !filterMatch(news, cfg) {
		t.Error("expected HIGH to match [medium, high]")
	}
	news.ValueLabel = "MEDIUM"
	if !filterMatch(news, cfg) {
		t.Error("expected MEDIUM to match [medium, high]")
	}
	news.ValueLabel = "LOW"
	if filterMatch(news, cfg) {
		t.Error("expected LOW to be filtered out by [medium, high]")
	}
	// Multi-label takes precedence over legacy single label.
	cfg2 := models.AutoPublishConfig{
		MinScoreLabel:  "high",
		MinScoreLabels: []string{"low"},
	}
	news.ValueLabel = "MEDIUM"
	if filterMatch(news, cfg2) {
		t.Error("expected multi-label [low] to take precedence and exclude MEDIUM")
	}
	news.ValueLabel = "LOW"
	if !filterMatch(news, cfg2) {
		t.Error("expected multi-label [low] to allow LOW")
	}
}

func TestActivePersistence(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	repo := repository.NewSettingsRepository(db)

	// 1. Initially default: active should be false
	sched1 := New(repo, Job{}, nil)
	if sched1.Active() {
		t.Errorf("expected initial Active() to be false, got true")
	}

	// 2. SetActive(true) should persist
	sched1.SetActive(true)
	if !sched1.Active() {
		t.Errorf("expected Active() to be true after SetActive(true)")
	}
	if !repo.GetAutoPublishActive() {
		t.Errorf("expected repo.GetAutoPublishActive() to be true")
	}

	// 3. Simulating server restart/rebuild with same database
	sched2 := New(repo, Job{}, nil)
	if !sched2.Active() {
		t.Errorf("expected sched2.Active() to be true after restart")
	}

	// 4. SetActive(false) should persist
	sched2.SetActive(false)
	if sched2.Active() {
		t.Errorf("expected Active() to be false after SetActive(false)")
	}
	if repo.GetAutoPublishActive() {
		t.Errorf("expected repo.GetAutoPublishActive() to be false")
	}

	// 5. Simulating restart after stop
	sched3 := New(repo, Job{}, nil)
	if sched3.Active() {
		t.Errorf("expected sched3.Active() to be false after stop and restart")
	}
}

func TestActiveFallbackToConfigEnabled(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	repo := repository.NewSettingsRepository(db)

	// User saved config with Enabled=true before active key was recorded
	if err := repo.SetAutoPublishConfig(models.AutoPublishConfig{Enabled: true}); err != nil {
		t.Fatalf("set config: %v", err)
	}

	sched := New(repo, Job{}, nil)
	if !sched.Active() {
		t.Errorf("expected Active() to be true via fallback to config.Enabled")
	}
}

func TestNoImmediateExecutionOnStartup(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	repo := repository.NewSettingsRepository(db)
	if err := repo.SetAutoPublishConfig(models.AutoPublishConfig{
		Enabled:             true,
		AutoPostEnabled:     true,
		IntervalMinutes:     60,
		PostIntervalMinutes: 60,
	}); err != nil {
		t.Fatalf("set config: %v", err)
	}
	if err := repo.SetAutoPublishActive(true); err != nil {
		t.Fatalf("set active: %v", err)
	}

	fetchExecuted := false
	postExecuted := false

	sched := New(repo, Job{
		Fetch: func(ctx context.Context) (fetcher.FetchStats, error) {
			fetchExecuted = true
			return fetcher.FetchStats{}, nil
		},
		AutoPost: func(ctx context.Context, cfg models.AutoPublishConfig) (int, error) {
			postExecuted = true
			return 0, nil
		},
	}, nil)

	sched.Start()
	time.Sleep(50 * time.Millisecond)
	sched.Stop()

	if fetchExecuted {
		t.Errorf("expected fetch to not execute immediately on startup")
	}
	if postExecuted {
		t.Errorf("expected auto post to not execute immediately on startup")
	}
}

