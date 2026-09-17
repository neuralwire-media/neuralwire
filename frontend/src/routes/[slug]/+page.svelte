<script lang="ts">
	import type { PageData } from './$types';
	import type { News } from '$lib/mockData';
	import Image from '$lib/Image.svelte';
	import BookmarkButton from '$lib/BookmarkButton.svelte';
	import { absoluteUrl, getSiteUrl } from '$lib/siteUrl';
	import { BASE_URL } from '$lib/api';
	import { getOgImageUrl } from '$lib/og';

	let { data }: { data: PageData } = $props();

	const article = $derived(data.article as News);
	const related = $derived(data.related as News[]);
	const displayItems = $derived([...related, ...related, ...related]);

	const socialOgImageUrl = $derived(
		getOgImageUrl({
			title: article.title,
			category: article.category,
			source: article.source,
			score: article.value_score,
			readTime: getReadingTime(article.summary || article.content)
		})
	);

	let copied = $state(false);

	function copyLink() {
		if (typeof navigator !== 'undefined' && navigator.clipboard) {
			navigator.clipboard.writeText(shareUrl).then(() => {
				copied = true;
				setTimeout(() => {
					copied = false;
				}, 2500);
			});
		}
	}

	function getReadingTime(text: string) {
		if (!text) return '1 min read';
		const words = text.split(/\s+/).length;
		const minutes = Math.ceil(words / 220);
		return `${minutes} min read`;
	}

	function shareOnWhatsApp() {
		openShare(
			'https://api.whatsapp.com/send?text=' + encodeURIComponent(article.title + ' ' + shareUrl)
		);
	}

	function shareOnTelegram() {
		openShare(
			'https://t.me/share/url?url=' +
				encodeURIComponent(shareUrl) +
				'&text=' +
				encodeURIComponent(article.title)
		);
	}

	// JSON-LD structured data, rendered into <svelte:head> below. The tag is
	// assembled via concatenation so the raw tag opener never appears
	// literally in this file. Dates come from the Neuralwire record
	// (published_at), not the source article's own date; authorship is the
	// site as an Organization (web curator), avoiding fictional authors. `</`
	// is escaped so a value containing the closing tag sequence cannot break
	// out of the tag.
	const articleJsonLdHtml = $derived(
		'<scr' +
			'ipt type="application/ld+json">' +
			JSON.stringify({
				'@context': 'https://schema.org',
				'@type': 'NewsArticle',
				headline: article.title,
				image: [absoluteUrl(article.image_url)],
				datePublished: article.published_at || article.created_at,
				dateModified: article.published_at || article.created_at,
				author: {
					'@type': 'Organization',
					name: 'NeuralWire',
					url: getSiteUrl()
				},
				publisher: {
					'@type': 'Organization',
					name: 'NeuralWire',
					url: getSiteUrl(),
					logo: {
						'@type': 'ImageObject',
						url: getSiteUrl() + '/favicon.svg'
					}
				},
				mainEntityOfPage: {
					'@type': 'WebPage',
					'@id': getSiteUrl() + '/' + article.slug
				}
			}).replace(/</g, '\\u003c') +
			'</scr' +
			'ipt>'
	);

	function formatDate(dateStr: string) {
		const d = new Date(dateStr);
		return d.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	// Canonical absolute URL of the article, used by the share buttons.
	const shareUrl = $derived(getSiteUrl() + '/' + article.slug);

	function openShare(url: string) {
		window.open(url, '_blank', 'noopener,width=600,height=500');
	}

	function shareOnX() {
		openShare(
			'https://twitter.com/intent/tweet?url=' +
				encodeURIComponent(shareUrl) +
				'&text=' +
				encodeURIComponent(article.title)
		);
	}

	function shareOnLinkedIn() {
		openShare(
			'https://www.linkedin.com/sharing/share-offsite/?url=' + encodeURIComponent(shareUrl)
		);
	}

	function shareOnReddit() {
		openShare(
			'https://www.reddit.com/submit?url=' +
				encodeURIComponent(shareUrl) +
				'&title=' +
				encodeURIComponent(article.title)
		);
	}

	// Feedback Form State
	let opinion = $state('');
	let email = $state('');
	let formSubmitted = $state(false);

	function handleFeedback(e: Event) {
		e.preventDefault();
		if (opinion.trim()) {
			formSubmitted = true;
			opinion = '';
			email = '';
		}
	}

	let sliderContainer = $state<HTMLDivElement>();

	$effect(() => {
		if (related.length > 0 && sliderContainer) {
			const cards = Array.from(sliderContainer.children) as HTMLElement[];
			if (cards.length >= related.length * 3) {
				sliderContainer.scrollLeft = cards[related.length].offsetLeft;
			}
		}
	});

	function handleScroll() {
		if (!sliderContainer || related.length === 0) return;
		const cards = Array.from(sliderContainer.children) as HTMLElement[];
		if (cards.length < related.length * 3) return;

		const setWidth = cards[related.length].offsetLeft;
		const maxScroll = sliderContainer.scrollWidth - sliderContainer.clientWidth;
		const currentScroll = sliderContainer.scrollLeft;

		// Silent wrap around when reaching the absolute boundaries
		if (currentScroll <= 10) {
			sliderContainer.scrollTo({
				left: currentScroll + setWidth,
				behavior: 'instant' as ScrollBehavior
			});
		} else if (currentScroll >= maxScroll - 10) {
			sliderContainer.scrollTo({
				left: currentScroll - setWidth,
				behavior: 'instant' as ScrollBehavior
			});
		}
	}

	function scrollLeft() {
		if (!sliderContainer || related.length === 0) return;
		const cards = Array.from(sliderContainer.children) as HTMLElement[];
		if (cards.length < related.length * 3) return;

		const firstSetEnd = cards[related.length].offsetLeft;
		const secondSetEnd = cards[related.length * 2].offsetLeft;
		const setWidth = secondSetEnd - firstSetEnd;
		let currentScroll = sliderContainer.scrollLeft;

		// Wrap instantly first if we are outside the middle set
		if (currentScroll < firstSetEnd - 10) {
			currentScroll += setWidth;
			sliderContainer.scrollTo({ left: currentScroll, behavior: 'instant' as ScrollBehavior });
		} else if (currentScroll >= secondSetEnd - 10) {
			currentScroll -= setWidth;
			sliderContainer.scrollTo({ left: currentScroll, behavior: 'instant' as ScrollBehavior });
		}

		// Find the active index closest to the current scroll
		let activeIndex = 0;
		let minDiff = Infinity;
		cards.forEach((card, idx) => {
			const diff = Math.abs(card.offsetLeft - currentScroll);
			if (diff < minDiff) {
				minDiff = diff;
				activeIndex = idx;
			}
		});

		// Go to previous index
		const targetIndex = activeIndex - 1;
		const targetCard = cards[targetIndex];
		if (targetCard) {
			sliderContainer.scrollTo({ left: targetCard.offsetLeft, behavior: 'smooth' });
		}
	}

	function scrollRight() {
		if (!sliderContainer || related.length === 0) return;
		const cards = Array.from(sliderContainer.children) as HTMLElement[];
		if (cards.length < related.length * 3) return;

		const firstSetEnd = cards[related.length].offsetLeft;
		const secondSetEnd = cards[related.length * 2].offsetLeft;
		const setWidth = secondSetEnd - firstSetEnd;
		let currentScroll = sliderContainer.scrollLeft;

		// Wrap instantly first if we are outside the middle set
		if (currentScroll < firstSetEnd - 10) {
			currentScroll += setWidth;
			sliderContainer.scrollTo({ left: currentScroll, behavior: 'instant' as ScrollBehavior });
		} else if (currentScroll >= secondSetEnd - 10) {
			currentScroll -= setWidth;
			sliderContainer.scrollTo({ left: currentScroll, behavior: 'instant' as ScrollBehavior });
		}

		// Find the active index closest to the current scroll
		let activeIndex = 0;
		let minDiff = Infinity;
		cards.forEach((card, idx) => {
			const diff = Math.abs(card.offsetLeft - currentScroll);
			if (diff < minDiff) {
				minDiff = diff;
				activeIndex = idx;
			}
		});

		// Go to next index
		const targetIndex = activeIndex + 1;
		const targetCard = cards[targetIndex];
		if (targetCard) {
			sliderContainer.scrollTo({ left: targetCard.offsetLeft, behavior: 'smooth' });
		}
	}

	function getViewerKey() {
		const existing = localStorage.getItem('nw_viewer_id');
		if (existing) return existing;

		const generated =
			typeof crypto !== 'undefined' && 'randomUUID' in crypto
				? crypto.randomUUID()
				: `nw-${Date.now()}-${Math.random().toString(36).slice(2)}`;

		localStorage.setItem('nw_viewer_id', generated);
		return generated;
	}

	let lastTrackedArticleId = $state<number | null>(null);

	$effect(() => {
		const currentId = article?.id;
		if (currentId && currentId !== lastTrackedArticleId) {
			lastTrackedArticleId = currentId;
			if (typeof window !== 'undefined') {
				const viewerKey = getViewerKey();
				fetch(`${BASE_URL}/news/${currentId}/view`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ viewer_key: viewerKey })
				}).catch((error) => {
					console.warn('view tracking failed', error);
				});
			}
		}
	});
