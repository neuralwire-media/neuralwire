import type { News } from './mockData';

const STORAGE_KEY = 'neuralwire_bookmarks_v1';

class BookmarkManager {
	private _items = $state<News[]>([]);
	private _initialized = false;

	constructor() {
		if (typeof window !== 'undefined') {
			this.load();
			window.addEventListener('storage', (e) => {
				if (e.key === STORAGE_KEY) {
					this.load();
				}
			});
		}
	}

	private load() {
		if (typeof window === 'undefined') return;
		try {
			const raw = localStorage.getItem(STORAGE_KEY);
			if (raw) {
				const parsed = JSON.parse(raw);
				if (Array.isArray(parsed)) {
					this._items = parsed;
				}
			} else {
				this._items = [];
			}
		} catch {
			this._items = [];
		}
		this._initialized = true;
	}

	private save() {
		if (typeof window === 'undefined') return;
		try {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(this._items));
		} catch {
			// Silently handle quota exceeded or local storage exceptions
		}
	}

	get items(): News[] {
		if (!this._initialized && typeof window !== 'undefined') {
			this.load();
		}
		return this._items;
	}

	get count(): number {
		return this.items.length;
	}

	isBookmarked(idOrSlug: number | string): boolean {
		return this.items.some((item) => item.id === idOrSlug || item.slug === idOrSlug);
	}

	toggle(article: News): boolean {
		const exists = this.isBookmarked(article.id || article.slug);
		if (exists) {
			this._items = this._items.filter(
				(item) => item.id !== article.id && item.slug !== article.slug
			);
			this.save();
			return false;
		} else {
			const entry: News = {
				id: article.id,
				title: article.title,
				slug: article.slug,
				url: article.url,
				source: article.source,
				category: article.category,
				summary: article.summary,
				content: '',
				image_url: article.image_url,
				view_count: article.view_count || 0,
				status: 'published',
				published_at: article.published_at,
				created_at: article.created_at || article.published_at,
				value_label: article.value_label,
				value_score: article.value_score
			};
			this._items = [
				entry,
				...this._items.filter((i) => i.id !== article.id && i.slug !== article.slug)
			];
			this.save();
			return true;
		}
	}

	remove(idOrSlug: number | string) {
		this._items = this._items.filter((item) => item.id !== idOrSlug && item.slug !== idOrSlug);
		this.save();
	}

	clear() {
		this._items = [];
		this.save();
	}
}

export const bookmarks = new BookmarkManager();
