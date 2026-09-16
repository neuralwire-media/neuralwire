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

	interface PipelineFunnel {
		total_articles: number;
		drafts_count: number;
		scored_count: number;
		published_count: number;
		rejected_count: number;
		conversion_rate: number;
		high_value_ratio: number;
	}

	interface SourceDiagnostic {
		id: number;
		name: string;
		url: string;
		category: string;
		enabled: boolean;
		last_fetched_at?: string;
		last_duration_ms: number;
		last_http_status: number;
		last_error: string;
		consecutive_failures: number;
		total_items_yielded: number;
		created_at: string;
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
			funnel?: PipelineFunnel;
		};
		system: {
			uptime_seconds: number;
			http_requests_total: number;
			http_errors_total: number;
			http_4xx_total?: number;
			http_5xx_total?: number;
			http_status_codes?: Record<string, number>;
			http_avg_latency_ms: number;
			fetch_cycles_total: number;
			fetch_cycles_failed: number;
			ai_calls_total: number;
			ai_calls_failed: number;
			ai_prompt_tokens?: number;
			ai_completion_tokens?: number;
			ai_total_tokens?: number;
			ai_tokens_by_model?: Record<string, number>;
			memory_alloc_mb: number;
			memory_sys_mb: number;
			num_goroutines: number;
		};
		sources?: SourceDiagnostic[];
	}

	interface FormattedStatusCode {
		code: string;
		count: number;
		percentage: string;
		label: string;
		category: '2xx' | '3xx' | '4xx' | '5xx' | 'other';
		badgeClass: string;
	}

	const statusLabelMap: Record<string, string> = {
		'200': 'OK / Success',
		'201': 'Created',
		'204': 'No Content',
		'301': 'Moved Permanently',
		'302': 'Found / Redirect',
		'304': 'Not Modified (Cached)',
		'400': 'Bad Request',
		'401': 'Unauthorized (Auth)',
		'403': 'Forbidden (CSRF/Auth)',
		'404': 'Not Found (Bot/Scan)',
		'429': 'Too Many Requests (Rate Limit)',
		'500': 'Internal Server Error',
		'502': 'Bad Gateway',
		'503': 'Service Unavailable',
		'504': 'Gateway Timeout'
	};

	let activeTab = $state<'traffic' | 'ai' | 'funnel'>('traffic');
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

	function formatDate(dateStr?: string): string {
		if (!dateStr) return 'Never';
		const d = new Date(dateStr);
		return d.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
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

	const server5xxErrors = $derived(data?.system.http_5xx_total ?? 0);
	const client4xxErrors = $derived(
		data?.system.http_4xx_total ??
			(data ? data.system.http_errors_total - (data.system.http_5xx_total ?? 0) : 0)
	);

	const serverErrorRate = $derived(
		data && data.system.http_requests_total > 0
			? ((server5xxErrors / data.system.http_requests_total) * 100).toFixed(2)
			: '0.00'
	);

	const clientErrorRate = $derived(
		data && data.system.http_requests_total > 0
			? ((client4xxErrors / data.system.http_requests_total) * 100).toFixed(2)
			: '0.00'
	);

	const sortedStatusCodes = $derived.by<FormattedStatusCode[]>(() => {
		if (!data || !data.system.http_status_codes) return [];
		const total = data.system.http_requests_total || 1;
		const entries = Object.entries(data.system.http_status_codes);

		return entries
			.map(([code, count]) => {
				const numCode = parseInt(code, 10);
				let category: '2xx' | '3xx' | '4xx' | '5xx' | 'other' = 'other';
				let badgeClass = 'text-slate-300 border-slate-700 bg-slate-800/40';

				if (numCode >= 200 && numCode < 300) {
					category = '2xx';
					badgeClass = 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10';
				} else if (numCode >= 300 && numCode < 400) {
					category = '3xx';
					badgeClass = 'text-cyan-400 border-cyan-500/30 bg-cyan-500/10';
				} else if (numCode >= 400 && numCode < 500) {
					category = '4xx';
					badgeClass = 'text-amber-400 border-amber-500/30 bg-amber-500/10';
				} else if (numCode >= 500) {
					category = '5xx';
					badgeClass = 'text-rose-400 border-rose-500/30 bg-rose-500/10';
				}

				return {
					code,
					count,
					percentage: ((count / total) * 100).toFixed(1),
					label: statusLabelMap[code] || `HTTP ${code}`,
					category,
					badgeClass
				};
			})
			.sort((a, b) => b.count - a.count);
	});

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

	// Estimated compute cost (approx $0.0002 per 1K blended tokens for small models like gpt-4o-mini / deepseek-chat)
	const estimatedAICost = $derived(
		data && (data.system.ai_total_tokens ?? 0) > 0
			? (((data.system.ai_total_tokens ?? 0) / 1000) * 0.0002).toFixed(4)
			: '0.0000'
	);

	const totalPromptTokens = $derived(data?.system.ai_prompt_tokens ?? 0);
	const totalCompletionTokens = $derived(data?.system.ai_completion_tokens ?? 0);
	const totalTokens = $derived(
		data?.system.ai_total_tokens ?? totalPromptTokens + totalCompletionTokens
	);

	const modelsTokensList = $derived.by(() => {
		if (!data || !data.system.ai_tokens_by_model) return [];
		return Object.entries(data.system.ai_tokens_by_model)
			.map(([model, count]) => ({
				model,
				count,
				percentage: totalTokens > 0 ? ((count / totalTokens) * 100).toFixed(1) : '0.0'
			}))
			.sort((a, b) => b.count - a.count);
	});

	onMount(() => {
		fetchAnalytics();
	});
</script>

<svelte:head>
	<title>Analytics &amp; System Telemetry | NeuralWire Admin</title>
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

	<!-- Sub-Tab Navigation Switcher -->
	<div class="mb-8 flex flex-wrap gap-2 border-b border-[rgba(255,255,255,0.06)] pb-4">
		<button
			type="button"
			onclick={() => (activeTab = 'traffic')}
			class="flex items-center gap-2 rounded-lg border px-4 py-2 font-mono text-xs font-bold tracking-wider uppercase transition-all {activeTab ===
			'traffic'
				? 'border-[#22D3EE]/50 bg-[#22D3EE]/10 text-[#22D3EE] shadow-[0_0_15px_rgba(34,211,238,0.15)]'
				: 'border-[rgba(255,255,255,0.06)] bg-[#0F172A]/40 text-slate-400 hover:border-[rgba(255,255,255,0.15)] hover:text-white'}"
		>
			<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
				/>
			</svg>
			<span>Traffic &amp; System</span>
		</button>

		<button
			type="button"
			onclick={() => (activeTab = 'ai')}
			class="flex items-center gap-2 rounded-lg border px-4 py-2 font-mono text-xs font-bold tracking-wider uppercase transition-all {activeTab ===
			'ai'
				? 'border-purple-500/50 bg-purple-500/10 text-purple-300 shadow-[0_0_15px_rgba(168,85,247,0.15)]'
				: 'border-[rgba(255,255,255,0.06)] bg-[#0F172A]/40 text-slate-400 hover:border-[rgba(255,255,255,0.15)] hover:text-white'}"
		>
			<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M13 10V3L4 14h7v7l9-11h-7z"
				/>
			</svg>
			<span>AI &amp; Cost Telemetry</span>
		</button>

		<button
			type="button"
			onclick={() => (activeTab = 'funnel')}
			class="flex items-center gap-2 rounded-lg border px-4 py-2 font-mono text-xs font-bold tracking-wider uppercase transition-all {activeTab ===
			'funnel'
				? 'border-emerald-500/50 bg-emerald-500/10 text-emerald-300 shadow-[0_0_15px_rgba(16,185,129,0.15)]'
				: 'border-[rgba(255,255,255,0.06)] bg-[#0F172A]/40 text-slate-400 hover:border-[rgba(255,255,255,0.15)] hover:text-white'}"
		>
			<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
				/>
			</svg>
			<span>Feed Sources &amp; Funnel</span>
		</button>
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
		<!-- TAB 1: Traffic & System Telemetry -->
		{#if activeTab === 'traffic'}
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

				<!-- HTTP Traffic & Server Health Card -->
				<div
					class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
				>
					<div
						class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-cyan-400/30 to-transparent"
					></div>
					<div>
						<div class="flex items-center justify-between">
							<span class="font-mono text-[9px] font-bold tracking-widest text-cyan-400 uppercase"
								>HTTP Traffic</span
							>
							<span
								class="rounded border px-1.5 py-0.5 font-mono text-[8px] font-bold {server5xxErrors >
								0
									? 'border-rose-500/40 bg-rose-500/10 text-rose-400'
									: 'border-emerald-500/40 bg-emerald-500/10 text-emerald-400'}"
							>
								{server5xxErrors > 0 ? `${server5xxErrors} 5XX FAULT` : 'HEALTHY'}
							</span>
						</div>
						<div class="mt-2 flex items-baseline space-x-2">
							<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
								{data.system.http_requests_total.toLocaleString()}
							</span>
							<span class="font-mono text-[10px] text-slate-500">REQS</span>
						</div>
					</div>
					<div
						class="mt-4 flex flex-col gap-1 border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px]"
					>
						<div class="flex items-center justify-between">
							<span class="text-slate-400">Server 5xx Error:</span>
							<span class="font-bold {server5xxErrors > 0 ? 'text-[#E11D48]' : 'text-emerald-400'}">
								{serverErrorRate}% ({server5xxErrors} err)
							</span>
						</div>
						<div class="flex items-center justify-between text-[9px] text-slate-500">
							<span>Client 4xx / Probes:</span>
							<span class="font-medium text-slate-400">
								{client4xxErrors.toLocaleString()} reqs ({clientErrorRate}%)
							</span>
						</div>
					</div>
				</div>

				<!-- Content Volume Card -->
				<div
					class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
				>
					<div
						class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-emerald-400/30 to-transparent"
					></div>
					<div>
						<span class="font-mono text-[9px] font-bold tracking-widest text-emerald-400 uppercase"
							>Published Archive</span
						>
						<div class="mt-2 flex items-baseline space-x-2">
							<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
								{data.content.published_count.toLocaleString()}
							</span>
							<span class="font-mono text-[10px] text-slate-500">LIVE</span>
						</div>
					</div>
					<div
						class="mt-4 flex items-center justify-between border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px]"
					>
						<span class="text-slate-400">Drafts in Queue:</span>
						<span class="font-bold text-amber-400">{data.content.drafts_count}</span>
					</div>
				</div>

				<!-- Runtime Node Health Card -->
				<div
					class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
				>
					<div
						class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-purple-400/30 to-transparent"
					></div>
					<div>
						<span class="font-mono text-[9px] font-bold tracking-widest text-purple-400 uppercase"
							>Node Latency</span
						>
						<div class="mt-2 flex items-baseline space-x-2">
							<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
								{data.system.http_avg_latency_ms}
							</span>
							<span class="font-mono text-[10px] text-slate-500">MS</span>
						</div>
					</div>
					<div
						class="mt-4 flex items-center justify-between border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px]"
					>
						<span class="text-slate-400">Memory Alloc:</span>
						<span class="font-bold text-purple-300">{data.system.memory_alloc_mb} MB</span>
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
							<span class="text-[10px] text-slate-500">Total: {totalCategorizedArticles} Posts</span
							>
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
									{((data.content.score_distribution.medium / totalScoredArticles) * 100).toFixed(
										0
									)}%
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

				<!-- HTTP Status Code Breakdown Grid -->
				<div class="mt-6 border-t border-[rgba(255,255,255,0.06)] pt-4">
					<div class="mb-3 flex items-center justify-between">
						<span class="text-[10px] font-bold tracking-wider text-slate-400 uppercase"
							>HTTP Response Status Code Breakdown</span
						>
						<span class="text-[10px] text-slate-500">
							{sortedStatusCodes.length} status {sortedStatusCodes.length === 1 ? 'code' : 'codes'} recorded
						</span>
					</div>

					{#if sortedStatusCodes.length === 0}
						<div class="py-3 text-center text-[11px] text-slate-500">
							No granular HTTP status telemetry available yet.
						</div>
					{:else}
						<div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
							{#each sortedStatusCodes as item (item.code)}
								<div
									class="flex items-center justify-between rounded-lg border border-[rgba(255,255,255,0.04)] bg-[#070A10]/40 px-3 py-2.5"
								>
									<div class="flex items-center gap-2.5">
										<span
											class="rounded border px-1.5 py-0.5 font-mono text-[10px] font-bold {item.badgeClass}"
										>
											{item.code}
										</span>
										<div class="flex flex-col">
											<span class="text-[11px] font-medium text-slate-200">{item.label}</span>
											<span class="text-[9px] text-slate-500"
												>{item.category === '4xx'
													? 'Client / Bot Probe'
													: item.category === '5xx'
														? 'Server Error'
														: item.category === '2xx'
															? 'Success'
															: item.category === '3xx'
																? 'Redirect / Cache'
																: 'Other'}</span
											>
										</div>
									</div>
									<div class="text-right">
										<span class="text-xs font-bold text-white">{item.count.toLocaleString()}</span>
										<span class="block text-[9px] text-slate-400">{item.percentage}%</span>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>

			<!-- TAB 2: AI & Cost Telemetry -->
		{:else if activeTab === 'ai'}
			<div class="space-y-8">
				<!-- Top AI KPI Grid -->
				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
					<!-- AI Calls & Success Rate -->
					<div
						class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
					>
						<div
							class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-purple-500/40 to-transparent"
						></div>
						<div>
							<div class="flex items-center justify-between">
								<span
									class="font-mono text-[9px] font-bold tracking-widest text-purple-400 uppercase"
									>AI Requests</span
								>
								<span
									class="rounded border border-purple-500/30 bg-purple-500/10 px-1.5 py-0.5 font-mono text-[8px] font-bold text-purple-300"
								>
									{aiSuccessRate}% SUCCESS
								</span>
							</div>
							<div class="mt-2 flex items-baseline space-x-2">
								<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
									{data.system.ai_calls_total.toLocaleString()}
								</span>
								<span class="font-mono text-[10px] text-slate-500">CALLS</span>
							</div>
						</div>
						<div
							class="mt-4 border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px] text-slate-400"
						>
							Failed calls: <span class="font-bold text-rose-400"
								>{data.system.ai_calls_failed}</span
							>
						</div>
					</div>

					<!-- Prompt Tokens -->
					<div
						class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
					>
						<div
							class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-cyan-400/30 to-transparent"
						></div>
						<div>
							<span class="font-mono text-[9px] font-bold tracking-widest text-cyan-400 uppercase"
								>Prompt Tokens</span
							>
							<div class="mt-2 flex items-baseline space-x-2">
								<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
									{totalPromptTokens.toLocaleString()}
								</span>
								<span class="font-mono text-[10px] text-slate-500">IN</span>
							</div>
						</div>
						<div
							class="mt-4 border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px] text-slate-400"
						>
							Input Context Transmitted
						</div>
					</div>

					<!-- Completion Tokens -->
					<div
						class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
					>
						<div
							class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-emerald-400/30 to-transparent"
						></div>
						<div>
							<span
								class="font-mono text-[9px] font-bold tracking-widest text-emerald-400 uppercase"
								>Generated Tokens</span
							>
							<div class="mt-2 flex items-baseline space-x-2">
								<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
									{totalCompletionTokens.toLocaleString()}
								</span>
								<span class="font-mono text-[10px] text-slate-500">OUT</span>
							</div>
						</div>
						<div
							class="mt-4 border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px] text-slate-400"
						>
							Summaries &amp; Scores Generated
						</div>
					</div>

					<!-- Total Compute Burn -->
					<div
						class="group glow-hover relative flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/30 p-5"
					>
						<div
							class="absolute top-0 left-0 h-[1px] w-full bg-gradient-to-r from-transparent via-amber-400/30 to-transparent"
						></div>
						<div>
							<span class="font-mono text-[9px] font-bold tracking-widest text-amber-400 uppercase"
								>Total Token Burn</span
							>
							<div class="mt-2 flex items-baseline space-x-2">
								<span class="font-serif text-3xl font-medium tracking-tight text-white sm:text-4xl">
									{totalTokens.toLocaleString()}
								</span>
								<span class="font-mono text-[10px] text-slate-500">TOKENS</span>
							</div>
						</div>
						<div
							class="mt-4 flex items-center justify-between border-t border-[rgba(255,255,255,0.04)] pt-2 font-mono text-[10px]"
						>
							<span class="text-slate-400">Est. Cost:</span>
							<span class="font-bold text-amber-300">${estimatedAICost} USD</span>
						</div>
					</div>
				</div>

				<!-- AI Compute Details Grid -->
				<div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
					<!-- Model Distribution -->
					<div class="rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
						<div
							class="mb-4 flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-3 font-mono text-xs"
						>
							<span class="font-bold tracking-wider text-purple-400 uppercase"
								>Model Consumption Breakdown</span
							>
							<span class="text-[10px] text-slate-500">{modelsTokensList.length} Active Models</span
							>
						</div>

						{#if modelsTokensList.length === 0}
							<div class="py-10 text-center font-mono text-xs text-slate-500">
								No model-specific token streams recorded yet.
							</div>
						{:else}
							<div class="space-y-4 font-mono text-xs">
								{#each modelsTokensList as item (item.model)}
									<div>
										<div class="mb-1 flex items-center justify-between">
											<span class="font-bold text-slate-200">{item.model}</span>
											<span class="text-purple-300"
												>{item.count.toLocaleString()} tokens ({item.percentage}%)</span
											>
										</div>
										<div class="h-2 w-full overflow-hidden rounded-full bg-slate-800">
											<div
												class="h-full rounded-full bg-gradient-to-r from-purple-500 to-cyan-400"
												style="width: {item.percentage}%"
											></div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>

					<!-- Cost & Performance Efficiency Card -->
					<div
						class="flex flex-col justify-between rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6"
					>
						<div>
							<div
								class="mb-4 flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-3 font-mono text-xs"
							>
								<span class="font-bold tracking-wider text-[#22D3EE] uppercase"
									>Compute Economics &amp; Efficiency</span
								>
								<span class="text-[10px] text-slate-500">Cost Profile</span>
							</div>

							<div class="space-y-4 font-mono text-xs">
								<div class="rounded-xl border border-[rgba(255,255,255,0.06)] bg-[#070A10]/50 p-4">
									<div class="flex items-center justify-between text-slate-300">
										<span>Avg Tokens per AI Call</span>
										<span class="font-bold text-white">
											{data.system.ai_calls_total > 0
												? Math.round(totalTokens / data.system.ai_calls_total).toLocaleString()
												: 0} tokens
										</span>
									</div>
								</div>

								<div class="rounded-xl border border-[rgba(255,255,255,0.06)] bg-[#070A10]/50 p-4">
									<div class="flex items-center justify-between text-slate-300">
										<span>Cumulative Estimated Burn</span>
										<span class="font-bold text-amber-300">${estimatedAICost} USD</span>
									</div>
									<p class="mt-1 text-[10px] text-slate-500">
										Computed at ~$0.0002/1k blended tokens across summarizer, categorization &amp;
										valuation stages.
									</p>
								</div>

								<div class="rounded-xl border border-[rgba(255,255,255,0.06)] bg-[#070A10]/50 p-4">
									<div class="flex items-center justify-between text-slate-300">
										<span>AI Pipeline Status</span>
										<span class="font-bold text-emerald-400">OPTIMAL / LOW LATENCY</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- TAB 3: Feed Sources & Funnel -->
		{:else if activeTab === 'funnel'}
			<div class="space-y-8">
				<!-- Pipeline Conversion Funnel -->
				<div class="rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
					<div
						class="mb-6 flex flex-col justify-between gap-2 border-b border-[rgba(255,255,255,0.06)] pb-4 sm:flex-row sm:items-center"
					>
						<div>
							<h3 class="font-mono text-xs font-bold tracking-wider text-emerald-400 uppercase">
								Auto-Publish Conversion Funnel
							</h3>
							<p class="font-sans text-xs text-slate-400">
								Transmission pipeline efficiency from raw RSS ingestion to live dispatch publication
							</p>
						</div>
						<div class="flex flex-wrap items-center gap-2.5">
							<span
								class="rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3 py-1 font-mono text-[11px] font-bold text-emerald-400"
							>
								{data.content.funnel?.conversion_rate ?? 0}% CONVERSION
							</span>
							<span
								class="rounded-full border border-cyan-500/30 bg-cyan-500/10 px-3 py-1 font-mono text-[11px] font-bold text-[#22D3EE]"
							>
								{fetchSuccessRate}% POLLING SUCCESS
							</span>
						</div>
					</div>

					<!-- Funnel Stage Progression -->
					<div class="grid grid-cols-1 gap-4 font-mono text-xs sm:grid-cols-2 lg:grid-cols-4">
						<!-- Stage 1: Ingested -->
						<div class="rounded-xl border border-slate-700 bg-[#070A10]/60 p-4">
							<span class="text-[10px] font-bold text-slate-400 uppercase">Stage 1: Ingested</span>
							<div class="mt-2 font-serif text-2xl font-bold text-white">
								{data.content.funnel?.total_articles ?? 0}
							</div>
							<p class="mt-1 text-[10px] text-slate-500">All captured articles across sources</p>
						</div>

						<!-- Stage 2: Scored & Classified -->
						<div class="rounded-xl border border-purple-500/30 bg-[#070A10]/60 p-4">
							<span class="text-[10px] font-bold text-purple-400 uppercase"
								>Stage 2: AI Evaluated</span
							>
							<div class="mt-2 font-serif text-2xl font-bold text-purple-300">
								{data.content.funnel?.scored_count ?? 0}
							</div>
							<p class="mt-1 text-[10px] text-slate-500">Passed quality gate &amp; value rating</p>
						</div>

						<!-- Stage 3: Published -->
						<div class="rounded-xl border border-emerald-500/30 bg-[#070A10]/60 p-4">
							<span class="text-[10px] font-bold text-emerald-400 uppercase"
								>Stage 3: Published</span
							>
							<div class="mt-2 font-serif text-2xl font-bold text-emerald-300">
								{data.content.funnel?.published_count ?? 0}
							</div>
							<p class="mt-1 text-[10px] text-slate-500">Live on public chronicle feed</p>
						</div>

						<!-- Stage 4: Queue / Rejected -->
						<div class="rounded-xl border border-amber-500/30 bg-[#070A10]/60 p-4">
							<span class="text-[10px] font-bold text-amber-400 uppercase"
								>Drafts &amp; Filtered</span
							>
							<div class="mt-2 font-serif text-2xl font-bold text-amber-300">
								{data.content.funnel?.drafts_count ?? 0}
							</div>
							<p class="mt-1 text-[10px] text-slate-500">
								Rejected: {data.content.funnel?.rejected_count ?? 0}
							</p>
						</div>
					</div>
				</div>

				<!-- Live RSS Sources Health & Diagnostics Grid -->
				<div class="rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 p-6">
					<div
						class="mb-4 flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-3 font-mono text-xs"
					>
						<span class="font-bold tracking-wider text-[#22D3EE] uppercase"
							>Feed Sources Health Diagnostics</span
						>
						<span class="text-[10px] text-slate-500">
							{data.sources?.length ?? 0} sources configured
						</span>
					</div>

					{#if !data.sources || data.sources.length === 0}
						<div class="py-12 text-center font-mono text-xs text-slate-500">
							No RSS sources found in database.
						</div>
					{:else}
						<div class="overflow-x-auto">
							<table class="w-full text-left font-mono text-xs">
								<thead>
									<tr
										class="border-b border-[rgba(255,255,255,0.08)] text-[10px] text-slate-400 uppercase"
									>
										<th class="pb-3 font-medium">Source Name</th>
										<th class="pb-3 font-medium">Category</th>
										<th class="pb-3 font-medium">Health Status</th>
										<th class="pb-3 font-medium">Latency</th>
										<th class="pb-3 font-medium">HTTP Code</th>
										<th class="pb-3 font-medium">Yielded Articles</th>
										<th class="pb-3 font-medium">Last Fetched</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-[rgba(255,255,255,0.04)]">
									{#each data.sources as src (src.id)}
										{@const isFailing = src.consecutive_failures >= 3}
										{@const isDegraded =
											src.consecutive_failures > 0 && src.consecutive_failures < 3}
										<tr class="transition-colors hover:bg-slate-800/20">
											<td class="py-3.5 pr-4">
												<div class="font-bold text-white">{src.name}</div>
												<div class="max-w-xs truncate text-[10px] text-slate-500">{src.url}</div>
											</td>
											<td class="py-3.5 pr-4">
												<span
													class="rounded bg-[#070A10] px-2 py-0.5 text-[10px] text-[#22D3EE] uppercase"
												>
													{src.category}
												</span>
											</td>
											<td class="py-3.5 pr-4">
												{#if !src.enabled}
													<span
														class="rounded border border-slate-700 bg-slate-800/40 px-2 py-0.5 text-[10px] text-slate-400"
													>
														PAUSED
													</span>
												{:else if isFailing}
													<span
														class="rounded border border-rose-500/40 bg-rose-500/10 px-2 py-0.5 text-[10px] font-bold text-rose-400"
														title={src.last_error || 'Polling errors'}
													>
														FAILING ({src.consecutive_failures}x)
													</span>
												{:else if isDegraded}
													<span
														class="rounded border border-amber-500/40 bg-amber-500/10 px-2 py-0.5 text-[10px] font-bold text-amber-400"
														title={src.last_error}
													>
														DEGRADED
													</span>
												{:else}
													<span
														class="rounded border border-emerald-500/40 bg-emerald-500/10 px-2 py-0.5 text-[10px] font-bold text-emerald-400"
													>
														HEALTHY
													</span>
												{/if}
											</td>
											<td class="py-3.5 pr-4 text-slate-300">
												{src.last_duration_ms > 0 ? `${src.last_duration_ms}ms` : '—'}
											</td>
											<td class="py-3.5 pr-4">
												{#if src.last_http_status > 0}
													<span
														class="text-[11px] {src.last_http_status >= 400
															? 'text-rose-400'
															: 'text-slate-300'}"
													>
														{src.last_http_status}
													</span>
												{:else}
													<span class="text-slate-500">—</span>
												{/if}
											</td>
											<td class="py-3.5 pr-4 font-bold text-[#22D3EE]">
												{src.total_items_yielded}
											</td>
											<td class="py-3.5 text-[11px] text-slate-400">
												{formatDate(src.last_fetched_at)}
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	{/if}
</section>
