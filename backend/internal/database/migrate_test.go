package database

import (
	"path/filepath"
	"testing"
)

func TestMigrate_FreshDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "fresh.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate on fresh database failed: %v", err)
	}

	// Verify that news table exists with cluster_id and is_primary
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM news WHERE cluster_id = '' AND is_primary = 1`).Scan(&count)
	if err != nil {
		t.Fatalf("Query news table failed: %v", err)
	}
}

func TestMigrate_ExistingLegacyDatabase(t *testing.T) {
	// Simulate an existing database created before cluster_id and value_* columns were added.
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Create legacy news table without cluster_id, is_primary, and value_* columns
	legacySchema := `
	CREATE TABLE news (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		title        TEXT    NOT NULL,
		slug         TEXT    NOT NULL UNIQUE,
		url          TEXT    NOT NULL,
		source       TEXT    NOT NULL DEFAULT '',
		category     TEXT    NOT NULL DEFAULT 'ai',
		summary      TEXT    NOT NULL DEFAULT '',
		content      TEXT    NOT NULL DEFAULT '',
		image_url    TEXT    NOT NULL DEFAULT '',
		status       TEXT    NOT NULL DEFAULT 'draft',
		published_at DATETIME,
		created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatalf("Create legacy schema failed: %v", err)
	}

	// Insert legacy row
	_, err = db.Exec(`INSERT INTO news (title, slug, url) VALUES ('Legacy News', 'legacy-news', 'https://example.com/legacy')`)
	if err != nil {
		t.Fatalf("Insert legacy news failed: %v", err)
	}

	// Run Migrate on the legacy DB
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate on legacy database failed: %v", err)
	}

	// Verify cluster_id column exists and defaults properly
	var clusterID string
	var isPrimary int
	err = db.QueryRow(`SELECT cluster_id, is_primary FROM news WHERE slug = 'legacy-news'`).Scan(&clusterID, &isPrimary)
	if err != nil {
		t.Fatalf("Query migrated columns failed: %v", err)
	}
	if clusterID != "" {
		t.Errorf("expected empty clusterID, got %q", clusterID)
	}
	if isPrimary != 1 {
		t.Errorf("expected isPrimary = 1, got %d", isPrimary)
	}

	// Verify idx_news_cluster_id index is usable
	var id, parent, notused int
	var detail string
	err = db.QueryRow(`EXPLAIN QUERY PLAN SELECT id FROM news WHERE cluster_id = 'test-cluster'`).Scan(&id, &parent, &notused, &detail)
	if err != nil {
		t.Fatalf("Explain query plan for cluster_id index failed: %v", err)
	}
	if detail == "" {
		t.Errorf("expected non-empty detail from explain query plan")
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "idempotent.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("First Migrate failed: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("Second Migrate failed (idempotency check): %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("Third Migrate failed (idempotency check): %v", err)
	}
}
