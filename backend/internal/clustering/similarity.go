// Package clustering provides semantic news deduplication and topic clustering
// using hybrid token Jaccard similarity, phrase n-grams, and entity matching.
package clustering

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "up": true, "about": true, "into": true,
	"over": true, "after": true, "is": true, "are": true, "was": true, "were": true,
	"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true, "will": true, "would": true, "should": true,
	"can": true, "could": true, "it": true, "its": true, "this": true, "that": true,
	"these": true, "those": true, "as": true, "how": true, "what": true, "why": true,
	"when": true, "where": true, "who": true, "which": true, "here": true, "there": true,
	"all": true, "any": true, "both": true, "each": true, "few": true, "more": true,
	"most": true, "other": true, "some": true, "such": true, "than": true, "too": true,
	"very": true, "says": true, "said": true, "new": true, "report": true, "reports": true,
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9\s]+`)

// StemWord applies a lightweight rule-based English suffix stemmer.
func StemWord(w string) string {
	if len(w) <= 3 {
		return w
	}
	// Common action synonyms in news headlines
	switch w {
	case "launches", "launched", "launching", "unveils", "unveiled", "unveiling", "reveals", "revealed", "revealing", "releases", "released", "releasing", "announces", "announced", "announcing", "introduces", "introduced", "introducing":
		return "announce_action"
	}
	suffixes := []string{
		"tions", "tion", "ments", "ment", "nesses", "ness",
		"ities", "ity", "ables", "able", "ibles", "ible",
		"fully", "ful", "ingly", "ing", "ally", "ical", "al",
		"ies", "ves", "ied", "ed", "es", "s",
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(w, suffix) && len(w)-len(suffix) >= 3 {
			return strings.TrimSuffix(w, suffix)
		}
	}
	return w
}

// CleanTokens tokenizes text into lowercase, stemmed words excluding stopwords.
func CleanTokens(text string) []string {
	text = strings.ToLower(text)
	text = nonAlphanumericRegex.ReplaceAllString(text, " ")
	words := strings.Fields(text)
	tokens := make([]string, 0, len(words))
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) == 0 || stopWords[w] {
			continue
		}
		if len(w) == 1 && !unicode.IsDigit(rune(w[0])) {
			continue
		}
		tokens = append(tokens, StemWord(w))
	}
	return tokens
}

// TokenSet returns a map set of tokens for fast intersection/union operations.
func TokenSet(tokens []string) map[string]bool {
	set := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		set[t] = true
	}
	return set
}

// WordBigrams generates consecutive word pairs (e.g. "openai releases" -> ["openai_releases"]).
func WordBigrams(tokens []string) map[string]bool {
	if len(tokens) < 2 {
		return map[string]bool{}
	}
	bigrams := make(map[string]bool, len(tokens)-1)
	for i := 0; i < len(tokens)-1; i++ {
		bigrams[tokens[i]+"_"+tokens[i+1]] = true
	}
	return bigrams
}

// JaccardSimilarity calculates the intersection over union of two token sets.
func JaccardSimilarity(setA, setB map[string]bool) float64 {
	if len(setA) == 0 || len(setB) == 0 {
		return 0.0
	}
	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

// OverlapCoefficient calculates intersection over minimum set size.
func OverlapCoefficient(setA, setB map[string]bool) float64 {
	if len(setA) == 0 || len(setB) == 0 {
		return 0.0
	}
	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}
	minLen := len(setA)
	if len(setB) < minLen {
		minLen = len(setB)
	}
	if minLen == 0 {
		return 0.0
	}
	return float64(intersection) / float64(minLen)
}

// Similarity evaluates whether two news articles cover the exact same story
// based on weighted title unigrams, title bigrams, title overlap, and summary overlap.
// Returns a similarity score between 0.0 and 1.0.
func Similarity(titleA, summaryA, titleB, summaryB string) float64 {
	tokensTitleA := CleanTokens(titleA)
	tokensTitleB := CleanTokens(titleB)

	if len(tokensTitleA) == 0 || len(tokensTitleB) == 0 {
		return 0.0
	}

	setTitleA := TokenSet(tokensTitleA)
	setTitleB := TokenSet(tokensTitleB)
	titleJaccard := JaccardSimilarity(setTitleA, setTitleB)
	titleOverlap := OverlapCoefficient(setTitleA, setTitleB)

	bigramsA := WordBigrams(tokensTitleA)
	bigramsB := WordBigrams(tokensTitleB)
	bigramJaccard := JaccardSimilarity(bigramsA, bigramsB)

	// Summary similarity
	tokensSummaryA := CleanTokens(summaryA)
	tokensSummaryB := CleanTokens(summaryB)
	summaryJaccard := 0.0
	if len(tokensSummaryA) > 0 && len(tokensSummaryB) > 0 {
		summaryJaccard = JaccardSimilarity(TokenSet(tokensSummaryA), TokenSet(tokensSummaryB))
	}

	// Weighted similarity score: Title words Jaccard (35%) + Title Overlap (35%) + Bigrams (15%) + Summary words (15%)
	score := (0.35 * titleJaccard) + (0.35 * titleOverlap) + (0.15 * bigramJaccard) + (0.15 * summaryJaccard)

	// Boost if title overlap is extremely high
	if titleOverlap >= 0.75 {
		score = mathMax(score, titleOverlap)
	}

	return score
}

// DefaultSimilarityThreshold is the minimum similarity required to group two articles
// into the same story cluster.
const DefaultSimilarityThreshold = 0.50

// IsMatch reports whether article A and B belong in the same story cluster.
func IsMatch(titleA, summaryA, titleB, summaryB string, threshold float64) bool {
	if threshold <= 0 {
		threshold = DefaultSimilarityThreshold
	}
	return Similarity(titleA, summaryA, titleB, summaryB) >= threshold
}

// GenerateClusterID returns a new random unique cluster identifier (e.g. "cl_a1b2c3d4e5f60718").
func GenerateClusterID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "cl_" + hex.EncodeToString(b)
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// IsASCII reports whether string contains only ASCII characters.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
