/**
 * Helper to construct dynamic Open Graph (OG) image URLs from the official
 * neuralwire-media/og service at https://og.neuralwire.info/api/og.
 */
export function getOgImageUrl(params: {
	title: string;
	category?: string;
	source?: string;
	score?: number;
	readTime?: string;
}): string {
	const OG_BASE = 'https://og.neuralwire.info/api/og';
	const url = new URL(OG_BASE);

	if (params.title) url.searchParams.set('title', params.title.trim());
	if (params.category) url.searchParams.set('category', params.category.trim());
	if (params.source) url.searchParams.set('source', params.source.trim());
	if (params.score && params.score > 0) url.searchParams.set('score', params.score.toString());
	if (params.readTime) url.searchParams.set('read_time', params.readTime.trim());

	return url.toString();
}
