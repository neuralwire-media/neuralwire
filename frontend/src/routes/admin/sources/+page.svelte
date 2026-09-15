<script lang="ts">
	import { onMount } from 'svelte';
	import { BASE_URL, getCategories } from '$lib/api';
	import type { Category } from '$lib/mockData';

	interface RSSSource {
		id: number;
		name: string;
		url: string;
		category: string;
		enabled: boolean;
		last_fetched_at: string | null;
		created_at: string;
	}

	interface ProbeItem {
		title: string;
		link: string;
		published_at?: string;
	}

	interface ProbeResult {
		valid: boolean;
		title?: string;
		description?: string;
		item_count?: number;
		items?: ProbeItem[];
		error?: string;
	}

	let sources = $state<RSSSource[]>([]);
	let categories = $state<Category[]>([]);
	let isLoading = $state(true);
	let errorMessage = $state('');
	let selectedCategory = $state('all');
	let searchQuery = $state('');

	// Toast state
	let toast = $state<{ type: 'success' | 'error'; title: string; message: string } | null>(null);
	let toastTimer: ReturnType<typeof setTimeout> | undefined;

	function showToast(type: 'success' | 'error', message: string, title?: string) {
		toast = {
			type,
			message,
			title: title ?? (type === 'success' ? 'Success' : 'Error')
		};
		if (toastTimer) clearTimeout(toastTimer);
		toastTimer = setTimeout(() => {
			toast = null;
		}, 5000);
	}

	// Modal States
	let showEditModal = $state(false);
	let isEditing = $state(false);
	let modalSourceId = $state<number | null>(null);
	let formName = $state('');
	let formURL = $state('');
	let formCategory = $state('ai');
	let formEnabled = $state(true);
	let isSaving = $state(false);

	// In-modal probe state
	let isProbing = $state(false);
	let probeResult = $state<ProbeResult | null>(null);

	// Delete confirmation modal state
	let showDeleteModal = $state(false);
	let sourceToDelete = $state<RSSSource | null>(null);
	let isDeleting = $state(false);

	async function loadSources() {
		const token = localStorage.getItem('admin_token');
		if (!token) return;
		isLoading = true;
		errorMessage = '';

		try {
			const [resSources, resCats] = await Promise.all([
				fetch(`${BASE_URL}/admin/sources`, {
					headers: {
						Authorization: `Bearer ${token}`,
						Accept: 'application/json'
					}
				}),
				getCategories()
			]);

			categories = resCats;

			if (resSources.ok) {
				const json = await resSources.json();
				sources = json.data || [];
			} else {
				if (resSources.status === 401 || resSources.status === 403) {
					localStorage.removeItem('admin_token');
					window.location.reload();
					return;
				}
				errorMessage = 'Failed to load RSS sources from backend.';
			}
		} catch (err) {
			console.error('load sources error:', err);
			errorMessage = 'Connection failure to server node.';
		} finally {
			isLoading = false;
		}
	}

	async function handleToggle(src: RSSSource) {
		const token = localStorage.getItem('admin_token');
		if (!token) return;

		const targetState = !src.enabled;
		// Optimistic UI update
		src.enabled = targetState;

		try {
			const res = await fetch(`${BASE_URL}/admin/sources/${src.id}/toggle`, {
				method: 'PATCH',
				headers: {
					Authorization: `Bearer ${token}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ enabled: targetState })
			});

			if (res.ok) {
				showToast(
					'success',
					`Feed "${src.name}" is now ${targetState ? 'ACTIVE' : 'DISABLED'}.`,
					'Status Updated'
				);
			} else {
				// Revert on failure
				src.enabled = !targetState;
				showToast('error', 'Failed to update feed state.', 'Update Failed');
			}
		} catch (err) {
			console.error('toggle error:', err);
			src.enabled = !targetState;
			showToast('error', 'Network error updating feed state.', 'Update Failed');
		}
	}

	function openAddModal() {
		isEditing = false;
		modalSourceId = null;
		formName = '';
		formURL = '';
		formCategory = categories.length > 0 ? categories[0].slug : 'ai';
		formEnabled = true;
		probeResult = null;
		showEditModal = true;
	}

	function openEditModal(src: RSSSource) {
		isEditing = true;
		modalSourceId = src.id;
		formName = src.name;
		formURL = src.url;
		formCategory = src.category;
		formEnabled = src.enabled;
		probeResult = null;
		showEditModal = true;
	}

	function openProbeForSource(src: RSSSource) {
		openEditModal(src);
		probeFeedURL();
	}

	async function probeFeedURL() {
		const urlToTest = formURL.trim();
		if (!urlToTest) {
			probeResult = { valid: false, error: 'Please enter a valid feed URL first.' };
			return;
		}

		const token = localStorage.getItem('admin_token');
		if (!token) return;

		isProbing = true;
		probeResult = null;

		try {
			const res = await fetch(`${BASE_URL}/admin/sources/test`, {
				method: 'POST',
				headers: {
					Authorization: `Bearer ${token}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ url: urlToTest })
			});

			if (res.ok) {
				probeResult = await res.json();
				if (probeResult && probeResult.valid && !formName && probeResult.title) {
					formName = probeResult.title;
				}
			} else {
				const errData = await res.json().catch(() => ({}));
				probeResult = {
					valid: false,
					error: errData.error || 'Server rejected probe request.'
				};
			}
		} catch (err) {
			console.error('probe error:', err);
			probeResult = {
				valid: false,
				error: 'Network timeout probing target RSS feed.'
			};
		} finally {
			isProbing = false;
		}
	}

	async function saveSource() {
		const token = localStorage.getItem('admin_token');
		if (!token) return;

		if (!formName.trim()) {
			showToast('error', 'Feed name cannot be empty.', 'Validation Error');
			return;
		}
		if (!formURL.trim()) {
			showToast('error', 'Feed URL cannot be empty.', 'Validation Error');
			return;
		}

		isSaving = true;
		try {
			const payload = {
				name: formName.trim(),
				url: formURL.trim(),
				category: formCategory.trim() || 'ai',
				enabled: formEnabled
			};

			let res: Response;
			if (isEditing && modalSourceId) {
				res = await fetch(`${BASE_URL}/admin/sources/${modalSourceId}`, {
					method: 'PUT',
					headers: {
						Authorization: `Bearer ${token}`,
						'Content-Type': 'application/json'
					},
					body: JSON.stringify(payload)
				});
			} else {
				res = await fetch(`${BASE_URL}/admin/sources`, {
					method: 'POST',
					headers: {
						Authorization: `Bearer ${token}`,
						'Content-Type': 'application/json'
					},
					body: JSON.stringify(payload)
				});
			}

			if (res.ok) {
				showToast(
					'success',
					isEditing ? `Source "${formName}" updated.` : `Source "${formName}" added.`,
					'Source Saved'
				);
				showEditModal = false;
				await loadSources();
			} else {
				const errData = await res.json().catch(() => ({}));
				showToast('error', errData.error || 'Failed to save feed source.', 'Save Error');
			}
		} catch (err) {
			console.error('save source error:', err);
			showToast('error', 'Failed to communicate with the server.', 'Save Error');
		} finally {
			isSaving = false;
		}
	}

	function confirmDelete(src: RSSSource) {
		sourceToDelete = src;
		showDeleteModal = true;
	}

	async function executeDelete() {
		if (!sourceToDelete) return;
		const token = localStorage.getItem('admin_token');
		if (!token) return;

		isDeleting = true;
		try {
			const res = await fetch(`${BASE_URL}/admin/sources/${sourceToDelete.id}`, {
				method: 'DELETE',
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (res.ok) {
				showToast('success', `Source "${sourceToDelete.name}" deleted.`, 'Source Deleted');
				showDeleteModal = false;
				sourceToDelete = null;
				await loadSources();
			} else {
				const errData = await res.json().catch(() => ({}));
				showToast('error', errData.error || 'Failed to delete source.', 'Delete Error');
			}
		} catch (err) {
			console.error('delete source error:', err);
			showToast('error', 'Failed to communicate with the server.', 'Delete Error');
		} finally {
			isDeleting = false;
		}
	}

	function formatDate(dt: string | null): string {
		if (!dt) return 'Never Polled';
		try {
			const d = new Date(dt);
			return d.toLocaleString('en-US', {
				month: 'short',
				day: 'numeric',
				hour: '2-digit',
				minute: '2-digit'
			});
		} catch {
			return dt;
		}
	}

	// Reactive derived lists & statistics
	const totalSourcesCount = $derived(sources.length);
	const activeSourcesCount = $derived(sources.filter((s) => s.enabled).length);
	const disabledSourcesCount = $derived(sources.filter((s) => !s.enabled).length);

	const availableCategories = $derived([
		'all',
		...Array.from(new Set(sources.map((s) => s.category.toLowerCase()))).sort()
	]);

	const filteredSources = $derived(
		sources.filter((src) => {
			const matchesCategory =
				selectedCategory === 'all' || src.category.toLowerCase() === selectedCategory.toLowerCase();
			const query = searchQuery.trim().toLowerCase();
			const matchesSearch =
				!query ||
				src.name.toLowerCase().includes(query) ||
				src.url.toLowerCase().includes(query) ||
				src.category.toLowerCase().includes(query);
			return matchesCategory && matchesSearch;
		})
	);

	onMount(() => {
		loadSources();
	});
</script>

<svelte:head>
	<title>RSS Feed Sources Manager | NeuralWire Admin</title>
</svelte:head>

<!-- Toast Notifications -->
{#if toast}
	<div
		class="animate-slide-in fixed top-4 right-4 z-50 flex max-w-sm items-start gap-3 rounded-xl border p-4 font-mono text-xs shadow-2xl"
		style="border-color: {toast.type === 'success' ? '#22D3EE' : '#E11D48'}; background: #0F172A;"
	>
		<div
			class="mt-0.5 h-2 w-2 flex-shrink-0 rounded-full"
			style="background: {toast.type === 'success'
				? '#22D3EE'
				: '#E11D48'}; box-shadow: 0 0 8px {toast.type === 'success' ? '#22D3EE' : '#E11D48'};"
		></div>
		<div class="flex-grow space-y-1">
			<div
				class="font-bold tracking-widest uppercase"
				style="color: {toast.type === 'success' ? '#22D3EE' : '#E11D48'};"
			>
				{toast.title}
			</div>
			<div class="leading-relaxed text-slate-300">{toast.message}</div>
		</div>
		<button
			onclick={() => (toast = null)}
			class="cursor-pointer text-slate-500 transition-colors hover:text-white"
		>
			[X]
		</button>
	</div>
{/if}

<section
	class="mx-auto flex max-w-7xl flex-grow flex-col justify-start px-4 py-12 sm:px-6 md:py-16 lg:px-8"
>
	<!-- Header & Telemetry Badges -->
	<div
		class="mb-8 flex flex-col items-start justify-between gap-4 border-b border-[rgba(255,255,255,0.08)] pb-6 sm:flex-row sm:items-end"
	>
		<div>
			<span class="tag-mono mb-1 block text-xs font-bold tracking-widest text-[#22D3EE]"
				>Feed Ingestion Engine</span
			>
			<h1 class="font-serif text-3xl font-medium text-white">RSS FEED SOURCES</h1>
		</div>
		<div class="flex flex-wrap items-center gap-3">
			<div class="flex items-center gap-2 font-mono text-xs text-slate-400">
				<span class="rounded bg-slate-800/80 px-2 py-1 text-[11px] text-slate-300"
					>{totalSourcesCount} Total</span
				>
				<span class="rounded bg-[#22D3EE]/10 px-2 py-1 text-[11px] font-bold text-[#22D3EE]"
					>{activeSourcesCount} Active</span
				>
				{#if disabledSourcesCount > 0}
					<span class="rounded bg-slate-800 px-2 py-1 text-[11px] text-slate-500"
						>{disabledSourcesCount} Disabled</span
					>
				{/if}
			</div>
			<button
				onclick={openAddModal}
				class="flex cursor-pointer items-center gap-1.5 rounded-lg border border-[#22D3EE]/40 bg-[#22D3EE]/10 px-3.5 py-1.5 font-mono text-xs font-bold text-[#22D3EE] transition-all hover:border-[#22D3EE] hover:bg-[#22D3EE]/20 hover:text-white"
			>
				<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"
					></path>
				</svg>
				Add RSS Source
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
					Loading Feed Registry...
				</div>
			</div>
		</div>
	{:else if errorMessage}
		<div
			class="mx-auto my-12 max-w-md rounded-xl border border-[#E11D48]/30 bg-[#E11D48]/5 p-8 text-center"
		>
			<p class="mb-2 font-mono text-xs tracking-wider text-[#E11D48] uppercase">REGISTRY ERROR</p>
			<p class="font-sans text-xs text-slate-400">{errorMessage}</p>
			<button
				onclick={() => loadSources()}
				class="mt-4 rounded-lg border border-[#E11D48]/30 bg-[#E11D48]/5 px-4 py-1.5 font-mono text-[10px] text-[#E11D48] uppercase transition-colors hover:border-[#E11D48] hover:bg-[#E11D48]/10"
			>
				Retry Query
			</button>
		</div>
	{:else}
		<!-- Controls: Category Pills & Search -->
		<div class="mb-6 flex flex-col items-stretch justify-between gap-4 md:flex-row md:items-center">
			<!-- Category Pills -->
			<div class="flex flex-wrap items-center gap-1.5">
				{#each availableCategories as cat}
					<button
						onclick={() => (selectedCategory = cat)}
						class="cursor-pointer rounded-lg px-3 py-1 font-mono text-xs uppercase transition-colors {selectedCategory ===
						cat
							? 'bg-[#22D3EE]/15 font-bold text-[#22D3EE] ring-1 ring-[#22D3EE]/40'
							: 'bg-[#0F172A]/40 text-slate-400 hover:bg-[#0F172A] hover:text-white'}"
					>
						{cat}
					</button>
				{/each}
			</div>

			<!-- Search Bar -->
			<div class="relative min-w-[240px]">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search feeds by name or URL..."
					class="w-full rounded-lg border border-[rgba(255,255,255,0.08)] bg-[#070A10] px-3 py-1.5 pl-8 font-mono text-xs text-white placeholder-slate-500 focus:border-[#22D3EE]/50 focus:outline-none"
				/>
				<svg
					class="absolute top-2.5 left-2.5 h-3.5 w-3.5 text-slate-500"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
					></path>
				</svg>
			</div>
		</div>

		<!-- Feed Sources Table -->
		<div
			class="overflow-hidden rounded-2xl border border-[rgba(255,255,255,0.08)] bg-[#0F172A]/20 shadow-xl"
		>
			<div class="overflow-x-auto">
				<table class="w-full text-left font-mono text-xs">
					<thead
						class="border-b border-[rgba(255,255,255,0.06)] bg-[#070A10]/60 text-[10px] text-slate-500 uppercase"
					>
						<tr>
							<th class="px-4 py-3 sm:px-6">Status</th>
							<th class="px-4 py-3">Source Name</th>
							<th class="px-4 py-3">Category</th>
							<th class="px-4 py-3">Feed Endpoint</th>
							<th class="px-4 py-3">Last Polled</th>
							<th class="px-4 py-3 text-right sm:px-6">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-[rgba(255,255,255,0.04)] text-slate-300">
						{#if filteredSources.length === 0}
							<tr>
								<td colspan="6" class="py-12 text-center text-slate-500">
									No RSS sources found matching the selected filter.
								</td>
							</tr>
						{:else}
							{#each filteredSources as src (src.id)}
								<tr class="transition-colors hover:bg-[#0F172A]/40">
									<!-- Status Toggle -->
									<td class="px-4 py-3.5 sm:px-6">
										<div class="flex items-center gap-2.5">
											<button
												type="button"
												role="switch"
												aria-checked={src.enabled}
												aria-label="Toggle {src.name} feed active state"
												onclick={() => handleToggle(src)}
												class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer items-center rounded-full border transition-colors {src.enabled
													? 'border-[#22D3EE]/50 bg-[#22D3EE]/20'
													: 'border-[rgba(255,255,255,0.15)] bg-[#070A10]'}"
											>
												<span
													class="inline-block h-3.5 w-3.5 transform rounded-full transition-transform duration-200 {src.enabled
														? 'translate-x-4 bg-[#22D3EE]'
														: 'translate-x-0.5 bg-slate-500'}"
												></span>
											</button>
											<span
												class="text-[10px] font-bold tracking-wider uppercase {src.enabled
													? 'text-[#22D3EE]'
													: 'text-slate-500'}"
											>
												{src.enabled ? 'ACTIVE' : 'OFF'}
											</span>
										</div>
									</td>

									<!-- Source Name -->
									<td class="px-4 py-3.5 font-sans font-medium text-white">
										{src.name}
									</td>

									<!-- Category Badge -->
									<td class="px-4 py-3.5">
										<span
											class="inline-block rounded border border-slate-700/60 bg-slate-800/40 px-2 py-0.5 text-[10px] text-slate-300 uppercase"
										>
											{src.category}
										</span>
									</td>

									<!-- URL Link -->
									<td class="max-w-[280px] truncate px-4 py-3.5">
										<a
											href={src.url}
											target="_blank"
											rel="noopener noreferrer"
											class="inline-flex items-center gap-1 text-slate-400 transition-colors hover:text-[#22D3EE]"
											title={src.url}
										>
											<span class="truncate">{src.url}</span>
											<svg
												class="h-3 w-3 flex-shrink-0 opacity-60"
												fill="none"
												viewBox="0 0 24 24"
												stroke="currentColor"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
												></path>
											</svg>
										</a>
									</td>

									<!-- Last Polled -->
									<td class="px-4 py-3.5 text-[11px] text-slate-400">
										{formatDate(src.last_fetched_at)}
									</td>

									<!-- Action Buttons -->
									<td class="px-4 py-3.5 text-right sm:px-6">
										<div class="flex items-center justify-end gap-2">
											<button
												onclick={() => openProbeForSource(src)}
												class="rounded border border-cyan-500/20 bg-cyan-500/5 px-2 py-1 text-[10px] text-cyan-400 transition-colors hover:border-cyan-500/50 hover:bg-cyan-500/15"
												title="Test & preview RSS XML response"
											>
												Probe
											</button>
											<button
												onclick={() => openEditModal(src)}
												class="rounded border border-slate-700 bg-slate-800/40 px-2 py-1 text-[10px] text-slate-300 transition-colors hover:border-slate-500 hover:text-white"
											>
												Edit
											</button>
											<button
												onclick={() => confirmDelete(src)}
												class="rounded border border-[#E11D48]/30 bg-[#E11D48]/5 px-2 py-1 text-[10px] text-[#E11D48] transition-colors hover:border-[#E11D48] hover:bg-[#E11D48]/15"
											>
												Delete
											</button>
										</div>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</section>

<!-- Add / Edit Modal -->
{#if showEditModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-sm"
		role="dialog"
		aria-modal="true"
	>
		<div
			class="relative max-h-[90vh] w-full max-w-xl overflow-y-auto rounded-2xl border border-[rgba(255,255,255,0.1)] bg-[#0A0E17] p-6 font-mono text-xs shadow-2xl"
		>
			<div
				class="mb-6 flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-4"
			>
				<div>
					<span class="text-[10px] font-bold tracking-wider text-[#22D3EE] uppercase">
						{isEditing ? 'Modify Source' : 'New Registration'}
					</span>
					<h2 class="font-serif text-xl font-medium text-white">
						{isEditing ? 'Edit RSS Feed' : 'Add RSS Source'}
					</h2>
				</div>
				<button
					onclick={() => (showEditModal = false)}
					class="cursor-pointer text-slate-500 hover:text-white"
				>
					[✕]
				</button>
			</div>

			<form
				onsubmit={(e) => {
					e.preventDefault();
					saveSource();
				}}
				class="space-y-4"
			>
				<!-- Feed URL -->
				<div>
					<label for="form-url" class="mb-1 block text-[10px] text-slate-400 uppercase">
						Feed Endpoint URL (RSS / Atom XML)
					</label>
					<div class="flex gap-2">
						<input
							id="form-url"
							type="url"
							bind:value={formURL}
							placeholder="https://example.com/feed.xml"
							required
							class="flex-grow rounded-lg border border-[rgba(255,255,255,0.1)] bg-[#070A10] px-3 py-2 text-xs text-white placeholder-slate-600 focus:border-[#22D3EE]/50 focus:outline-none"
						/>
						<button
							type="button"
							onclick={probeFeedURL}
							disabled={isProbing || !formURL}
							class="flex-shrink-0 cursor-pointer rounded-lg border border-[#22D3EE]/40 bg-[#22D3EE]/10 px-3 py-2 text-xs font-bold text-[#22D3EE] transition-all hover:bg-[#22D3EE]/20 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{#if isProbing}
								Probing...
							{:else}
								Test &amp; Preview
							{/if}
						</button>
					</div>
				</div>

				<!-- Probe Preview Box -->
				{#if isProbing}
					<div
						class="rounded-xl border border-slate-800 bg-[#070A10]/60 p-4 text-center text-slate-400"
					>
						<div
							class="mx-auto mb-2 h-4 w-4 animate-spin rounded-full border border-slate-700 border-t-[#22D3EE]"
						></div>
						Parsing feed XML headers and items...
					</div>
				{:else if probeResult}
					{#if probeResult.valid}
						<div class="space-y-2 rounded-xl border border-emerald-500/30 bg-emerald-500/5 p-4">
							<div class="flex items-center justify-between">
								<span class="font-bold text-emerald-400 uppercase">✓ Valid RSS Feed</span>
								<span class="rounded bg-emerald-500/20 px-2 py-0.5 text-[10px] text-emerald-300">
									{probeResult.item_count ?? 0} items found
								</span>
							</div>
							{#if probeResult.title}
								<div class="font-sans font-medium text-white">{probeResult.title}</div>
							{/if}
							{#if probeResult.items && probeResult.items.length > 0}
								<div class="space-y-1.5 border-t border-emerald-500/15 pt-2">
									<span class="text-[10px] text-slate-400 uppercase"
										>Latest Detected Headlines:</span
									>
									<ul class="list-inside list-disc space-y-1 text-[11px] text-slate-300">
										{#each probeResult.items as itm}
											<li class="truncate font-sans">{itm.title}</li>
										{/each}
									</ul>
								</div>
							{/if}
						</div>
					{:else}
						<div class="rounded-xl border border-[#E11D48]/30 bg-[#E11D48]/5 p-3 text-[#E11D48]">
							<span class="font-bold uppercase">✕ Probe Failed:</span>
							<p class="mt-1 text-[11px] text-slate-300">{probeResult.error}</p>
						</div>
					{/if}
				{/if}

				<!-- Feed Name -->
				<div>
					<label for="form-name" class="mb-1 block text-[10px] text-slate-400 uppercase">
						Display Name
					</label>
					<input
						id="form-name"
						type="text"
						bind:value={formName}
						placeholder="e.g. OpenAI Research Blog"
						required
						class="w-full rounded-lg border border-[rgba(255,255,255,0.1)] bg-[#070A10] px-3 py-2 text-xs text-white placeholder-slate-600 focus:border-[#22D3EE]/50 focus:outline-none"
					/>
				</div>

				<!-- Category & Status Row -->
				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
					<div>
						<label for="form-category" class="mb-1 block text-[10px] text-slate-400 uppercase">
							Default Category
						</label>
						<select
							id="form-category"
							bind:value={formCategory}
							class="w-full cursor-pointer rounded-lg border border-[rgba(255,255,255,0.1)] bg-[#070A10] px-3 py-2 text-xs text-white focus:border-[#22D3EE]/50 focus:outline-none"
						>
							{#each categories as cat}
								<option value={cat.slug}>{cat.name}</option>
							{/each}
						</select>
					</div>

					<div>
						<span class="mb-1 block text-[10px] text-slate-400 uppercase">Ingestion Status</span>
						<div
							class="flex h-[38px] items-center justify-between rounded-lg border border-[rgba(255,255,255,0.1)] bg-[#070A10] px-3"
						>
							<span class="text-xs text-slate-300">Active Polling</span>
							<button
								type="button"
								role="switch"
								aria-checked={formEnabled}
								aria-label="Toggle feed active state"
								onclick={() => (formEnabled = !formEnabled)}
								class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer items-center rounded-full border transition-colors {formEnabled
									? 'border-[#22D3EE]/50 bg-[#22D3EE]/20'
									: 'border-[rgba(255,255,255,0.15)] bg-[#070A10]'}"
							>
								<span
									class="inline-block h-3.5 w-3.5 transform rounded-full transition-transform duration-200 {formEnabled
										? 'translate-x-4 bg-[#22D3EE]'
										: 'translate-x-0.5 bg-slate-500'}"
								></span>
							</button>
						</div>
					</div>
				</div>

				<!-- Form Actions -->
				<div
					class="mt-6 flex items-center justify-end gap-3 border-t border-[rgba(255,255,255,0.06)] pt-4"
				>
					<button
						type="button"
						onclick={() => (showEditModal = false)}
						class="cursor-pointer rounded-lg border border-slate-700 bg-slate-800/40 px-4 py-2 text-xs text-slate-300 transition-colors hover:text-white"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={isSaving}
						class="cursor-pointer rounded-lg border border-[#22D3EE]/40 bg-[#22D3EE]/15 px-4 py-2 text-xs font-bold text-[#22D3EE] transition-all hover:bg-[#22D3EE]/25 hover:text-white disabled:opacity-50"
					>
						{isSaving ? 'Saving...' : isEditing ? 'Save Changes' : 'Create Source'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteModal && sourceToDelete}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-sm"
		role="dialog"
		aria-modal="true"
	>
		<div
			class="w-full max-w-md rounded-2xl border border-[#E11D48]/30 bg-[#0A0E17] p-6 font-mono text-xs shadow-2xl"
		>
			<div class="mb-4">
				<span class="text-[10px] font-bold text-[#E11D48] uppercase">CONFIRMATION REQUIRED</span>
				<h3 class="mt-1 font-serif text-lg text-white">Delete RSS Source</h3>
			</div>

			<p class="leading-relaxed text-slate-300">
				Are you sure you want to permanently remove <strong class="text-white"
					>"{sourceToDelete.name}"</strong
				>? The scheduler will cease fetching articles from this feed.
			</p>

			<div class="mt-6 flex items-center justify-end gap-3">
				<button
					type="button"
					onclick={() => {
						showDeleteModal = false;
						sourceToDelete = null;
					}}
					class="cursor-pointer rounded-lg border border-slate-700 bg-slate-800/40 px-4 py-2 text-slate-300 hover:text-white"
				>
					Cancel
				</button>
				<button
					type="button"
					onclick={executeDelete}
					disabled={isDeleting}
					class="cursor-pointer rounded-lg border border-[#E11D48]/40 bg-[#E11D48]/20 px-4 py-2 font-bold text-[#E11D48] transition-all hover:bg-[#E11D48]/30 hover:text-white disabled:opacity-50"
				>
					{isDeleting ? 'Deleting...' : 'Delete Feed'}
				</button>
			</div>
		</div>
	</div>
{/if}
