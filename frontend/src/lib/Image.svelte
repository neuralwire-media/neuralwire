<script lang="ts">
	import { extractFirstImage } from './image';

	let {
		src,
		content,
		alt = '',
		class: className = '',
		loading = 'lazy'
	}: {
		src?: string | null;
		content?: string | null;
		alt?: string;
		class?: string;
		// Images below the fold default to lazy loading; set `loading="eager"`
		// for the above-the-fold hero/featured image so it paints immediately
		// and stays crawlable without scroll-triggered fetching.
		loading?: 'lazy' | 'eager';
	} = $props();

	function isValidImageSource(url: string | null | undefined): boolean {
		if (!url) return false;
		// Skip known dead/deprecated endpoints that return HTTP 404
		if (url.includes('images.unsplash.com/featured/') || url.includes('source.unsplash.com')) {
			return false;
		}
		return url.startsWith('http://') || url.startsWith('https://') || url.startsWith('/');
	}

	// Computed array of image source candidates
	let candidates = $derived.by(() => {
		const list: string[] = [];
		if (isValidImageSource(src)) {
			list.push(src!);
		}
		if (content) {
			const extracted = extractFirstImage(content);
			if (isValidImageSource(extracted) && extracted !== src) {
				list.push(extracted!);
			}
		}
		return list;
	});

	let attemptIndex = $state(0);

	// The active image URL to try rendering
	let activeSrc = $derived.by(() => {
		if (attemptIndex < candidates.length) {
			return candidates[attemptIndex];
		}
		return null;
	});

	// Reset search attempt index if candidate list changes (e.g. page navigation)
	$effect(() => {
		if (candidates.length >= 0) {
			attemptIndex = 0;
		}
	});

	function handleError() {
		attemptIndex += 1;
	}

	function getResponsiveSrcSet(url: string): string | undefined {
		if (url.includes('platform.theverge.com') && (url.includes('w=') || url.includes('width='))) {
			const u400 = url.replace(/([?&]w(?:idth)?=)\d+/, '$1400');
			const u800 = url.replace(/([?&]w(?:idth)?=)\d+/, '$1800');
			const u1200 = url.replace(/([?&]w(?:idth)?=)\d+/, '$11200');
			return `${u400} 400w, ${u800} 800w, ${u1200} 1200w`;
		}
		return undefined;
	}

	let srcSet = $derived.by(() => {
		if (activeSrc) {
			return getResponsiveSrcSet(activeSrc);
		}
		return undefined;
	});
</script>

{#if activeSrc}
	<img
		src={activeSrc}
		srcset={srcSet}
		sizes={srcSet ? '(max-width: 640px) 400px, (max-width: 1024px) 800px, 1200px' : undefined}
		{alt}
		{loading}
		decoding="async"
		fetchpriority={loading === 'eager' ? 'high' : 'auto'}
		class={className}
		onerror={handleError}
	/>
{:else}
	<!-- Cybernetic editorial fallback graphic -->
	<div
		class="{className} relative flex items-center justify-center overflow-hidden bg-gradient-to-br from-[#080C14] via-[#0F172A] to-[#0A0E17]"
	>
		<div class="bg-grid-pattern absolute inset-0 opacity-20"></div>
		<svg
			class="absolute inset-0 h-full w-full opacity-25"
			xmlns="http://www.w3.org/2000/svg"
			viewBox="0 0 400 225"
			preserveAspectRatio="xMidYMid slice"
			aria-hidden="true"
		>
			<defs>
				<linearGradient id="neural-glow" x1="0%" y1="0%" x2="100%" y2="100%">
					<stop offset="0%" stop-color="#22D3EE" stop-opacity="0.6" />
					<stop offset="50%" stop-color="#38BDF8" stop-opacity="0.3" />
					<stop offset="100%" stop-color="#818CF8" stop-opacity="0.1" />
				</linearGradient>
			</defs>
			<g stroke="url(#neural-glow)" stroke-width="1" fill="none">
				<path d="M 50 180 L 120 110 L 200 130 L 280 70 L 350 90" stroke-dasharray="3 3" />
				<path d="M 30 50 L 100 80 L 180 40 L 260 90 L 370 40" stroke-dasharray="4 4" />
				<circle cx="120" cy="110" r="3" fill="#22D3EE" fill-opacity="0.7" />
				<circle cx="200" cy="130" r="2.5" fill="#38BDF8" fill-opacity="0.7" />
				<circle cx="280" cy="70" r="3.5" fill="#22D3EE" fill-opacity="0.8" />
				<circle cx="100" cy="80" r="2" fill="#818CF8" fill-opacity="0.6" />
				<circle cx="260" cy="90" r="3" fill="#38BDF8" fill-opacity="0.6" />
			</g>
		</svg>
		<div class="absolute h-28 w-28 animate-pulse rounded-full bg-[#22D3EE]/5 blur-2xl"></div>
		<div
			class="relative flex flex-col items-center justify-center gap-1 rounded border border-[#22D3EE]/20 bg-[#0A0E17]/80 px-3 py-1.5 backdrop-blur-sm"
		>
			<span
				class="pointer-events-none font-mono text-[9px] font-semibold tracking-widest text-[#22D3EE]/60 uppercase"
			>
				NEURALWIRE
			</span>
			<span
				class="pointer-events-none font-mono text-[8px] tracking-wider text-slate-500 uppercase"
			>
				INTELLIGENCE
			</span>
		</div>
	</div>
{/if}
