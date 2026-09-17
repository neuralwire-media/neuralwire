<script lang="ts">
	import { onMount } from 'svelte';
	import { BASE_URL, type TrendingArticle, type TrendingResponse } from '$lib/api';

	let { initialArticles }: { initialArticles?: TrendingArticle[] } = $props();

	let fetchedArticles = $state<TrendingArticle[] | null>(null);
	let isFetching = $state(false);
	let errorMessage = $state('');
	let windowLabel = $state('week');

	const articles = $derived(fetchedArticles ?? initialArticles ?? []);
	const isLoading = $derived(isFetching && articles.length === 0);

	async function fetchTrending() {
		if (articles.length === 0) {
			isFetching = true;
		}
		errorMessage = '';

		try {
			const res = await fetch(`${BASE_URL}/news/trending?window=week&limit=10`, {
				headers: { Accept: 'application/json' }
			});

			if (!res.ok) {
				if (articles.length === 0) errorMessage = 'Trending signal unavailable.';
				return;
			}

			const result = (await res.json()) as TrendingResponse;
			windowLabel = result.window || 'week';
			fetchedArticles = result.data || [];
		} catch (error) {
			console.warn('Trending news fetch failed', error);
			if (articles.length === 0) errorMessage = 'Trending signal unavailable.';
		} finally {
			isFetching = false;
		}
	}

	onMount(() => {
		if (!initialArticles || initialArticles.length === 0) {
			fetchTrending();
		}
	});
</script>

<div class="flex h-full w-full flex-col" aria-labelledby="trending-heading">
	<!-- Section Header -->
	<div
		class="mb-3 flex flex-shrink-0 items-end justify-between border-b border-[rgba(255,255,255,0.08)] pb-3"
	>
		<div>
			<h2
				id="trending-heading"
				class="mb-1 font-mono text-[11px] font-bold tracking-widest text-[#22D3EE] uppercase"
			>
				Trending Signal
			</h2>
			<p class="font-serif text-lg font-medium text-white lg:text-xl">MOST READ</p>
		</div>
		<div class="font-mono text-[9px] tracking-widest text-slate-500 uppercase">
			WINDOW: {windowLabel}
		</div>
	</div>

	{#if isLoading}
		<!-- Loading state -->
		<div class="flex gap-3 overflow-hidden lg:flex-grow lg:flex-col lg:justify-between lg:gap-2">
			{#each [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] as idx}
				<div
					class="h-14 w-60 flex-shrink-0 animate-pulse rounded-lg border border-[rgba(255,255,255,0.06)] bg-[#0A0E17]/60 lg:h-11 lg:w-full"
					aria-label="Loading trending item {idx}"
				></div>
			{/each}
		</div>
	{:else if articles.length > 0}
		<!-- Mobile View (< lg): Horizontal Swipeable Strip -->
		<div class="no-scrollbar flex snap-x snap-mandatory gap-3 overflow-x-auto pt-1 pb-2 lg:hidden">
			{#each articles as article, index}
				<a
					href="/{article.slug}"
					class="group flex w-64 flex-shrink-0 snap-start items-start gap-3 rounded-lg border border-[rgba(255,255,255,0.07)] bg-[#0A0E17]/60 p-3 transition-all hover:border-[#22D3EE]/35 hover:bg-[#22D3EE]/5"
				>
					<span class="font-mono text-xl leading-none font-bold text-[#22D3EE]">
						{String(index + 1).padStart(2, '0')}
					</span>
					<h3
						class="line-clamp-2 font-serif text-xs leading-snug font-medium text-slate-200 transition-colors group-hover:text-[#22D3EE]"
					>
						{article.title}
					</h3>
				</a>
			{/each}
		</div>

		<!-- Desktop View (>= lg): Vertical List with Dividers (Angka + Judul Saja, Auto-Distributed) -->
		<div
			class="hidden flex-col divide-y divide-[rgba(255,255,255,0.06)] lg:flex {articles.length >= 8
				? 'h-full flex-grow justify-between'
				: ''}"
		>
			{#each articles as article, index}
				<a
					href="/{article.slug}"
					class="group flex items-center gap-3 py-2 transition-colors {articles.length >= 8
						? 'flex-1'
						: ''}"
				>
					<span
						class="flex-shrink-0 font-mono text-xl leading-none font-bold text-[#22D3EE]/90 transition-transform group-hover:scale-105"
					>
						{String(index + 1).padStart(2, '0')}
					</span>
					<h3
						class="line-clamp-2 font-serif text-sm leading-snug font-medium text-slate-200 transition-colors group-hover:text-[#22D3EE]"
					>
						{article.title}
					</h3>
				</a>
			{/each}
		</div>
	{:else}
		<div
			class="rounded-lg border border-dashed border-[rgba(255,255,255,0.08)] bg-[#0A0E17]/40 px-4 py-6 text-center"
		>
			<p class="font-mono text-xs tracking-wider text-slate-500 uppercase">
				{errorMessage || 'No trending articles registered yet.'}
			</p>
		</div>
	{/if}
</div>
