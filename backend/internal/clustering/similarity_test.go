package clustering

import (
	"context"
	"testing"
	"time"
)

func TestCleanTokens(t *testing.T) {
	text := "OpenAI Announces GPT-5 With Advanced Reasoning and $100M Compute!"
	tokens := CleanTokens(text)

	wantWords := map[string]bool{
		"openai": true, "announce_action": true, "gpt": true, "5": true,
		"advanc": true, "reason": true, "100m": true, "compute": true,
	}

	for _, w := range tokens {
		if !wantWords[w] {
			t.Errorf("unexpected token %q in output: %v", w, tokens)
		}
	}
}

func TestSimilaritySameStoryDifferentHeadlines(t *testing.T) {
	// TechCrunch vs The Verge vs Ars Technica covering the same event
	titleA := "OpenAI launches GPT-5 with revolutionary mathematical reasoning"
	summaryA := "OpenAI today unveiled GPT-5, featuring breakthrough reasoning capabilities in mathematics and coding benchmark tasks."

	titleB := "OpenAI announces GPT-5 model with breakthrough reasoning benchmarks"
	summaryB := "The AI startup OpenAI has released its newest GPT-5 model, demonstrating state-of-the-art coding and mathematical reasoning."

	score := Similarity(titleA, summaryA, titleB, summaryB)
	if score < DefaultSimilarityThreshold {
		t.Errorf("expected similarity >= %f for same story, got %f", DefaultSimilarityThreshold, score)
	}

	if !IsMatch(titleA, summaryA, titleB, summaryB, DefaultSimilarityThreshold) {
		t.Error("expected IsMatch to be true for identical news story from different sources")
	}
}

func TestSimilarityDifferentStories(t *testing.T) {
	titleA := "NVIDIA launches Blackwell Ultra AI GPUs for data centers"
	summaryA := "NVIDIA has announced its next-generation Blackwell Ultra chips designed for high-density enterprise AI training."

	titleB := "Apple introduces new M4 chip for iPad Pro lineup"
	summaryB := "Apple today revealed the new iPad Pro powered by the M4 processor with upgraded neural engine."

	score := Similarity(titleA, summaryA, titleB, summaryB)
	if score >= DefaultSimilarityThreshold {
		t.Errorf("expected similarity < %f for different stories, got %f", DefaultSimilarityThreshold, score)
	}

	if IsMatch(titleA, summaryA, titleB, summaryB, DefaultSimilarityThreshold) {
		t.Error("expected IsMatch to be false for distinct stories")
	}
}

type fakeClusterStore struct {
	candidates []CandidateNews
}

func (f *fakeClusterStore) FindRecentCandidates(ctx context.Context, since time.Time) ([]CandidateNews, error) {
	return f.candidates, nil
}

func TestAssignCluster(t *testing.T) {
	existingCluster := "cl_test_12345"
	store := &fakeClusterStore{
		candidates: []CandidateNews{
			{
				ID:        10,
				ClusterID: existingCluster,
				Title:     "Anthropic reveals Claude 3.7 Sonnet hybrid reasoning model",
				Summary:   "Anthropic has introduced Claude 3.7 Sonnet, blending fast response generation with extended thinking mode.",
				CreatedAt: time.Now(),
			},
		},
	}

	service := NewService(ServiceOptions{
		Store:     store,
		Window:    48 * time.Hour,
		Threshold: 0.50,
	})

	// Match candidate
	matchRes := service.AssignCluster(context.Background(),
		"Anthropic releases Claude 3.7 Sonnet with hybrid reasoning capabilities",
		"Anthropic today announced Claude 3.7 Sonnet which introduces selectable extended reasoning mode for developers.",
	)

	if !matchRes.Matched {
		t.Errorf("expected matched = true, got false (score: %f)", matchRes.Score)
	}
	if matchRes.ClusterID != existingCluster {
		t.Errorf("expected clusterID = %q, got %q", existingCluster, matchRes.ClusterID)
	}
	if matchRes.IsPrimary {
		t.Errorf("expected isPrimary = false for duplicate coverage, got true")
	}

	// Distinct story candidate
	distinctRes := service.AssignCluster(context.Background(),
		"Tesla showcases new Optimus Gen 3 humanoid robotics",
		"Tesla unveiled its next generation robotics hardware at the tech summit.",
	)

	if distinctRes.Matched {
		t.Errorf("expected matched = false for distinct story, got true")
	}
	if !distinctRes.IsPrimary {
		t.Errorf("expected isPrimary = true for new cluster, got false")
	}
	if distinctRes.ClusterID == "" || distinctRes.ClusterID == existingCluster {
		t.Errorf("expected new unique clusterID, got %q", distinctRes.ClusterID)
	}
}
