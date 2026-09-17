<script lang="ts">
	import { getSiteUrl } from '$lib/siteUrl';

	const aboutJsonLdHtml =
		'<scr' +
		'ipt type="application/ld+json">' +
		JSON.stringify({
			'@context': 'https://schema.org',
			'@type': 'AboutPage',
			name: 'About NeuralWire',
			url: `${getSiteUrl()}/about`,
			description:
				'NeuralWire is an independent curation chronicle and AI-assisted digest platform indexing frontier artificial intelligence research, neural architectures, and computational breakthroughs.',
			mainEntity: {
				'@type': 'Organization',
				name: 'NeuralWire',
				url: getSiteUrl(),
				logo: `${getSiteUrl()}/favicon.svg`
			}
		}).replace(/</g, '\\u003c') +
		'</scr' +
		'ipt>';
</script>

<svelte:head>
	<title>About | NeuralWire AI News & Editorial Curation</title>
	<meta
		name="description"
		content="NeuralWire is an independent curation chronicle and AI-assisted digest platform indexing frontier artificial intelligence research, neural architectures, and computational breakthroughs."
	/>
	<meta name="robots" content="index, follow" />
	<link rel="canonical" href="{getSiteUrl()}/about" />
	<meta property="og:title" content="About | NeuralWire AI News & Editorial Curation" />
	<meta
		property="og:description"
		content="NeuralWire is an independent curation chronicle and AI-assisted digest platform indexing frontier artificial intelligence research, neural architectures, and computational breakthroughs."
	/>
	<meta property="og:type" content="website" />
	<meta property="og:url" content="{getSiteUrl()}/about" />
	<meta property="og:image" content="{getSiteUrl()}/favicon.svg" />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content="About | NeuralWire AI News & Editorial Curation" />
	<meta
		name="twitter:description"
		content="NeuralWire is an independent curation chronicle and AI-assisted digest platform indexing frontier artificial intelligence research, neural architectures, and computational breakthroughs."
	/>
	<meta name="twitter:image" content="{getSiteUrl()}/favicon.svg" />

	<!-- Structured Data (Schema.org AboutPage & Organization) -->
	{@html aboutJsonLdHtml}
</svelte:head>

