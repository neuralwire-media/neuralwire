package repository

import (
	"database/sql"
	"fmt"
	"time"

	"neuralwire/backend/internal/models"
)

// RSSSourceRepository persists configured RSS feeds.
type RSSSourceRepository struct {
	db *sql.DB
}

// NewRSSSourceRepository creates an RSSSourceRepository.
func NewRSSSourceRepository(db *sql.DB) *RSSSourceRepository {
	return &RSSSourceRepository{db: db}
}

// ListEnabled returns all feeds that should be polled.
func (r *RSSSourceRepository) ListEnabled() ([]models.RSSSource, error) {
	rows, err := r.db.Query(`
		SELECT id, name, url, category, enabled, last_fetched_at,
		       COALESCE(last_duration_ms, 0), COALESCE(last_http_status, 0),
		       COALESCE(last_error, ''), COALESCE(consecutive_failures, 0),
		       COALESCE(total_items_yielded, 0), created_at
		FROM rss_sources
		WHERE enabled = 1
		ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list rss sources: %w", err)
	}
	defer rows.Close()

	sources := []models.RSSSource{}
	for rows.Next() {
		var s models.RSSSource
		var enabled int
		var lastFetched sql.NullString
		var createdAt string
		if err := rows.Scan(
			&s.ID, &s.Name, &s.URL, &s.Category, &enabled, &lastFetched,
			&s.LastDurationMs, &s.LastHTTPStatus, &s.LastError,
			&s.ConsecutiveFailures, &s.TotalItemsYielded, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan rss source: %w", err)
		}
		s.Enabled = enabled == 1
		if lastFetched.Valid {
			t, err := parseSQLiteTime(lastFetched.String)
			if err != nil {
				return nil, fmt.Errorf("parse last_fetched_at: %w", err)
			}
			s.LastFetchedAt = &t
		}
		t, err := parseSQLiteTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse rss source created_at: %w", err)
		}
		s.CreatedAt = t
		sources = append(sources, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rss sources: %w", err)
	}
	return sources, nil
}

// UpdateLastFetched records the time a source was last polled.
func (r *RSSSourceRepository) UpdateLastFetched(id int64, t time.Time) error {
	if _, err := r.db.Exec(
		`UPDATE rss_sources SET last_fetched_at = ? WHERE id = ?`,
		t.UTC().Format(time.RFC3339), id,
	); err != nil {
		return fmt.Errorf("update last_fetched: %w", err)
	}
	return nil
}

// RecordFetchResult updates a source's polling diagnostics.
func (r *RSSSourceRepository) RecordFetchResult(id int64, t time.Time, durationMs int64, httpStatus int, itemsYielded int, fetchErr string) error {
	var errStr string
	if fetchErr != "" {
		errStr = fetchErr
	}
	if fetchErr != "" || (httpStatus != 0 && httpStatus >= 400) {
		_, err := r.db.Exec(`
			UPDATE rss_sources
			SET last_fetched_at = ?,
			    last_duration_ms = ?,
			    last_http_status = ?,
			    last_error = ?,
			    consecutive_failures = consecutive_failures + 1,
			    total_items_yielded = total_items_yielded + ?
			WHERE id = ?`,
			t.UTC().Format(time.RFC3339), durationMs, httpStatus, errStr, itemsYielded, id,
		)
		if err != nil {
			return fmt.Errorf("record fetch failure: %w", err)
		}
	} else {
		_, err := r.db.Exec(`
			UPDATE rss_sources
			SET last_fetched_at = ?,
			    last_duration_ms = ?,
			    last_http_status = ?,
			    last_error = '',
			    consecutive_failures = 0,
			    total_items_yielded = total_items_yielded + ?
			WHERE id = ?`,
			t.UTC().Format(time.RFC3339), durationMs, httpStatus, itemsYielded, id,
		)
		if err != nil {
			return fmt.Errorf("record fetch success: %w", err)
		}
	}
	return nil
}

// ListAll returns all configured RSS sources.
func (r *RSSSourceRepository) ListAll() ([]models.RSSSource, error) {
	rows, err := r.db.Query(`
		SELECT id, name, url, category, enabled, last_fetched_at,
		       COALESCE(last_duration_ms, 0), COALESCE(last_http_status, 0),
		       COALESCE(last_error, ''), COALESCE(consecutive_failures, 0),
		       COALESCE(total_items_yielded, 0), created_at
		FROM rss_sources
		ORDER BY category ASC, name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all rss sources: %w", err)
	}
	defer rows.Close()

	sources := []models.RSSSource{}
	for rows.Next() {
		var s models.RSSSource
		var enabled int
		var lastFetched sql.NullString
		var createdAt string
		if err := rows.Scan(
			&s.ID, &s.Name, &s.URL, &s.Category, &enabled, &lastFetched,
			&s.LastDurationMs, &s.LastHTTPStatus, &s.LastError,
			&s.ConsecutiveFailures, &s.TotalItemsYielded, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan rss source: %w", err)
		}
		s.Enabled = enabled == 1
		if lastFetched.Valid {
			t, err := parseSQLiteTime(lastFetched.String)
			if err != nil {
				return nil, fmt.Errorf("parse last_fetched_at: %w", err)
			}
			s.LastFetchedAt = &t
		}
		t, err := parseSQLiteTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse rss source created_at: %w", err)
		}
		s.CreatedAt = t
		sources = append(sources, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rss sources: %w", err)
	}
	return sources, nil
}

