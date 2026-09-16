package database

import (
	"path/filepath"
	"testing"
)

func TestSeed_FreshDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "seed_fresh.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	if err := Seed(db); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	// Verify categories
	var catCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&catCount); err != nil {
		t.Fatalf("Query categories count failed: %v", err)
	}
	if catCount != len(defaultCategories) {
		t.Errorf("categories count = %d, want %d", catCount, len(defaultCategories))
	}

	// Verify sources
	var srcCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rss_sources`).Scan(&srcCount); err != nil {
		t.Fatalf("Query rss_sources count failed: %v", err)
	}
	if srcCount != len(defaultRSSSources) {
		t.Errorf("rss_sources count = %d, want %d", srcCount, len(defaultRSSSources))
	}

	// Verify initial_seed_done flag
	var flag string
	if err := db.QueryRow(`SELECT value FROM app_settings WHERE key = 'initial_seed_done'`).Scan(&flag); err != nil {
		t.Fatalf("Query initial_seed_done failed: %v", err)
	}
	if flag != "1" {
		t.Errorf("initial_seed_done = %q, want '1'", flag)
	}
}

func TestSeed_DoesNotResurrectDeletedSources(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "seed_deleted.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("Initial Seed failed: %v", err)
	}

	// Simulate admin deleting 2 default sources
	deletedURL1 := defaultRSSSources[0].url
	deletedURL2 := defaultRSSSources[1].url
	if _, err := db.Exec(`DELETE FROM rss_sources WHERE url IN (?, ?)`, deletedURL1, deletedURL2); err != nil {
		t.Fatalf("Delete sources failed: %v", err)
	}

	// Verify count is now len(defaultRSSSources) - 2
	var currentCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rss_sources`).Scan(&currentCount); err != nil {
		t.Fatalf("Query count failed: %v", err)
	}
	if currentCount != len(defaultRSSSources)-2 {
		t.Fatalf("expected %d sources after deletion, got %d", len(defaultRSSSources)-2, currentCount)
	}

	// Simulate server restart / redeploy: Seed() is called again
	if err := Seed(db); err != nil {
		t.Fatalf("Second Seed call failed: %v", err)
	}

	// Verify that deleted sources were NOT resurrected
	var afterRestartCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rss_sources`).Scan(&afterRestartCount); err != nil {
		t.Fatalf("Query count after restart failed: %v", err)
	}
	if afterRestartCount != len(defaultRSSSources)-2 {
		t.Errorf("Seed resurrected deleted sources! Count = %d, want %d", afterRestartCount, len(defaultRSSSources)-2)
	}

	var existsCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rss_sources WHERE url IN (?, ?)`, deletedURL1, deletedURL2).Scan(&existsCount); err != nil {
		t.Fatalf("Query deleted URLs failed: %v", err)
	}
	if existsCount != 0 {
		t.Errorf("expected 0 deleted sources, found %d", existsCount)
	}
}

func TestSeed_ExistingDatabaseWithSettings(t *testing.T) {
	// Simulate existing production DB that has settings but never had initial_seed_done flag
	dbPath := filepath.Join(t.TempDir(), "seed_legacy.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Insert only 10 sources and settings
	for i := 0; i < 10; i++ {
		s := defaultRSSSources[i]
		_, _ = db.Exec(`INSERT INTO rss_sources (name, url, category) VALUES (?, ?, ?)`, s.name, s.url, s.category)
	}
	for k, v := range defaultSettings {
		_, _ = db.Exec(`INSERT INTO app_settings (key, value) VALUES (?, ?)`, k, v)
	}

	// Call Seed: should detect existing settings, set initial_seed_done = '1', and NOT insert the other 15 sources
	if err := Seed(db); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	var srcCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rss_sources`).Scan(&srcCount); err != nil {
		t.Fatalf("Query count failed: %v", err)
	}
	if srcCount != 10 {
		t.Errorf("sources count = %d, want 10 (should not re-insert the other 15 sources)", srcCount)
	}

	var flag string
	if err := db.QueryRow(`SELECT value FROM app_settings WHERE key = 'initial_seed_done'`).Scan(&flag); err != nil {
		t.Fatalf("Query flag failed: %v", err)
	}
	if flag != "1" {
		t.Errorf("flag = %q, want '1'", flag)
	}
}
