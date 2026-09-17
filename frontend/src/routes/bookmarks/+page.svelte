<script lang="ts">
	import { bookmarks } from '$lib/bookmarks.svelte';
	import Image from '$lib/Image.svelte';

	let searchQuery = $state('');

	function formatDate(dateStr: string) {
		if (!dateStr) return '';
		const d = new Date(dateStr);
		return d.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function getReadingTime(text: string) {
		if (!text) return '1 min read';
		const words = text.split(/\s+/).length;
		const minutes = Math.ceil(words / 220);
		return `${minutes} min read`;
	}

	let filteredItems = $derived.by(() => {
		const q = searchQuery.trim().toLowerCase();
		if (!q) return bookmarks.items;
		return bookmarks.items.filter(
			(item) =>
				item.title.toLowerCase().includes(q) ||
				item.summary.toLowerCase().includes(q) ||
				item.category.toLowerCase().includes(q) ||
				item.source.toLowerCase().includes(q)
		);
	});

	function handleClearAll() {
		if (confirm('Are you sure you want to remove all saved transmissions?')) {
			bookmarks.clear();
		}
	}
</script>

<svelte:head>
	<title>Saved Archives | NeuralWire</title>
	<meta name="description" content="Personal offline intelligence and saved NeuralWire archives." />
	<meta name="robots" content="noindex, follow" />
</svelte:head>

<div class="min-h-screen bg-[#0A0E17] pb-24 text-slate-100">
	<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 md:py-12 lg:px-8">
		<!-- Header Section -->
		<div
			class="mb-8 flex flex-col justify-between gap-4 border-b border-[rgba(255,255,255,0.08)] pb-6 md:flex-row md:items-end"
		>
			<div>
				<div class="mb-2 flex items-center gap-2">
					<span
						class="rounded border border-[#22D3EE]/30 bg-[#22D3EE]/10 px-2 py-0.5 font-mono text-[10px] font-bold tracking-widest text-[#22D3EE] uppercase"
					>
						BOOKMARKS
					</span>
				</div>
				<h1 class="font-serif text-3xl font-medium text-white md:text-4xl">SAVED ARCHIVES</h1>
			</div>

			<!-- Actions: Search & Clear All -->
			{#if bookmarks.count > 0}
				<div class="flex flex-wrap items-center gap-3">
					<div class="relative min-w-[200px]">
						<input
							type="text"
							placeholder="Filter saved..."
							bind:value={searchQuery}
							class="w-full rounded-lg border border-[rgba(255,255,255,0.08)] bg-[#0F172A] px-3 py-1.5 pl-8 font-mono text-xs text-slate-200 placeholder-slate-500 transition-all focus:border-[#22D3EE]/50 focus:ring-1 focus:ring-[#22D3EE]/30 focus:outline-none"
						/>
						<svg
							class="absolute top-2 left-2.5 h-3.5 w-3.5 text-slate-500"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
							/>
						</svg>
					</div>

					<button
						type="button"
						onclick={handleClearAll}
						class="rounded-lg border border-red-500/20 bg-red-500/10 px-3 py-1.5 font-mono text-xs tracking-wider text-red-400 transition-all hover:border-red-500/40 hover:bg-red-500/20 hover:text-red-300"
					>
						CLEAR ALL
					</button>
				</div>
			{/if}
		</div>

		<!-- Bookmarks Grid / Empty State -->
		{#if bookmarks.count === 0}
			<!-- Empty State -->
			<div
				class="my-16 flex flex-col items-center justify-center rounded-2xl border border-dashed border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 px-4 py-20 text-center"
			>
				<div
					class="mb-6 flex h-20 w-20 items-center justify-center rounded-2xl border border-[#22D3EE]/30 bg-[#22D3EE]/5 text-[#22D3EE] shadow-[0_0_30px_rgba(34,211,238,0.15)]"
				>
					<svg class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="1.5"
							d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
						/>
					</svg>
				</div>
				<h2 class="mb-2 font-serif text-2xl font-normal text-white">NO SAVED ARCHIVES</h2>
				<p class="max-w-md font-sans text-sm font-light text-slate-400">
					Your personal offline dossier is currently empty. Click the bookmark icon on any article
					card or story to store it for quick reference.
				</p>
				<a
					href="/"
					class="mt-8 rounded-lg border border-[#22D3EE]/40 bg-[#22D3EE]/10 px-5 py-2.5 font-mono text-xs font-bold tracking-widest text-[#22D3EE] uppercase transition-all hover:bg-[#22D3EE] hover:text-[#0A0E17] hover:shadow-[0_0_20px_rgba(34,211,238,0.4)]"
				>
					EXPLORE DISPATCHES
				</a>
			</div>
		{:else if filteredItems.length === 0}
			<div class="py-16 text-center text-slate-400">
				<p class="font-mono text-sm">NO SAVED ARCHIVES MATCHING "{searchQuery}"</p>
			</div>
		{:else}
			<div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3 lg:gap-8 2xl:grid-cols-4">
				{#each filteredItems as post (post.id || post.slug)}
					<article
						class="group glow-hover relative flex flex-col overflow-hidden rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/25"
					>
						<!-- Thumbnail -->
						<a
							href="/{post.slug}"
							class="relative block aspect-[16/10] w-full overflow-hidden border-b border-[rgba(255,255,255,0.08)] bg-[#0A0E17]"
						>
							<Image
								src={post.image_url}
								content={post.content}
								alt={post.title}
								class="h-full w-full object-cover opacity-75 transition-all duration-550 group-hover:scale-105 group-hover:opacity-100"
							/>
							<div class="absolute bottom-2 left-2">
								<span
									class="tag-mono rounded border border-[#22D3EE]/30 bg-[#0A0E17]/90 px-2 py-0.5 text-[10px] font-bold text-[#22D3EE] backdrop-blur-sm"
								>
									{post.category?.toUpperCase() || 'AI'}
								</span>
							</div>
						</a>

						<!-- Card Body -->
						<div class="flex flex-grow flex-col space-y-3 p-5">
							<!-- Meta -->
							<div class="flex items-center space-x-2 font-mono text-[10px] text-slate-500">
								<span>{formatDate(post.published_at)}</span>
								<span>•</span>
								<span>{getReadingTime(post.summary)}</span>
							</div>

							<!-- Title -->
							<h3
								class="flex-grow font-serif text-lg leading-snug font-normal text-white transition-colors group-hover:text-[#22D3EE]"
							>
								<a href="/{post.slug}" class="line-clamp-2">
									{post.title}
								</a>
							</h3>

							<!-- Summary -->
							<p class="line-clamp-3 font-sans text-xs leading-relaxed font-light text-slate-400">
								{post.summary}
							</p>

							<!-- Footer / Actions -->
							<div
								class="flex items-center justify-between border-t border-[rgba(255,255,255,0.05)] pt-4 font-mono text-[10px] text-slate-500"
							>
								<span class="text-slate-400">{post.source?.toUpperCase() || 'NEURALWIRE'}</span>
								<div class="flex items-center gap-3">
									<button
										type="button"
										onclick={() => bookmarks.remove(post.id || post.slug)}
										class="text-slate-500 transition-colors hover:text-red-400"
										title="Remove from saved"
									>
										REMOVE
									</button>
									<a
										href="/{post.slug}"
										class="flex items-center space-x-1 text-[#22D3EE]/70 group-hover:text-[#22D3EE]"
									>
										<span>READ</span>
										<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M9 5l7 7-7 7"
											/>
										</svg>
									</a>
								</div>
							</div>
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</div>
</div>
