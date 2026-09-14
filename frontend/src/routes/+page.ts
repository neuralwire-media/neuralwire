import { getNewsPage, getTrendingNews } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	const [newsRes, trending] = await Promise.all([
		getNewsPage(fetch, undefined, undefined, 1, 15),
		getTrendingNews(fetch)
	]);
	return {
		news: newsRes.articles,
		total: newsRes.total,
		totalPages: newsRes.totalPages,
		pageSize: newsRes.pageSize,
		trending
	};
};
