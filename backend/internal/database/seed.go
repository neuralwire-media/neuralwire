package database

import "database/sql"

// defaultCategories is seeded on first boot.
var defaultCategories = []struct {
	name string
	slug string
}{
	{name: "AI", slug: "ai"},
	{name: "Machine Learning", slug: "machine-learning"},
	{name: "Research", slug: "research"},
	{name: "Tools", slug: "tools"},
	{name: "Industry", slug: "industry"},
}

// defaultRSSSources is seeded on first boot. The fetcher polls every
// enabled source and inserts articles as drafts.
var defaultRSSSources = []struct {
	name     string
	url      string
	category string
}{
	// AI category (5 sources)
	{name: "OpenAI Blog", url: "https://openai.com/blog/rss.xml", category: "ai"},
	{name: "Google AI Blog", url: "https://blog.google/technology/ai/rss/", category: "ai"},
	{name: "Anthropic Blog", url: "https://www.anthropic.com/feed", category: "ai"},
	{name: "Meta AI Blog", url: "https://ai.meta.com/blog/rss/", category: "ai"},
	{name: "DeepMind Blog", url: "https://deepmind.google/blog/rss.xml", category: "ai"},
	// Tools category (5 sources)
	{name: "Hugging Face Blog", url: "https://huggingface.co/blog/feed.xml", category: "tools"},
	{name: "AWS Machine Learning", url: "https://aws.amazon.com/blogs/machine-learning/feed/", category: "tools"},
	{name: "GitHub Blog", url: "https://github.blog/feed/", category: "tools"},
	{name: "LangChain Blog", url: "https://blog.langchain.dev/rss/", category: "tools"},
	{name: "NVIDIA Developer Blog", url: "https://developer.nvidia.com/blog/feed", category: "tools"},
	// Research category (5 sources)
	{name: "MIT AI News", url: "https://news.mit.edu/topic/mitartificial-intelligence2-rss.xml", category: "research"},
	{name: "arXiv AI", url: "https://rss.arxiv.org/rss/cs.AI", category: "research"},
	{name: "Stanford HAI", url: "https://hai.stanford.edu/news/rss.xml", category: "research"},
	{name: "Microsoft Research", url: "https://www.microsoft.com/en-us/research/feed/", category: "research"},
	{name: "The Gradient", url: "https://gradientpub.substack.com/feed", category: "research"},
	// Industry category (5 sources)
	{name: "TechCrunch AI", url: "https://techcrunch.com/category/artificial-intelligence/feed/", category: "industry"},
	{name: "VentureBeat AI", url: "https://venturebeat.com/category/ai/feed/", category: "industry"},
	{name: "The Verge AI", url: "https://www.theverge.com/rss/ai-artificial-intelligence/index.xml", category: "industry"},
	{name: "Ars Technica Tech", url: "https://feeds.arstechnica.com/arstechnica/technology-lab", category: "industry"},
	{name: "Wired AI", url: "https://www.wired.com/feed/tag/ai/latest/rss", category: "industry"},
	// Machine Learning category (5 sources)
	{name: "Machine Learning Mastery", url: "https://machinelearningmastery.com/feed/", category: "machine-learning"},
	{name: "Apple ML Research", url: "https://machinelearning.apple.com/rss.xml", category: "machine-learning"},
	{name: "Google Cloud AI & ML", url: "https://cloud.google.com/blog/products/ai-machine-learning/rss/", category: "machine-learning"},
	{name: "MarkTechPost", url: "https://www.marktechpost.com/feed/", category: "machine-learning"},
	{name: "Towards AI", url: "https://towardsai.net/feed", category: "machine-learning"},
}

// defaultSettings seeds the admin-configurable scoring thresholds on first
// boot. Defaults: LOW <60, MEDIUM 60-79, HIGH >=80.
var defaultSettings = map[string]string{
	"score_low_max":    "59",
	"score_medium_min": "60",
	"score_medium_max": "79",
	"score_high_min":   "80",
}

// Seed inserts default categories, RSS sources, and scoring settings on first
// boot. If initial seeding has already been performed, Seed is a no-op so that
// admin-deleted sources and categories are never resurrected on server restart.
func Seed(db *sql.DB) error {
	var val string
	err := db.QueryRow(`SELECT value FROM app_settings WHERE key = 'initial_seed_done'`).Scan(&val)
	if err == nil && val == "1" {
		return nil
	}

	// For existing databases where initial_seed_done was not yet recorded,
	// check if app_settings already has default score thresholds. If so, mark
	// initial_seed_done and skip re-seeding to preserve deleted sources.
	var settingsCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM app_settings WHERE key LIKE 'score_%'`).Scan(&settingsCount); err == nil && settingsCount > 0 {
		_, _ = db.Exec(`INSERT OR IGNORE INTO app_settings (key, value) VALUES ('initial_seed_done', '1')`)
		return nil
	}

	for _, c := range defaultCategories {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO categories (name, slug) VALUES (?, ?)`,
			c.name, c.slug,
		); err != nil {
			return err
		}
	}

	for _, s := range defaultRSSSources {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO rss_sources (name, url, category) VALUES (?, ?, ?)`,
			s.name, s.url, s.category,
		); err != nil {
			return err
		}
	}

	for k, v := range defaultSettings {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO app_settings (key, value) VALUES (?, ?)`,
			k, v,
		); err != nil {
			return err
		}
	}

	_, err = db.Exec(`INSERT OR REPLACE INTO app_settings (key, value) VALUES ('initial_seed_done', '1')`)
	return err
}
