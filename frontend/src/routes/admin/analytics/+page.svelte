<script lang="ts">
	import { onMount } from 'svelte';
	import { BASE_URL } from '$lib/api';

	interface TopArticle {
		id: number;
		title: string;
		slug: string;
		category: string;
		view_count: number;
	}

	interface CategoryCount {
		category: string;
		count: number;
	}

	interface AnalyticsPayload {
		views: {
			total_views: number;
			top_articles: TopArticle[];
		};
		content: {
			drafts_count: number;
			published_count: number;
			rejected_count: number;
			category_distribution: CategoryCount[];
			score_distribution: {
				high: number;
				medium: number;
				low: number;
			};
		};
		system: {
			uptime_seconds: number;
			http_requests_total: number;
			http_errors_total: number;
			http_avg_latency_ms: number;
			fetch_cycles_total: number;
			fetch_cycles_failed: number;
			ai_calls_total: number;
			ai_calls_failed: number;
			memory_alloc_mb: number;
			memory_sys_mb: number;
			num_goroutines: number;
		};
	}

	let data = $state<AnalyticsPayload | null>(null);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let errorMessage = $state('');
	let lastUpdated = $state('');

	async function fetchAnalytics(manual = false) {
		if (manual) isRefreshing = true;
		else isLoading = true;
		errorMessage = '';

		const token = localStorage.getItem('admin_token');
		if (!token) return;

		try {
			const res = await fetch(`${BASE_URL}/admin/analytics`, {
				headers: {
					Authorization: `Bearer ${token}`,
					Accept: 'application/json'
				}
			});

			if (res.ok) {
				data = await res.json();
				lastUpdated = new Date().toLocaleTimeString('en-US', {
					hour: '2-digit',
					minute: '2-digit',
					second: '2-digit'
				});
			} else {
				if (res.status === 401 || res.status === 403) {
					localStorage.removeItem('admin_token');
					window.location.reload();
					return;
				}
				errorMessage = 'Failed to retrieve telemetry and analytics stream.';
			}
		} catch (err) {
			console.error('Analytics request failed:', err);
			errorMessage = 'Telemetry node connection timeout.';
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}

	function formatUptime(seconds: number): string {
		const d = Math.floor(seconds / (3600 * 24));
		const h = Math.floor((seconds % (3600 * 24)) / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		const s = seconds % 60;
		if (d > 0) return `${d}d ${h}h ${m}m`;
		if (h > 0) return `${h}h ${m}m ${s}s`;
		return `${m}m ${s}s`;
	}

	const maxTopViews = $derived(
		data && data.views.top_articles.length > 0
			? Math.max(...data.views.top_articles.map((a) => a.view_count), 1)
			: 1
	);

	const totalCategorizedArticles = $derived(
		data ? data.content.category_distribution.reduce((acc, c) => acc + c.count, 0) || 1 : 1
	);

	const totalScoredArticles = $derived(
		data
			? data.content.score_distribution.high +
					data.content.score_distribution.medium +
					data.content.score_distribution.low || 1
			: 1
	);

	const httpErrorRate = $derived(
		data && data.system.http_requests_total > 0
			? ((data.system.http_errors_total / data.system.http_requests_total) * 100).toFixed(2)
			: '0.00'
	);

	const fetchSuccessRate = $derived(
		data && data.system.fetch_cycles_total > 0
			? (
					((data.system.fetch_cycles_total - data.system.fetch_cycles_failed) /
						data.system.fetch_cycles_total) *
					100
				).toFixed(1)
			: '100.0'
	);

	const aiSuccessRate = $derived(
		data && data.system.ai_calls_total > 0
			? (
					((data.system.ai_calls_total - data.system.ai_calls_failed) /
						data.system.ai_calls_total) *
					100
				).toFixed(1)
			: '100.0'
	);

	onMount(() => {
		fetchAnalytics();
	});
</script>

<svelte:head>
	<title>Analytics &amp; System Telemetry | NeuralWire</title>
</svelte:head>

<section
	class="mx-auto flex max-w-7xl flex-grow flex-col justify-start px-4 py-12 sm:px-6 md:py-16 lg:px-8"
>
	<!-- Header -->
	<div
		class="mb-8 flex flex-col items-start justify-between gap-4 border-b border-[rgba(255,255,255,0.08)] pb-6 sm:flex-row sm:items-end"
	>
		<div>
			<span class="tag-mono mb-1 block text-xs font-bold tracking-widest text-[#22D3EE]"
				>Telemetry &amp; Insights</span
			>
			<h1 class="font-serif text-3xl font-medium text-white">SYSTEM ANALYTICS</h1>
		</div>
		<div class="flex flex-col items-end gap-2 text-right">
			<div class="font-mono text-[10px] text-slate-500">
				Last Snapshot: <span class="text-slate-300">{lastUpdated || '—'}</span>
			</div>
			<button
				onclick={() => fetchAnalytics(true)}
				disabled={isLoading || isRefreshing}
				class="flex cursor-pointer items-center gap-2 rounded border border-[#22D3EE]/30 bg-[#22D3EE]/5 px-3 py-1.5 font-mono text-xs text-[#22D3EE] transition-all hover:border-[#22D3EE] hover:bg-[#22D3EE]/10 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
			>
				<svg
					class="h-3.5 w-3.5 {isRefreshing ? 'animate-spin' : ''}"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
					/>
				</svg>
				{isRefreshing ? 'Refreshing...' : 'Refresh Metrics'}
			</button>
		</div>
	</div>

	{#if isLoading}
		<div class="flex flex-grow items-center justify-center py-24">
			<div class="space-y-4 text-center font-mono text-xs">
				<div
					class="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-slate-800 border-t-[#22D3EE]"
				></div>
				<div class="animate-pulse tracking-widest text-slate-500 uppercase">
					Aggregating Telemetry Stream...
				</div>
			</div>
		</div>
	{:else if errorMessage}
		<div
			class="mx-auto my-12 max-w-md rounded-xl border border-[#E11D48]/30 bg-[#E11D48]/5 p-8 text-center"
		>
			<p class="mb-2 font-mono text-xs tracking-wider text-[#E11D48] uppercase">
				TELEMETRY OFFLINE
			</p>
			<p class="font-sans text-xs text-slate-400">{errorMessage}</p>
			<button
				onclick={() => fetchAnalytics()}
				class="mt-4 rounded-lg border border-[#E11D48]/30 bg-[#E11D48]/5 px-4 py-1.5 font-mono text-[10px] text-[#E11D48] uppercase transition-colors hover:border-[#E11D48] hover:bg-[#E11D48]/10"
			>
				Retry Probe
			</button>
		</div>
	{:else if data}
		<!-- KPI Summary Cards -->
		<div class="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<!-- Total Views Card -->
			<div
				class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
			>
				<div
					class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-[#22D3EE]/40 to-transparent"
				></div>
				<div>
					<span class="font-mono text-[9px] font-bold tracking-widest text-[#22D3EE] uppercase"
						>Reader Engagement</span
					>
					<div class="mt-2 flex items-baseline space-x-2">
						<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
							{data.views.total_views.toLocaleString()}
						</span>
						<span class="font-mono text-[10px] text-slate-500">READS</span>
					</div>
				</div>
				<div
					class="mt-4 border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px] text-slate-400"
				>
					Cumulative Article Views
				</div>
			</div>

			<!-- HTTP Traffic & Error Rate Card -->
			<div
				class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
			>
				<div
					class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-cyan-400/30 to-transparent"
				></div>
				<div>
					<span class="font-mono text-[9px] font-bold tracking-widest text-cyan-400 uppercase"
						>HTTP Traffic</span
					>
					<div class="mt-2 flex items-baseline space-x-2">
						<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
							{data.system.http_requests_total.toLocaleString()}
						</span>
						<span class="font-mono text-[10px] text-slate-500">REQS</span>
					</div>
				</div>
				<div
					class="mt-4 flex items-center justify-between border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px]"
				>
					<span class="text-slate-400">Error Rate:</span>
					<span class="font-bold {Number(httpErrorRate) > 5 ? 'text-[#E11D48]' : 'text-[#22D3EE]'}">
						{httpErrorRate}% ({data.system.http_errors_total} err)
					</span>
				</div>
			</div>

			<!-- Fetcher Health Card -->
			<div
				class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
			>
				<div
					class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-emerald-400/30 to-transparent"
				></div>
				<div>
					<span class="font-mono text-[9px] font-bold tracking-widest text-emerald-400 uppercase"
						>RSS Fetcher</span
					>
					<div class="mt-2 flex items-baseline space-x-2">
						<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
							{data.system.fetch_cycles_total}
						</span>
						<span class="font-mono text-[10px] text-slate-500">CYCLES</span>
					</div>
				</div>
				<div
					class="mt-4 flex items-center justify-between border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px]"
				>
					<span class="text-slate-400">Success Rate:</span>
					<span
						class="font-bold {Number(fetchSuccessRate) < 90
							? 'text-amber-400'
							: 'text-emerald-400'}"
					>
						{fetchSuccessRate}%
					</span>
				</div>
			</div>

			<!-- AI Engine Health Card -->
			<div
				class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
			>
				<div
					class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-purple-400/30 to-transparent"
				></div>
				<div>
					<span class="font-mono text-[9px] font-bold tracking-widest text-purple-400 uppercase"
						>AI Pipeline</span
					>
					<div class="mt-2 flex items-baseline space-x-2">
						<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
							{data.system.ai_calls_total}
						</span>
						<span class="font-mono text-[10px] text-slate-500">CALLS</span>
					</div>
				</div>
				<div
					class="mt-4 flex items-center justify-between border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px]"
				>
					<span class="text-slate-400">AI Success:</span>
					<span class="font-bold text-purple-300">
						{aiSuccessRate}%
					</span>
				</div>
			</div>
		</div>

		<!-- Main 2-Column Analytics Grid -->
		<div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
			<!-- Col 1: Top Most-Viewed Articles -->
			<div
				class="flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6"
			>
				<div>
					<div
						class="mb-4 flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-3"
					>
						<span class="font-mono text-xs font-bold tracking-wider text-[#22D3EE] uppercase"
							>Top Read Articles</span
						>
						<span class="font-mono text-[10px] text-slate-500">By Total Views</span>
					</div>

					{#if data.views.top_articles.length === 0}
						<div class="py-12 text-center font-mono text-xs text-slate-500">
							No article views recorded yet.
						</div>
					{:else}
						<div class="space-y-4 font-mono text-xs">
							{#each data.views.top_articles as art, idx (art.id)}
								<div class="space-y-1.5">
									<div class="flex items-center justify-between gap-3 text-slate-200">
										<div class="flex items-center gap-2 overflow-hidden">
											<span
												class="inline-flex h-4 w-4 items-center justify-center rounded bg-slate-800 text-[9px] font-bold text-slate-400"
											>
												#{idx + 1}
											</span>
											<a
												href="/admin/preview/{art.id}"
												class="truncate font-serif text-sm text-slate-200 transition-colors hover:text-[#22D3EE]"
												title={art.title}
											>
												{art.title}
											</a>
										</div>
										<span class="font-bold whitespace-nowrap text-[#22D3EE]">
											{art.view_count.toLocaleString()} views
										</span>
									</div>
									<!-- Visual Bar -->
									<div class="relative h-1.5 w-full overflow-hidden rounded-full bg-slate-800/80">
										<div
											class="h-full rounded-full bg-gradient-to-r from-[#22D3EE]/40 to-[#22D3EE]"
											style="width: {(art.view_count / maxTopViews) * 100}%"
										></div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>

			<!-- Col 2: Content Breakdown & Value Scoring -->
			<div class="space-y-6">
				<!-- Category Distribution -->
				<div class="rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
					<div
						class="mb-4 flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-3 font-mono text-xs"
					>
						<span class="font-bold tracking-wider text-[#22D3EE] uppercase"
							>Category Distribution</span
						>
						<span class="text-[10px] text-slate-500">Total: {totalCategorizedArticles} Posts</span>
					</div>

					<div class="space-y-3 font-mono text-xs">
						{#each data.content.category_distribution as cat (cat.category)}
							{@const pct = ((cat.count / totalCategorizedArticles) * 100).toFixed(1)}
							<div>
								<div class="mb-1 flex items-center justify-between text-[11px]">
									<span class="text-slate-300 uppercase">{cat.category}</span>
									<span class="text-slate-400">{cat.count} ({pct}%)</span>
								</div>
								<div class="h-1.5 w-full overflow-hidden rounded-full bg-slate-800">
									<div class="h-full rounded-full bg-[#22D3EE]" style="width: {pct}%"></div>
								</div>
							</div>
						{/each}
					</div>
				</div>

				<!-- Value Scoring Breakdown -->
				<div class="rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
					<div
						class="mb-4 flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-3 font-mono text-xs"
					>
						<span class="font-bold tracking-wider text-[#22D3EE] uppercase"
							>Editorial Value Score Distribution</span
						>
						<span class="text-[10px] text-slate-500">Auto-Evaluated</span>
					</div>

					<div class="grid grid-cols-3 gap-3 text-center font-mono">
						<div class="rounded-xl border border-[#22D3EE]/30 bg-[#22D3EE]/5 p-3">
							<span class="block text-[10px] font-bold text-[#22D3EE]">HIGH (≥80)</span>
							<span class="mt-1 block font-serif text-xl font-medium text-white">
								{data.content.score_distribution.high}
							</span>
							<span class="text-[9px] text-slate-500">
								{((data.content.score_distribution.high / totalScoredArticles) * 100).toFixed(0)}%
							</span>
						</div>

						<div class="rounded-xl border border-amber-500/30 bg-amber-500/5 p-3">
							<span class="block text-[10px] font-bold text-amber-400">MED (60–79)</span>
							<span class="mt-1 block font-serif text-xl font-medium text-white">
								{data.content.score_distribution.medium}
							</span>
							<span class="text-[9px] text-slate-500">
								{((data.content.score_distribution.medium / totalScoredArticles) * 100).toFixed(0)}%
							</span>
						</div>

						<div class="rounded-xl border border-slate-700 bg-slate-800/20 p-3">
							<span class="block text-[10px] font-bold text-slate-400">LOW (&lt;60)</span>
							<span class="mt-1 block font-serif text-xl font-medium text-white">
								{data.content.score_distribution.low}
							</span>
							<span class="text-[9px] text-slate-500">
								{((data.content.score_distribution.low / totalScoredArticles) * 100).toFixed(0)}%
							</span>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Technical Server Telemetry Row -->
		<div
			class="mt-8 rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/10 p-6 font-mono text-xs"
		>
			<div
				class="mb-4 flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-3"
			>
				<span class="font-bold tracking-wider text-slate-400 uppercase"
					>Go Server Runtime &amp; Resource Telemetry</span
				>
				<span class="text-[10px] text-[#22D3EE]"
					>Uptime: {formatUptime(data.system.uptime_seconds)}</span
				>
			</div>

			<div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
				<div class="rounded-lg border border-[rgba(255,255,255,0.04)] bg-[#070A10]/50 p-3">
					<span class="block text-[10px] text-slate-500 uppercase">Memory Alloc</span>
					<span class="mt-1 block text-sm font-bold text-slate-200">
						{data.system.memory_alloc_mb} MB
					</span>
				</div>
				<div class="rounded-lg border border-[rgba(255,255,255,0.04)] bg-[#070A10]/50 p-3">
					<span class="block text-[10px] text-slate-500 uppercase">System Memory</span>
					<span class="mt-1 block text-sm font-bold text-slate-200">
						{data.system.memory_sys_mb} MB
					</span>
				</div>
				<div class="rounded-lg border border-[rgba(255,255,255,0.04)] bg-[#070A10]/50 p-3">
					<span class="block text-[10px] text-slate-500 uppercase">Goroutines</span>
					<span class="mt-1 block text-sm font-bold text-slate-200">
						{data.system.num_goroutines}
					</span>
				</div>
				<div class="rounded-lg border border-[rgba(255,255,255,0.04)] bg-[#070A10]/50 p-3">
					<span class="block text-[10px] text-slate-500 uppercase">Avg Response Time</span>
					<span class="mt-1 block text-sm font-bold text-[#22D3EE]">
						{data.system.http_avg_latency_ms} ms
					</span>
				</div>
			</div>
		</div>
	{/if}
</section>
