package repository

import (
	"context"
	"testing"
	"time"

	"neuralwire/backend/internal/database"
	"neuralwire/backend/internal/models"
)

func newTestDB(t *testing.T) *NewsRepository {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return NewNewsRepository(db)
}

func TestClusteringRepositoryOperations(t *testing.T) {
	repo := newTestDB(t)
	ctx := context.Background()

	// 1. Create two articles in the same cluster
	clusterID := "cluster-test-123"
	id1, err := repo.Create(models.News{
		Title:     "OpenAI announces GPT-5 release date",
		URL:       "https://techcrunch.com/gpt-5-announced",
		Source:    "TechCrunch",
		Category:  "ai",
		Summary:   "OpenAI today officially announced the release timeline for its GPT-5 frontier model.",
		ClusterID: clusterID,
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("create news 1: %v", err)
	}

	id2, err := repo.Create(models.News{
		Title:     "OpenAI unveils next-gen GPT-5 model launch",
		URL:       "https://theverge.com/gpt-5-launch",
		Source:    "The Verge",
		Category:  "ai",
		Summary:   "The Verge reports that OpenAI has unveiled the schedule for GPT-5.",
		ClusterID: clusterID,
		IsPrimary: false,
	})
	if err != nil {
		t.Fatalf("create news 2: %v", err)
	}

	// Publish both articles
	if err := repo.SetStatus(id1, models.StatusPublished); err != nil {
		t.Fatalf("publish 1: %v", err)
	}
	if err := repo.SetStatus(id2, models.StatusPublished); err != nil {
		t.Fatalf("publish 2: %v", err)
	}

	// 2. FindRecentCandidates
	candidates, err := repo.FindRecentCandidates(ctx, time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("find recent candidates: %v", err)
	}
	if len(candidates) != 2 {
		t.Errorf("candidate count = %d, want 2", len(candidates))
	}

	// 3. GetClusterMembers
	members, err := repo.GetClusterMembers(clusterID)
	if err != nil {
		t.Fatalf("get cluster members: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("cluster members count = %d, want 2", len(members))
	}
	if !members[0].IsPrimary || members[1].IsPrimary {
		t.Errorf("expected member 1 to be primary and member 2 to be non-primary")
	}

	// 4. AttachClusterCoverage
	n1, err := repo.GetByID(id1)
	if err != nil || n1 == nil {
		t.Fatalf("get by id 1: %v", err)
	}
	if n1.ClusterCount != 1 {
		t.Errorf("n1 cluster count = %d, want 1", n1.ClusterCount)
	}
	if len(n1.ClusterCoverage) != 1 || n1.ClusterCoverage[0].ID != id2 {
		t.Errorf("n1 coverage incorrect: got %+v", n1.ClusterCoverage)
	}

	// 5. ListPublished batch cluster counts
	published, total, err := repo.ListPublished("", "", 1, 10)
	if err != nil {
		t.Fatalf("list published: %v", err)
	}
	if total != 2 {
		t.Fatalf("total published = %d, want 2", total)
	}
	for _, p := range published {
		if p.ClusterCount != 1 {
			t.Errorf("article %d cluster count = %d, want 1", p.ID, p.ClusterCount)
		}
	}

	// 6. SetPrimaryInCluster: promote id2 to primary
	if err := repo.SetPrimaryInCluster(id2, clusterID); err != nil {
		t.Fatalf("set primary in cluster: %v", err)
	}

	n1Updated, _ := repo.GetByID(id1)
	n2Updated, _ := repo.GetByID(id2)
	if n1Updated.IsPrimary {
		t.Errorf("expected id1 to be non-primary after switch")
	}
	if !n2Updated.IsPrimary {
		t.Errorf("expected id2 to be primary after switch")
	}
}

func TestAutoPublishCandidatesCaseInsensitive(t *testing.T) {
	repo := newTestDB(t)

	// Create 3 draft articles with different value_labels
	_, err := repo.Create(models.News{
		Title:      "Article High 1",
		URL:        "https://example.com/high-1",
		Category:   "ai",
		ValueLabel: "HIGH",
	})
	if err != nil {
		t.Fatalf("create high 1: %v", err)
	}

	_, err = repo.Create(models.News{
		Title:      "Article Medium 1",
		URL:        "https://example.com/med-1",
		Category:   "tools",
		ValueLabel: "MEDIUM",
	})
	if err != nil {
		t.Fatalf("create med 1: %v", err)
	}

	_, err = repo.Create(models.News{
		Title:      "Article Low 1",
		URL:        "https://example.com/low-1",
		Category:   "research",
		ValueLabel: "LOW",
	})
	if err != nil {
		t.Fatalf("create low 1: %v", err)
	}

	// 1. Query with uppercase labels
	candidates, err := repo.AutoPublishCandidates(nil, []string{"HIGH"}, 10)
	if err != nil {
		t.Fatalf("auto publish candidates uppercase: %v", err)
	}
	if len(candidates) != 1 || candidates[0].ValueLabel != "HIGH" {
		t.Errorf("expected 1 HIGH candidate with uppercase query, got %d", len(candidates))
	}

	// 2. Query with lowercase labels
	candidates, err = repo.AutoPublishCandidates(nil, []string{"high"}, 10)
	if err != nil {
		t.Fatalf("auto publish candidates lowercase: %v", err)
	}
	if len(candidates) != 1 || candidates[0].ValueLabel != "HIGH" {
		t.Errorf("expected 1 HIGH candidate with lowercase query, got %d", len(candidates))
	}

	// 3. Query with mixed-case labels and categories
	candidates, err = repo.AutoPublishCandidates([]string{"AI", "Tools"}, []string{"High", "Medium"}, 10)
	if err != nil {
		t.Fatalf("auto publish candidates mixed: %v", err)
	}
	if len(candidates) != 2 {
		t.Errorf("expected 2 candidates with mixed query, got %d", len(candidates))
	}
}

func TestDeleteCleansUpArticleViews(t *testing.T) {
	repo := newTestDB(t)

	id, err := repo.Create(models.News{
		Title:    "Test News For Deletion",
		URL:      "https://example.com/delete-test",
		Category: "ai",
	})
	if err != nil {
		t.Fatalf("create news: %v", err)
	}

	// Record views
	if err := repo.RecordView(id, "viewer-1"); err != nil {
		t.Fatalf("record view 1: %v", err)
	}
	if err := repo.RecordView(id, "viewer-2"); err != nil {
		t.Fatalf("record view 2: %v", err)
	}

	// Verify analytics has 2 views
	analytics, err := repo.GetAnalytics(5)
	if err != nil || analytics.TotalViews != 2 {
		t.Fatalf("expected total views = 2, got %d (err: %v)", analytics.TotalViews, err)
	}

	// Delete article
	if err := repo.Delete(id); err != nil {
		t.Fatalf("delete news: %v", err)
	}

	// Verify article views table is clean
	var count int
	_ = repo.db.QueryRow(`SELECT COUNT(*) FROM article_views WHERE news_id = ?`, id).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 orphaned views after delete, got %d", count)
	}

	// Verify analytics total views is now 0
	analytics, err = repo.GetAnalytics(5)
	if err != nil || analytics.TotalViews != 0 {
		t.Errorf("expected total views = 0 after delete, got %d", analytics.TotalViews)
	}
}

func TestBulkDeleteCleansUpArticleViews(t *testing.T) {
	repo := newTestDB(t)

	id1, _ := repo.Create(models.News{Title: "News 1", URL: "https://example.com/1", Category: "ai"})
	id2, _ := repo.Create(models.News{Title: "News 2", URL: "https://example.com/2", Category: "tools"})

	_ = repo.RecordView(id1, "v1")
	_ = repo.RecordView(id2, "v2")

	processed, err := repo.BulkDelete([]int64{id1, id2})
	if err != nil || processed != 2 {
		t.Fatalf("bulk delete failed: processed %d, err: %v", processed, err)
	}

	var count int
	_ = repo.db.QueryRow(`SELECT COUNT(*) FROM article_views`).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 orphaned views after bulk delete, got %d", count)
	}
}

func TestListTrendingWindows(t *testing.T) {
	repo := newTestDB(t)

	id, err := repo.Create(models.News{
		Title:    "Trending Frontier Model",
		URL:      "https://example.com/trending-frontier",
		Category: "ai",
	})
	if err != nil {
		t.Fatalf("create news: %v", err)
	}
	if err := repo.SetStatus(id, models.StatusPublished); err != nil {
		t.Fatalf("set status: %v", err)
	}

	if err := repo.RecordView(id, "viewer-1"); err != nil {
		t.Fatalf("record view: %v", err)
	}

	for _, w := range []TrendingWindow{TrendingDay, TrendingWeek, TrendingMonth, TrendingAll} {
		trending, err := repo.ListTrending(w, 10)
		if err != nil {
			t.Fatalf("ListTrending(%s) err: %v", w, err)
		}
		if len(trending) != 1 {
			t.Fatalf("ListTrending(%s) count = %d, want 1", w, len(trending))
		}
		if trending[0].ID != id {
			t.Errorf("ListTrending(%s) id = %d, want %d", w, trending[0].ID, id)
		}
		if trending[0].ViewCount != 1 {
			t.Errorf("ListTrending(%s) view count = %d, want 1", w, trending[0].ViewCount)
		}
	}
}