// GetByID retrieves a single RSS source by its ID.
func (r *RSSSourceRepository) GetByID(id int64) (*models.RSSSource, error) {
	var s models.RSSSource
	var enabled int
	var lastFetched sql.NullString
	var createdAt string

	err := r.db.QueryRow(`
		SELECT id, name, url, category, enabled, last_fetched_at,
		       COALESCE(last_duration_ms, 0), COALESCE(last_http_status, 0),
		       COALESCE(last_error, ''), COALESCE(consecutive_failures, 0),
		       COALESCE(total_items_yielded, 0), created_at
		FROM rss_sources
		WHERE id = ?`, id,
	).Scan(
		&s.ID, &s.Name, &s.URL, &s.Category, &enabled, &lastFetched,
		&s.LastDurationMs, &s.LastHTTPStatus, &s.LastError,
		&s.ConsecutiveFailures, &s.TotalItemsYielded, &createdAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get rss source: %w", err)
	}

	s.Enabled = enabled == 1
	if lastFetched.Valid {
		t, err := parseSQLiteTime(lastFetched.String)
		if err != nil {
			return nil, fmt.Errorf("parse last_fetched_at: %w", err)
		}
		s.LastFetchedAt = &t
	}
	t, err := parseSQLiteTime(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse rss source created_at: %w", err)
	}
	s.CreatedAt = t
	return &s, nil
}

// Create inserts a new RSS source into the database.
func (r *RSSSourceRepository) Create(s models.RSSSource) (int64, error) {
	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	res, err := r.db.Exec(`
		INSERT INTO rss_sources (name, url, category, enabled, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		s.Name, s.URL, s.Category, enabled,
	)
	if err != nil {
		return 0, fmt.Errorf("create rss source: %w", err)
	}
	return res.LastInsertId()
}

// Update modifies an existing RSS source.
func (r *RSSSourceRepository) Update(s models.RSSSource) error {
	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	res, err := r.db.Exec(`
		UPDATE rss_sources
		SET name = ?, url = ?, category = ?, enabled = ?
		WHERE id = ?`,
		s.Name, s.URL, s.Category, enabled, s.ID,
	)
	if err != nil {
		return fmt.Errorf("update rss source: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("rss source not found")
	}
	return nil
}

// ToggleEnabled toggles the enabled state of an RSS source.
func (r *RSSSourceRepository) ToggleEnabled(id int64, enabled bool) error {
	val := 0
	if enabled {
		val = 1
	}
	res, err := r.db.Exec(`UPDATE rss_sources SET enabled = ? WHERE id = ?`, val, id)
	if err != nil {
		return fmt.Errorf("toggle rss source enabled: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("rss source not found")
	}
	return nil
}

// Delete removes an RSS source by ID.
func (r *RSSSourceRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM rss_sources WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete rss source: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("rss source not found")
	}
	return nil
}