<section class="mx-auto max-w-4xl px-4 py-16 sm:px-6 md:py-24 lg:px-8">
	<!-- Navigation Breadcrumb -->
	<div class="mb-8">
		<a
			href="/"
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
			<span>RETURN TO FEED</span>
		</a>
	</div>

	<!-- Heading Header -->
	<div class="mb-14 border-b border-[rgba(255,255,255,0.08)] pb-10">
		<span class="tag-mono mb-3 block text-xs font-bold tracking-widest text-[#22D3EE] uppercase"
			>// MISSION & MANIFESTO</span
		>
		<h1 class="mb-6 font-serif text-4xl font-medium text-white sm:text-5xl">
			BRIDGING CODE AND HUMANITY
		</h1>
		<p class="max-w-3xl font-sans text-base leading-relaxed font-light text-slate-300 sm:text-lg">
			Neuralwire is an independent curation chronicle and AI-assisted news digest dedicated to
			tracking the frontier of artificial intelligence, neural networks, quantum computing, and
			computational industry.
		</p>
		<p class="mt-4 max-w-3xl font-sans text-sm leading-relaxed font-light text-slate-400">
			Our mission is to filter market noise and transform dense technical preprints, lab blogs, and
			scientific journals into structured, objective executive digests—while always connecting our
			readers directly to the original primary research.
		</p>
	</div>

	<!-- Core Values Grid -->
	<div class="mb-16">
		<span class="tag-mono mb-6 block text-xs font-bold tracking-widest text-[#22D3EE] uppercase"
			>// EDITORIAL PILLARS</span
		>
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
			<div class="glow-hover rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
				<span class="mb-2 block font-mono text-[10px] tracking-wider text-[#22D3EE] uppercase"
					>CURATORIAL RIGOR</span
				>
				<h2 class="mb-2 font-serif text-lg font-normal text-white">Filtering the Noise</h2>
				<p class="font-sans text-xs leading-relaxed font-light text-slate-400">
					We analyze technical announcements and arXiv preprints with strict objectivity. We cut
					through commercial marketing hype to highlight verifiable mathematical and engineering
					progress.
				</p>
			</div>

			<div class="glow-hover rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
				<span class="mb-2 block font-mono text-[10px] tracking-wider text-[#22D3EE] uppercase"
					>CYBERNETIC CONVERGENCE</span
				>
				<h2 class="mb-2 font-serif text-lg font-normal text-white">Futurist Perspective</h2>
				<p class="font-sans text-xs leading-relaxed font-light text-slate-400">
					The convergence of silicon cognition and human intellect is the defining transition of our
					era. We examine the interfaces between generative systems, automated reasoning, and human
					perception.
				</p>
			</div>

			<div class="glow-hover rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
				<span class="mb-2 block font-mono text-[10px] tracking-wider text-[#22D3EE] uppercase"
					>COMPUTE & PHYSICS</span
				>
				<h2 class="mb-2 font-serif text-lg font-normal text-white">Physical Foundations</h2>
				<p class="font-sans text-xs leading-relaxed font-light text-slate-400">
					From the thermodynamics of hyperscale compute clusters to topological quantum qubits, we
					chronicle the physical hardware enabling tomorrow’s machine intelligence breakthroughs.
				</p>
			</div>

			<div class="glow-hover rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
				<span class="mb-2 block font-mono text-[10px] tracking-wider text-[#22D3EE] uppercase"
					>ATTRIBUTION & ETHICS</span
				>
				<h2 class="mb-2 font-serif text-lg font-normal text-white">Author-First Sourcing</h2>
				<p class="font-sans text-xs leading-relaxed font-light text-slate-400">
					We believe in fair, ethical knowledge distribution. Every digest we publish prominently
					features direct backlinks to original publishers, research labs, and academic journals.
				</p>
			</div>
		</div>
	</div>

	<!-- The Curation Engine (How It Works) -->
	<div class="mb-16 border-t border-[rgba(255,255,255,0.08)] pt-12">
		<span class="tag-mono mb-3 block text-xs font-bold tracking-widest text-[#22D3EE] uppercase"
			>// THE CURATION ENGINE</span
		>
		<h2 class="mb-6 font-serif text-2xl font-medium text-white sm:text-3xl">
			How NeuralWire Operates
		</h2>

		<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
			<div class="rounded-xl border border-[rgba(255,255,255,0.06)] bg-[#0A0E17]/60 p-5">
				<div class="mb-3 font-mono text-xs font-bold text-[#22D3EE]">01 / AGGREGATION</div>
				<h3 class="mb-2 font-serif text-base font-medium text-white">Multi-Source Indexing</h3>
				<p class="font-sans text-xs leading-relaxed text-slate-400">
					Our automated ingestion pipeline continuously monitors public RSS feeds from leading AI
					research labs, open-access preprint repositories, and engineering publications.
				</p>
			</div>

			<div class="rounded-xl border border-[rgba(255,255,255,0.06)] bg-[#0A0E17]/60 p-5">
				<div class="mb-3 font-mono text-xs font-bold text-[#22D3EE]">02 / DISTILLATION</div>
				<h3 class="mb-2 font-serif text-base font-medium text-white">AI-Assisted Synthesis</h3>
				<p class="font-sans text-xs leading-relaxed text-slate-400">
					Large language models distill dense findings into concise executive summaries,
					highlighting methodology, impact scores, and architectural insights for technical readers.
				</p>
			</div>

			<div class="rounded-xl border border-[rgba(255,255,255,0.06)] bg-[#0A0E17]/60 p-5">
				<div class="mb-3 font-mono text-xs font-bold text-[#22D3EE]">03 / ATTRIBUTION</div>
				<h3 class="mb-2 font-serif text-base font-medium text-white">Direct Source Citation</h3>
				<p class="font-sans text-xs leading-relaxed text-slate-400">
					Every summary provides full credit with an immediate one-click link to the primary source
					article, driving engagement and referral readership back to original authors.
				</p>
			</div>
		</div>
	</div>

	<!-- Transparency, Fair Use & Editorial Integrity -->
	<div class="mb-16 rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-6 sm:p-8">
		<span class="tag-mono mb-2 block text-xs font-bold tracking-widest text-[#22D3EE] uppercase"
			>// EDITORIAL TRANSPARENCY & FAIR USE</span
		>
		<h2 class="mb-4 font-serif text-xl font-medium text-white sm:text-2xl">
			Fair Use & Discovery Commitment
		</h2>
		<p class="font-sans text-xs leading-relaxed text-slate-300 sm:text-sm">
			Neuralwire operates under international <strong>Fair Use</strong> principles. We do not reproduce
			full articles or paywalled content. Our goal is to function as a discovery index that celebrates
			original research while helping practitioners navigate the fast-moving AI landscape.
		</p>
		<div
			class="mt-4 flex flex-wrap items-center gap-4 border-t border-[rgba(255,255,255,0.06)] pt-4 font-mono text-xs"
		>
			<a href="/copyright" class="inline-flex items-center gap-1 text-[#22D3EE] hover:underline">
				<span>View Copyright & DMCA Takedown Policy</span>
				<span>&rarr;</span>
			</a>
			<span class="text-slate-600">•</span>
			<a
				href="https://status.neuralwire.info"
				target="_blank"
				rel="noopener noreferrer"
				class="inline-flex items-center gap-1.5 text-slate-300 transition-colors hover:text-[#22D3EE]"
			>
				<span class="inline-block h-1.5 w-1.5 rounded-full bg-emerald-400"></span>
				<span>System Status & Uptime</span>
				<span>&rarr;</span>
			</a>
		</div>
	</div>

	<!-- Action & Contact Grid (CTA) -->
	<div class="border-t border-[rgba(255,255,255,0.08)] pt-12">
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-3">
			<div
				class="flex flex-col justify-between rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0A0E17]/40 p-5"
			>
				<div>
					<div class="mb-2 font-mono text-xs text-[#22D3EE]">OPEN SYNDICATION</div>
					<h3 class="mb-2 font-serif text-base font-medium text-white">RSS Feed</h3>
					<p class="mb-4 font-sans text-xs text-slate-400">
						Subscribe directly in Feedly, Inoreader, or any standard RSS feed reader.
					</p>
				</div>
				<a
					href="/rss.xml"
					target="_blank"
					class="inline-block rounded-lg border border-[rgba(34,211,238,0.3)] bg-[rgba(34,211,238,0.08)] px-4 py-2 text-center font-mono text-xs font-medium text-[#22D3EE] transition-colors hover:bg-[rgba(34,211,238,0.15)]"
				>
					OPEN RSS FEED
				</a>
			</div>

			<div
				class="flex flex-col justify-between rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0A0E17]/40 p-5"
			>
				<div>
					<div class="mb-2 font-mono text-xs text-[#22D3EE]">EDITORIAL TIPS</div>
					<h3 class="mb-2 font-serif text-base font-medium text-white">Submit Research</h3>
					<p class="mb-4 font-sans text-xs text-slate-400">
						Have a breakthrough, preprint, or lab release you’d like featured on Neuralwire?
					</p>
				</div>
				<a
					href="mailto:contact@neuralwire.info"
					class="inline-block rounded-lg border border-[rgba(255,255,255,0.15)] bg-slate-800/40 px-4 py-2 text-center font-mono text-xs font-medium text-slate-200 transition-colors hover:border-slate-400"
				>
					CONTACT EDITORIAL
				</a>
			</div>

			<div
				class="flex flex-col justify-between rounded-xl border border-[rgba(255,255,255,0.08)] bg-[#0A0E17]/40 p-5"
			>
				<div>
					<div class="mb-2 font-mono text-xs text-[#22D3EE]">LIVE INTELLIGENCE</div>
					<h3 class="mb-2 font-serif text-base font-medium text-white">Explore News</h3>
					<p class="mb-4 font-sans text-xs text-slate-400">
						Browse the real-time feed categorized by research, tooling, and systems.
					</p>
				</div>
				<a
					href="/"
					class="inline-block rounded-lg border border-[rgba(34,211,238,0.3)] bg-[#22D3EE] px-4 py-2 text-center font-mono text-xs font-bold text-[#0A0E17] transition-colors hover:bg-[#06B6D4]"
				>
					EXPLORE FEED
				</a>
			</div>
		</div>
	</div>
</section>
