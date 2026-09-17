package ai

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ImageGenerator handles generating cover images using AI or stock fallbacks.
type ImageGenerator interface {
	// Generate returns a cover image URL based on the title and category.
	Generate(ctx context.Context, title, category string) string
}

type openAIImageGenerator struct {
	apiKey        string
	imageEndpoint string
	enabled       bool
	// unsupported is set once the upstream reports it cannot generate images
	// (404/405/501), so later calls skip straight to the stock fallback.
	unsupported bool
	client      *http.Client
	logger      *slog.Logger
}

// NewImageGenerator creates an ImageGenerator client. When enabled is false,
// generation is skipped entirely and only the stock fallback is used.
func NewImageGenerator(apiKey, baseURL string, enabled bool, logger *slog.Logger) ImageGenerator {
	if logger == nil {
		logger = slog.Default()
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/images/generations"
	return &openAIImageGenerator{
		apiKey:        apiKey,
		imageEndpoint: endpoint,
		enabled:       enabled,
		client:        &http.Client{Timeout: 30 * time.Second},
		logger:        logger,
	}
}

type imageReq struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
	Model  string `json:"model"`
}

type imageResp struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func (g *openAIImageGenerator) Generate(ctx context.Context, title, category string) string {
	// Skip AI generation when disabled by config or after the provider
	// reported it does not support image generation.
	if !g.enabled || g.unsupported {
		return GetCuratedTechImage(category, title)
	}

	apiKey, baseURL, _ := resolveAIConfig(g.apiKey, "", g.imageEndpoint)

	imageEndpoint := g.imageEndpoint
	if baseURL != "" {
		imageEndpoint = strings.TrimRight(baseURL, "/") + "/images/generations"
	}

	// 1. Try DALL-E generation if API key is configured
	if apiKey != "" {
		prompt := fmt.Sprintf(
			"A high-quality, professional, modern minimalist technology featured cover image illustrating: %s. Dark cybernetic tech aesthetic.",
			title,
		)
		url, unsupported, ok := generateImage(ctx, g.client, g.logger, imageEndpoint, apiKey, prompt)
		if unsupported {
			g.unsupported = true
			g.logger.Warn("ai: image generation not supported by provider; skipping future attempts")
		}
		if ok {
			return url
		}
	}

	// 2. Fallback to curated, high-resolution technology photo URL distributed stably by title hash.
	return GetCuratedTechImage(category, title)
}

// extractKeywords extracts clean, meaningful keywords from the article title.
func extractKeywords(title string) []string {
	words := strings.Fields(strings.ToLower(title))
	var filtered []string
	stopwords := map[string]bool{
		"how": true, "what": true, "why": true, "with": true, "from": true,
		"this": true, "that": true, "your": true, "over": true, "under": true,
		"their": true, "about": true, "using": true, "built": true, "build": true,
		"and": true, "the": true, "for": true, "our": true, "new": true, "more": true,
	}
	for _, w := range words {
		w = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, w)
		if len(w) > 2 && !stopwords[w] {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

// GetDynamicUnsplashURL returns an empty string because Unsplash stock image pool was replaced by dynamic OG generator.
func GetDynamicUnsplashURL(category, title string) string {
	return ""
}

// GetDynamicOGURL returns a bespoke Open Graph image URL from https://og.neuralwire.info/api/og.
func GetDynamicOGURL(category, title, source string, score int) string {
	base := "https://og.neuralwire.info/api/og"
	q := url.Values{}
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle != "" {
		q.Set("title", cleanTitle)
	}
	cleanCat := strings.TrimSpace(category)
	if cleanCat != "" {
		q.Set("category", cleanCat)
	}
	cleanSource := strings.TrimSpace(source)
	if cleanSource != "" {
		q.Set("source", cleanSource)
	}
	if score > 0 {
		q.Set("score", strconv.Itoa(score))
	}
	return base + "?" + q.Encode()
}

// GetCuratedTechImage returns a dynamic Open Graph image URL from https://og.neuralwire.info/api/og
// matching the title and category.
func GetCuratedTechImage(category, title string) string {
	return GetDynamicOGURL(category, title, "", 0)
}
