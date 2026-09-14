import { getNewsPage, getCategories } from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, params }) => {
	const { slug } = params;
	const categories = await getCategories(fetch);
	const category = categories.find((cat) => cat.slug === slug);

	if (!category) {
		error(404, {
			message: `Category '${slug}' not found`
		});
	}

	const newsRes = await getNewsPage(fetch, slug, undefined, 1, 15);
	return {
		category,
		news: newsRes.articles,
		total: newsRes.total,
		totalPages: newsRes.totalPages,
		pageSize: newsRes.pageSize
	};
};
