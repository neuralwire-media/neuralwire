import { getNews, getTrendingNews } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	const [news, trending] = await Promise.all([getNews(fetch), getTrendingNews(fetch)]);
	return {
		news,
		trending
	};
};
