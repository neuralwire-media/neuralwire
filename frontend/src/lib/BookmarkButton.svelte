<script lang="ts">
	import { bookmarks } from '$lib/bookmarks.svelte';
	import type { News } from '$lib/mockData';

	let {
		article,
		size = 'sm',
		showText = false,
		class: className = ''
	}: {
		article: News;
		size?: 'sm' | 'md' | 'lg';
		showText?: boolean;
		class?: string;
	} = $props();

	let isSaved = $derived(bookmarks.isBookmarked(article.id || article.slug));

	function handleToggle(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		bookmarks.toggle(article);
	}
</script>

<button
	type="button"
	onclick={handleToggle}
	class="inline-flex items-center gap-1.5 transition-all focus:outline-none {className}"
	aria-label={isSaved ? 'Remove from bookmarks' : 'Save article to bookmarks'}
	title={isSaved ? 'Saved in bookmarks' : 'Bookmark this article'}
>
	{#if isSaved}
		<svg
			class="{size === 'lg'
				? 'h-5 w-5'
				: size === 'md'
					? 'h-4 w-4'
					: 'h-3.5 w-3.5'} text-[#22D3EE] drop-shadow-[0_0_8px_rgba(34,211,238,0.6)]"
			viewBox="0 0 24 24"
			fill="currentColor"
		>
			<path d="M5 5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v16l-7-3.5L5 21V5z" />
		</svg>
	{:else}
		<svg
			class="{size === 'lg'
				? 'h-5 w-5'
				: size === 'md'
					? 'h-4 w-4'
					: 'h-3.5 w-3.5'} text-slate-400 hover:text-[#22D3EE]"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M5 5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v16l-7-3.5L5 21V5z"
			/>
		</svg>
	{/if}

	{#if showText}
		<span
			class="font-mono text-xs font-medium tracking-wider uppercase {isSaved
				? 'text-[#22D3EE]'
				: 'text-slate-300'}"
		>
			{isSaved ? 'SAVED' : 'BOOKMARK'}
		</span>
	{/if}
</button>
