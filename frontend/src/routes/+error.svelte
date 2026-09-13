<script lang="ts">
	import { page } from '$app/stores';
</script>

<svelte:head>
	<title>{$page.status || 404} — Not Found | Neuralwire</title>
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

<section class="flex min-h-[65vh] flex-col items-center justify-center px-4 py-20 text-center">
	<div class="mx-auto max-w-lg space-y-6">
		<div
			class="inline-block rounded-full border border-[#22D3EE]/30 bg-[#22D3EE]/10 px-4 py-1.5 font-mono text-xs font-bold tracking-widest text-[#22D3EE] uppercase"
		>
			ERROR {$page.status || 404}
		</div>
		<h1 class="font-serif text-3xl font-light tracking-tight text-white sm:text-5xl">
			Page Not Found
		</h1>
		<p class="font-sans text-sm leading-relaxed text-slate-400">
			{$page.error?.message ||
				'The article or page you are looking for does not exist, has been removed, or is not yet published.'}
		</p>
		<div class="pt-4">
			<a
				href="/"
				class="inline-flex items-center space-x-2 rounded-xl border border-[#22D3EE]/40 bg-[#22D3EE]/10 px-6 py-3 font-mono text-xs font-bold tracking-wider text-[#22D3EE] uppercase transition-all hover:border-[#22D3EE] hover:bg-[#22D3EE]/20 hover:text-white"
			>
				<span>← RETURN TO HOMEPAGE</span>
			</a>
		</div>
	</div>
</section>
