package clustering

import (
	"context"
	"time"
)

// CandidateNews represents the minimal fields needed for clustering matching.
type CandidateNews struct {
	ID        int64
	ClusterID string
	Title     string
	Summary   string
	CreatedAt time.Time
}

// ClusterStore defines repository access needed by the cluster service.
type ClusterStore interface {
	FindRecentCandidates(ctx context.Context, since time.Time) ([]CandidateNews, error)
}

// Service manages semantic grouping of articles into story clusters.
type Service struct {
	store     ClusterStore
	window    time.Duration
	threshold float64
}

// ServiceOptions configures the clustering service.
type ServiceOptions struct {
	Store     ClusterStore
	Window    time.Duration
	Threshold float64
}

// NewService builds a clustering Service.
func NewService(opts ServiceOptions) *Service {
	if opts.Window <= 0 {
		opts.Window = 48 * time.Hour // Default 48-hour comparison window
	}
	if opts.Threshold <= 0 {
		opts.Threshold = DefaultSimilarityThreshold
	}
	return &Service{
		store:     opts.Store,
		window:    opts.Window,
		threshold: opts.Threshold,
	}
}

// MatchResult is the result of attempting to match an article against existing clusters.
type MatchResult struct {
	Matched   bool
	ClusterID string
	IsPrimary bool
	Score     float64
	MatchedID int64
}

// AssignCluster evaluates candidate matches within the time window and returns
// the appropriate cluster ID and primary flag for a new article.
func (s *Service) AssignCluster(ctx context.Context, title, summary string) MatchResult {
	if s.store != nil {
		since := time.Now().Add(-s.window)
		candidates, err := s.store.FindRecentCandidates(ctx, since)
		if err == nil && len(candidates) > 0 {
			var bestScore float64
			var bestCandidate *CandidateNews

			for i := range candidates {
				c := &candidates[i]
				score := Similarity(title, summary, c.Title, c.Summary)
				if score > bestScore && score >= s.threshold {
					bestScore = score
					bestCandidate = c
				}
			}

			if bestCandidate != nil {
				targetCluster := bestCandidate.ClusterID
				if targetCluster == "" {
					targetCluster = GenerateClusterID()
				}
				return MatchResult{
					Matched:   true,
					ClusterID: targetCluster,
					IsPrimary: false,
					Score:     bestScore,
					MatchedID: bestCandidate.ID,
				}
			}
		}
	}

	// No match found -> create fresh cluster as primary
	return MatchResult{
		Matched:   false,
		ClusterID: GenerateClusterID(),
		IsPrimary: true,
		Score:     0.0,
		MatchedID: 0,
	}
}
