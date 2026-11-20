<script lang="ts">
	import { page } from '$app/state';
	import EyeField from '$lib/components/EyeField.svelte';
</script>

<svelte:head>
	<title>{page.status} · Faan Rossouw</title>
</svelte:head>

<EyeField />

<main class="lost">
	<div class="box" data-eye-avoid>
		<p class="code">{page.status}</p>
		{#if page.status === 404}
			<p>Nothing to see here. Everyone looked.</p>
		{:else}
			<p>Something went wrong.</p>
		{/if}
		<p class="links">
			<a href="/">faanross.com</a><span class="sep">{' / '}</span><a href="https://aionsec.ai">aionsec.ai</a>
		</p>
	</div>
</main>

<style>
	.lost {
		position: relative;
		z-index: 1;
		min-height: 100vh;
		min-height: 100dvh;
		display: flex;
		align-items: flex-end;
		padding: 7rem clamp(24px, 11vw, 220px) clamp(48px, 12vh, 128px);
		pointer-events: none;
	}

	.box {
		pointer-events: auto;
		color: var(--fg-dim);
	}

	.code {
		color: var(--purple);
		font-size: 13px;
		margin-bottom: 1.2rem;
	}

	.links {
		margin-top: 1.4rem;
	}

	.sep {
		color: var(--fg-faint);
	}
</style>