</script>

<svelte:head>
	<title>{article.title} | NeuralWire</title>
	<meta name="description" content={article.summary} />
	<meta name="robots" content="index, follow" />
	<link rel="canonical" href="{getSiteUrl()}/{article.slug}" />
	<!-- Article Specific OG (Branded Social Card First) -->
	<meta property="og:title" content={article.title} />
	<meta property="og:description" content={article.summary} />
	<meta property="og:type" content="article" />
	<meta property="og:url" content="{getSiteUrl()}/{article.slug}" />
	<meta property="og:image" content={socialOgImageUrl} />
	<meta property="og:image:width" content="1200" />
	<meta property="og:image:height" content="630" />
	<!-- Twitter Card -->
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={article.title} />
	<meta name="twitter:description" content={article.summary} />
	<meta name="twitter:image" content={socialOgImageUrl} />
	<!-- Structured data: NewsArticle -->
	{@html articleJsonLdHtml}
</svelte:head>

<article class="relative w-full flex-grow pb-16">
	<!-- Top Header Grid Background -->
	<div
		class="bg-grid-pattern pointer-events-none absolute inset-0 h-[500px] border-b border-[rgba(255,255,255,0.03)] bg-gradient-to-b from-[#0F172A]/10 to-transparent opacity-10"
	></div>

	<!-- Main Container -->
	<div class="relative z-10 mx-auto max-w-4xl px-4 pt-12 sm:px-6 md:pt-16 lg:px-8">
		<!-- Back to Feed link -->
		<div class="mb-8">
			<a
				href="/"
				onclick={(e) => {
					if (typeof window !== 'undefined' && window.history.length > 1) {
						e.preventDefault();
						window.history.back();
					}
				}}
				class="group inline-flex items-center space-x-2 font-mono text-xs text-slate-400 transition-colors hover:text-[#22D3EE]"
			>
				<svg
					class="h-4 w-4 transform transition-transform group-hover:-translate-x-1"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M10 19l-7-7m0 0l7-7m-7 7h18"
					/>
				</svg>
				<span>RETURN TO CHRONICLE FEED</span>
			</a>
		</div>

		<!-- Category & Date Header -->
		<div class="mb-6 flex flex-wrap items-center gap-3 text-xs">
			<a
				href="/category/{article.category}"
				class="tag-mono rounded border border-[#22D3EE]/30 bg-[#22D3EE]/5 px-2.5 py-0.5 font-bold text-[#22D3EE] uppercase transition-all hover:bg-[#22D3EE]/10"
			>
				{article.category.replace('-', ' ')}
			</a>
			<span class="font-mono text-slate-600">•</span>
			<span class="font-mono text-slate-400">{formatDate(article.published_at)}</span>
			<span class="font-mono text-slate-600">•</span>
			<span class="font-mono font-bold text-[#22D3EE]">AI DIGEST SUMMARY</span>
		</div>

		<!-- Title & Subtitle -->
		<h1
			class="mb-8 font-serif text-3xl leading-tight font-medium text-white sm:text-4xl md:text-5xl"
		>
			{article.title}
		</h1>

		<!-- Large Main Image -->
		<div
			class="relative mb-12 aspect-[16/9] w-full overflow-hidden rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]"
		>
			<Image
				src={article.image_url}
				content={article.content}
				alt={article.title}
				loading="eager"
				class="h-full w-full object-cover opacity-90"
			/>
			<div
				class="absolute bottom-4 left-4 rounded-lg border border-[rgba(255,255,255,0.05)] bg-[#0A0E17]/80 px-3 py-1.5 font-mono text-[10px] text-slate-400 backdrop-blur-sm"
			>
				SOURCE: {article.source.toUpperCase()}
			</div>
		</div>
		<!-- Sticky Action & Quick Share Bar -->
		<div
			class="sticky top-20 z-20 mb-10 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#070A10]/95 px-4 py-2.5 shadow-2xl backdrop-blur-md"
		>
			<div class="flex items-center gap-3">
				<BookmarkButton
					{article}
					size="md"
					showText={true}
					class="rounded-lg border border-[rgba(255,255,255,0.08)] bg-[#0F172A] px-3 py-1.5 hover:border-[#22D3EE]/40 hover:bg-[#22D3EE]/5"
				/>
				<span class="h-4 w-[1px] bg-white/10"></span>
				<span class="font-mono text-xs text-slate-400">
					{getReadingTime(article.summary || article.content)}
				</span>
			</div>

			<!-- Quick Share Buttons -->
			<div class="flex items-center gap-1.5">
				<button
					type="button"
					onclick={copyLink}
					class="relative flex h-8 items-center gap-1.5 rounded-lg border border-[rgba(255,255,255,0.08)] bg-[#0F172A] px-2.5 font-mono text-xs text-slate-300 transition-all hover:border-[#22D3EE]/40 hover:bg-[#22D3EE]/5 hover:text-[#22D3EE]"
					title="Copy link to clipboard"
				>
					{#if copied}
						<svg
							class="h-3.5 w-3.5 text-[#22D3EE]"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M5 13l4 4L19 7"
							/>
						</svg>
						<span class="text-[11px] font-bold text-[#22D3EE]">COPIED</span>
					{:else}
						<svg
							class="h-3.5 w-3.5 text-slate-400"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
							/>
						</svg>
						<span class="hidden sm:inline">COPY</span>
					{/if}
				</button>

				<button
					type="button"
					onclick={shareOnX}
					class="flex h-8 w-8 items-center justify-center rounded-lg border border-[rgba(255,255,255,0.08)] bg-[#0F172A] text-slate-400 transition-all hover:border-[#22D3EE]/40 hover:bg-[#22D3EE]/5 hover:text-[#22D3EE]"
					title="Share on X"
					aria-label="Share on X"
				>
					<svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 24 24">
						<path
							d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"
						/>
					</svg>
				</button>

				<button
					type="button"
					onclick={shareOnLinkedIn}
					class="flex h-8 w-8 items-center justify-center rounded-lg border border-[rgba(255,255,255,0.08)] bg-[#0F172A] text-slate-400 transition-all hover:border-[#22D3EE]/40 hover:bg-[#22D3EE]/5 hover:text-[#22D3EE]"
					title="Share on LinkedIn"
					aria-label="Share on LinkedIn"
				>
					<svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 24 24">
						<path
							d="M19 3a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14m-.5 15.5v-5.3a3.26 3.26 0 0 0-3.26-3.26c-.85 0-1.84.52-2.28 1.3v-1.11h-2.79v8.37h2.79v-4.93c0-.77.62-1.4 1.39-1.4a1.4 1.4 0 0 1 1.4 1.4v4.93h2.75M6.46 10.9v8.37H9.2V10.9H6.46M7.83 6.27a1.62 1.62 0 1 0 0 3.24 1.62 1.62 0 0 0 0-3.24z"
						/>
					</svg>
				</button>

				<button
					type="button"
					onclick={shareOnWhatsApp}
					class="flex h-8 w-8 items-center justify-center rounded-lg border border-[rgba(255,255,255,0.08)] bg-[#0F172A] text-slate-400 transition-all hover:border-[#22D3EE]/40 hover:bg-[#22D3EE]/5 hover:text-[#22D3EE]"
					title="Share on WhatsApp"
					aria-label="Share on WhatsApp"
				>
					<svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 24 24">
						<path
							d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0 0 12.04 2m.01 1.67c2.2 0 4.26.86 5.82 2.42a8.225 8.225 0 0 1 2.41 5.83c0 4.54-3.7 8.24-8.24 8.24-1.48 0-2.93-.4-4.2-1.15l-.3-.18-3.12.82.83-3.04-.2-.31a8.196 8.196 0 0 1-1.26-4.38c0-4.54 3.7-8.24 8.24-8.24M8.53 7.33c-.16 0-.35.06-.53.27-.18.2-.69.67-.69 1.64s.71 1.9 1.01 2.11c.1.1 1.37 2.1 3.33 2.94.46.2.83.33 1.11.42.47.15.9.13 1.23.08.38-.06 1.15-.47 1.31-.92.16-.46.16-.85.11-.93-.05-.08-.18-.13-.38-.23s-1.15-.57-1.33-.63c-.18-.06-.31-.1-.44.1-.13.2-.5.63-.61.76-.11.13-.23.15-.43.05s-.85-.31-1.62-.99c-.6-.54-1-1.2-1.12-1.4-.11-.2-.01-.31.09-.41.09-.09.2-.23.3-.35.1-.11.13-.2.2-.33.06-.13.03-.25-.01-.35s-.44-1.07-.61-1.47c-.16-.38-.33-.33-.45-.33"
						/>
					</svg>
				</button>

				<button
					type="button"
					onclick={shareOnTelegram}
					class="flex h-8 w-8 items-center justify-center rounded-lg border border-[rgba(255,255,255,0.08)] bg-[#0F172A] text-slate-400 transition-all hover:border-[#22D3EE]/40 hover:bg-[#22D3EE]/5 hover:text-[#22D3EE]"
					title="Share on Telegram"
					aria-label="Share on Telegram"
				>
					<svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 24 24">
						<path
							d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.2.2 0 0 0-.05-.18c-.06-.05-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.75-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.14.27-.01.06.01.24 0 .38z"
						/>
					</svg>
				</button>
			</div>
		</div>

		<!-- Content Presentation -->
		<div class="mb-16 space-y-10">
			<!-- AI Digest & Key Takeaways Highlight Card -->
			{#if article.summary}
				<div
					class="relative overflow-hidden rounded-2xl border border-[#22D3EE]/25 bg-gradient-to-br from-[#22D3EE]/5 via-[#0F172A]/50 to-[#0A0E17] p-6 backdrop-blur-md md:p-8"
				>
					<div class="mb-4 flex items-center justify-between border-b border-[#22D3EE]/15 pb-3">
						<div class="flex items-center gap-2.5">
							<span
								class="inline-block h-2.5 w-2.5 animate-pulse rounded-full bg-[#22D3EE] shadow-[0_0_8px_rgba(34,211,238,0.8)]"
							></span>
							<h2 class="font-mono text-xs font-bold tracking-widest text-[#22D3EE] uppercase">
								Executive Intelligence Brief
							</h2>
						</div>
						<span class="font-mono text-[10px] tracking-wider text-slate-500 uppercase">
							KEY TAKEAWAYS
						</span>
					</div>
					<div
						class="article-content max-w-none font-sans text-base leading-relaxed font-light text-slate-200 md:text-lg"
					>
						<p>{article.summary}</p>
					</div>
				</div>
			{:else}
				<div
					class="rounded-2xl border border-[#22D3EE]/15 bg-[#22D3EE]/3 p-6 backdrop-blur-sm md:p-8"
				>
					<span class="mb-3 block font-mono text-[10px] font-bold tracking-wider text-[#22D3EE]">
						Neural AI Digest
					</span>
					<div
						class="article-content max-w-none font-sans text-base leading-relaxed font-light text-slate-200 md:text-lg"
					>
						<p>This brief is being prepared. Read the full story at the original source below.</p>
					</div>
				</div>
			{/if}

			<!-- Multi-Source Coverage Card (Clustered Articles from Other Sources) -->
			{#if article.cluster_coverage && article.cluster_coverage.length > 0}
				<div
					class="relative my-10 overflow-hidden rounded-2xl border border-purple-500/20 bg-[#0F172A]/40 p-6 backdrop-blur-sm md:p-8"
				>
					<div class="mb-4 flex flex-wrap items-center justify-between gap-2">
						<div class="flex items-center gap-2.5">
							<div
								class="flex h-7 w-7 items-center justify-center rounded-lg border border-purple-500/30 bg-purple-500/10 text-purple-400"
							>
								<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
									/>
								</svg>
							</div>
							<div>
								<h3 class="font-mono text-xs font-bold tracking-widest text-purple-300 uppercase">
									Multi-Source Coverage
								</h3>
								<p class="font-sans text-[11px] text-slate-400">
									Liputan peristiwa serupa dari {article.cluster_coverage.length} perspektif media lainnya
								</p>
							</div>
						</div>
						<span
							class="rounded-full border border-purple-500/30 bg-purple-500/10 px-2.5 py-0.5 font-mono text-[10px] font-bold text-purple-300"
						>
							+{article.cluster_coverage.length} Sumber Lain
						</span>
					</div>

					<div class="grid grid-cols-1 gap-3 pt-2 sm:grid-cols-2">
						{#each article.cluster_coverage as coverage}
							<a
								href="/{coverage.slug}"
								class="group flex flex-col justify-between rounded-xl border border-[rgba(255,255,255,0.06)] bg-[#070A10]/60 p-4 transition-all duration-200 hover:border-purple-500/40 hover:bg-[#070A10]"
							>
								<div>
									<div
										class="mb-2 flex items-center justify-between font-mono text-[10px] text-slate-500"
									>
										<span class="font-bold text-purple-400">{coverage.source.toUpperCase()}</span>
										<span>{formatDate(coverage.published_at || coverage.created_at)}</span>
									</div>
									<h4
										class="line-clamp-2 font-serif text-sm font-medium text-slate-200 transition-colors group-hover:text-purple-300"
									>
										{coverage.title}
									</h4>
									{#if coverage.summary}
										<p class="mt-2 line-clamp-2 font-sans text-xs text-slate-400">
											{coverage.summary}
										</p>
									{/if}
								</div>
								<div
									class="mt-3 flex items-center justify-end font-mono text-[10px] text-purple-400/80 group-hover:text-purple-300"
								>
									<span>BACA PERSPEKTIF INI &rarr;</span>
								</div>
							</a>
						{/each}
					</div>
				</div>
			{/if}

			<!-- External Source CTA Card -->
			<div
				class="relative my-12 overflow-hidden rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/40 p-6 text-center backdrop-blur-sm md:p-8"
			>
				<div
					class="pointer-events-none absolute top-[-50%] right-[-50%] h-64 w-64 rounded-full bg-[#22D3EE]/3 blur-3xl"
				></div>
				<h3 class="mb-3 font-mono text-[10px] font-bold tracking-widest text-[#22D3EE] uppercase">
					Original Coverage
				</h3>
				<p class="mx-auto mb-6 max-w-lg font-sans text-xs leading-relaxed text-slate-400">
					For interactive features, code blocks, or full research diagrams, read the original
					publication.
				</p>
				<a
					href={article.url}
					target="_blank"
					rel="noopener noreferrer"
					class="inline-block rounded-xl bg-[#22D3EE] px-8 py-3.5 font-mono text-xs font-bold tracking-widest text-[#0A0E17] uppercase shadow-[0_0_15px_rgba(34,211,238,0.2)] transition-all hover:scale-[1.02] hover:bg-[#22D3EE]/90 active:scale-[0.98]"
				>
					READ FULL STORY ON {article.source.toUpperCase()}
				</a>
			</div>
		</div>
		<!-- Share & Source telemetry footer -->
		<div
			class="mb-16 flex flex-col items-center justify-between gap-4 border-y border-[rgba(255,255,255,0.08)] py-6 font-mono text-xs text-slate-500 sm:flex-row"
		>
			<div>
				<span>Origin: {article.source}</span>
			</div>

			<div class="flex items-center space-x-4">
				<span class="text-slate-400">Share:</span>
				<button
					type="button"
					onclick={shareOnX}
					class="cursor-pointer transition-colors hover:text-[#22D3EE]"
				>
					X
				</button>
				<button
					type="button"
					onclick={shareOnLinkedIn}
					class="cursor-pointer transition-colors hover:text-[#22D3EE]"
				>
					LinkedIn
				</button>
				<button
					type="button"
					onclick={shareOnReddit}
					class="cursor-pointer transition-colors hover:text-[#22D3EE]"
				>
					Reddit
				</button>
			</div>
		</div>

		<!-- Telemetry feedback form -->
		<section
			class="relative mb-20 overflow-hidden rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-6 md:p-8"
		>
			<!-- subtle glow -->
			<div
				class="pointer-events-none absolute top-[-10%] right-[-10%] h-32 w-32 rounded-full bg-[#22D3EE]/3 blur-2xl"
			></div>

			<span class="tag-mono mb-2 block text-[10px] font-bold tracking-wider text-[#22D3EE]"
				>Feedback Receiver</span
			>
			<h3 class="mb-2 font-serif text-xl font-medium text-white">OPINION TRANSMISSION</h3>
			<p class="mb-6 font-sans text-xs leading-relaxed font-light text-slate-400">
				Provide your mathematical or ethical stance on this report. All submissions are queued for
				model adjustment evaluation.
			</p>

			{#if formSubmitted}
				<div class="rounded-xl border border-[#22D3EE]/30 bg-[#22D3EE]/5 p-6 text-center">
					<svg
						class="mx-auto mb-3 h-8 w-8 animate-pulse text-[#22D3EE]"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<p class="mb-1 font-mono text-sm tracking-wider text-[#22D3EE] uppercase">
						TRANSMISSION RECEIVED
					</p>
					<p class="font-sans text-[11px] text-slate-400">
						Telemetry logs have been queued. Thank you for participating in public cognition cycles.
					</p>
				</div>
			{:else}
				<form onsubmit={handleFeedback} class="space-y-4">
					<div>
						<label
							for="opinion"
							class="mb-1.5 block font-mono text-[10px] tracking-wider text-slate-400 uppercase"
							>Opinion Content</label
						>
						<textarea
							id="opinion"
							rows="4"
							placeholder="Type opinion here..."
							bind:value={opinion}
							required
							class="w-full rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#070A10] px-3 py-2.5 font-mono text-xs text-slate-100 placeholder-slate-600 focus:border-[#22D3EE]/50 focus:outline-none"
						></textarea>
					</div>

					<div class="grid grid-cols-1 items-end gap-4 sm:grid-cols-2">
						<div>
							<label
								for="email"
								class="mb-1.5 block font-mono text-[10px] tracking-wider text-slate-400 uppercase"
								>Email Signature (Optional)</label
							>
							<input
								type="email"
								id="email"
								placeholder="quantum@neutralwire.media"
								bind:value={email}
								class="w-full rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#070A10] px-3 py-2 font-mono text-xs text-slate-100 placeholder-slate-600 focus:border-[#22D3EE]/50 focus:outline-none"
							/>
						</div>
						<div class="text-right">
							<button
								type="submit"
								class="accent-glow-glow w-full cursor-pointer rounded-xl bg-[#22D3EE] px-6 py-2 font-mono text-xs font-bold tracking-widest text-[#0A0E17] uppercase transition-all hover:bg-[#22D3EE]/90 sm:w-auto"
							>
								Submit Opinion
							</button>
						</div>
					</div>
				</form>
			{/if}
		</section>

		<!-- Related Recommendations Feed -->
		{#if related.length > 0}
			<section class="border-t border-[rgba(255,255,255,0.08)] pt-12">
				<h3 class="mb-6 font-mono text-xs font-bold tracking-widest text-[#22D3EE] uppercase">
					Related Transmissions
				</h3>

				<div class="group relative">
					<!-- Left Scroll Button -->
					{#if related.length > 3}
						<button
							onclick={scrollLeft}
							class="absolute top-1/2 left-[-20px] z-10 flex h-10 w-10 translate-y-[-50%] cursor-pointer items-center justify-center rounded-full border border-[rgba(255,255,255,0.08)] bg-[#0A0E17]/95 text-slate-400 opacity-80 shadow-lg backdrop-blur-sm transition-all group-hover:opacity-100 hover:border-[#22D3EE] hover:text-white md:left-[-24px]"
							aria-label="Scroll left"
						>
							<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2.5"
									d="M15 19l-7-7 7-7"
								/>
							</svg>
						</button>
					{/if}

					<!-- Slider Container -->
					<div
						bind:this={sliderContainer}
						onscroll={handleScroll}
						class="no-scrollbar flex snap-x snap-mandatory gap-6 overflow-x-auto py-2"
					>
						{#each displayItems as post}
							<article
								class="group glow-hover flex w-full flex-shrink-0 snap-start snap-always flex-col overflow-hidden rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 sm:w-[calc(50%-12px)] md:w-[calc(33.333%-16px)]"
							>
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
								</a>

								<div class="flex flex-grow flex-col space-y-2 p-4">
									<div class="font-mono text-[9px] text-slate-500">
										{formatDate(post.published_at)}
									</div>
									<h4
										class="line-clamp-2 font-serif text-sm leading-snug font-normal text-white transition-colors group-hover:text-[#22D3EE]"
									>
										<a href="/{post.slug}">
											{post.title}
										</a>
									</h4>
								</div>
							</article>
						{/each}
					</div>

					<!-- Right Scroll Button -->
					{#if related.length > 3}
						<button
							onclick={scrollRight}
							class="absolute top-1/2 right-[-20px] z-10 flex h-10 w-10 translate-y-[-50%] cursor-pointer items-center justify-center rounded-full border border-[rgba(255,255,255,0.08)] bg-[#0A0E17]/95 text-slate-400 opacity-80 shadow-lg backdrop-blur-sm transition-all group-hover:opacity-100 hover:border-[#22D3EE] hover:text-white md:right-[-24px]"
							aria-label="Scroll right"
						>
							<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2.5"
									d="M9 5l7 7-7 7"
								/>
							</svg>
						</button>
					{/if}
				</div>
			</section>
		{/if}
	</div>
</article>

<style>
	/* Custom rich text rendering styles for article body */
	:global(.article-content p) {
		margin-bottom: 1.5rem;
		line-height: 1.8;
		font-weight: 300;
		color: #cbd5e1; /* slate-300 */
	}

	:global(.article-content h2) {
		font-family: 'Playfair Display', Georgia, serif;
		font-size: 1.5rem;
		font-weight: 500;
		color: #ffffff;
		margin-top: 2.25rem;
		margin-bottom: 1rem;
		padding-bottom: 0.5rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
	}

	@media (min-width: 768px) {
		:global(.article-content h2) {
			font-size: 1.75rem;
		}
	}

	:global(.article-content blockquote) {
		font-family: 'Playfair Display', Georgia, serif;
		font-style: italic;
		font-size: 1.125rem;
		border-left: 2px solid #22d3ee;
		padding-left: 1.25rem;
		padding-top: 0.75rem;
		padding-bottom: 0.75rem;
		margin: 2rem 0;
		background-color: rgba(34, 211, 238, 0.03);
		border-top-right-radius: 0.5rem;
		border-bottom-right-radius: 0.5rem;
	}

	:global(.article-content em) {
		color: #ffffff;
	}

	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none; /* IE and Edge */
		scrollbar-width: none; /* Firefox */
	}
</style>
