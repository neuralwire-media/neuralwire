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
