<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';

	interface Props {
		title: string;
		workshop: string;
		workshopTitle: string;
		prev?: { slug: string; title: string } | null;
		next?: { slug: string; title: string } | null;
		children?: import('svelte').Snippet;
	}

	let { title, workshop, workshopTitle, prev = null, next = null, children }: Props = $props();

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});
</script>

<svelte:head>
	<title>{title} | {workshopTitle}</title>
</svelte:head>

<article class="lesson">
	{#if mounted}
		<div class="lesson-header" in:fly={{ y: 20, duration: 600 }}>
			<a href="/courses/{workshop}" class="back-link">
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<polyline points="15 18 9 12 15 6"></polyline>
				</svg>
				Back to {workshopTitle}
			</a>
			<h1>{title}</h1>
		</div>

		<div class="lesson-content" in:fly={{ y: 30, duration: 600, delay: 200 }}>
			{@render children?.()}
		</div>

		<nav class="lesson-nav" in:fly={{ y: 20, duration: 600, delay: 400 }}>
			{#if prev}
				<a href="/courses/{workshop}/{prev.slug}" class="nav-prev">
					<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<polyline points="15 18 9 12 15 6"></polyline>
					</svg>
					<span>
						<small>Previous</small>
						{prev.title}
					</span>
				</a>
			{:else}
				<div></div>
			{/if}

			{#if next}
				<a href="/courses/{workshop}/{next.slug}" class="nav-next">
					<span>
						<small>Next</small>
						{next.title}
					</span>
					<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<polyline points="9 18 15 12 9 6"></polyline>
					</svg>
				</a>
			{:else}
				<div></div>
			{/if}
		</nav>
	{/if}
</article>

<style>
	.lesson {
		max-width: 900px;
		margin: 0 auto;
		padding: 80px 24px;
	}

	.lesson-header {
		margin-bottom: 40px;
	}

	.back-link {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: rgba(255, 255, 255, 0.6);
		text-decoration: none;
		margin-bottom: 20px;
		transition: color 0.3s ease;
	}

	.back-link:hover {
		color: var(--aion-purple);
	}

	h1 {
		font-size: clamp(28px, 4vw, 38px);
		font-weight: 700;
		line-height: 1.3;
	}

	.lesson-content {
		font-size: 16px;
		line-height: 1.8;
		color: rgba(255, 255, 255, 0.85);
	}

	.lesson-content :global(h2) {
		font-size: 24px;
		font-weight: 600;
		margin-top: 48px;
		margin-bottom: 16px;
		color: var(--aion-purple);
	}

	.lesson-content :global(h3) {
		font-size: 20px;
		font-weight: 600;
		margin-top: 36px;
		margin-bottom: 12px;
		color: var(--white);
	}

	.lesson-content :global(h4) {
		font-size: 17px;
		font-weight: 600;
		margin-top: 28px;
		margin-bottom: 10px;
		color: rgba(255, 255, 255, 0.9);
	}

	.lesson-content :global(p) {
		margin-bottom: 20px;
	}

	.lesson-content :global(ul),
	.lesson-content :global(ol) {
		margin-bottom: 20px;
		padding-left: 24px;
	}

	.lesson-content :global(li) {
		margin-bottom: 8px;
	}

	.lesson-content :global(a) {
		color: var(--aion-purple);
		text-decoration: none;
		border-bottom: 1px solid transparent;
		transition: border-color 0.3s ease;
	}

	.lesson-content :global(a:hover) {
		border-bottom-color: var(--aion-purple);
	}

	.lesson-content :global(code) {
		font-family: 'JetBrains Mono', monospace;
		font-size: 14px;
		background: rgba(61, 61, 71, 0.6);
		padding: 2px 6px;
		border-radius: 4px;
	}

	.lesson-content :global(pre) {
		background: rgba(20, 20, 25, 0.8);
		border: 1px solid rgba(189, 147, 249, 0.2);
		border-radius: 8px;
		padding: 20px;
		overflow-x: auto;
		margin: 24px 0;
	}

	.lesson-content :global(pre code) {
		background: none;
		padding: 0;
		font-size: 13px;
		line-height: 1.6;
	}

	.lesson-content :global(blockquote) {
		border-left: 3px solid var(--aion-purple);
		margin: 24px 0;
		padding: 12px 20px;
		background: rgba(189, 147, 249, 0.1);
		border-radius: 0 8px 8px 0;
	}

	.lesson-content :global(blockquote p) {
		margin: 0;
	}

	.lesson-content :global(img) {
		max-width: 100%;
		height: auto;
		border-radius: 8px;
		margin: 24px 0;
	}

	.lesson-content :global(table) {
		width: 100%;
		border-collapse: collapse;
		margin: 24px 0;
	}

	.lesson-content :global(th),
	.lesson-content :global(td) {
		padding: 12px;
		border: 1px solid rgba(189, 147, 249, 0.2);
		text-align: left;
	}

	.lesson-content :global(th) {
		background: rgba(189, 147, 249, 0.1);
		font-weight: 600;
	}

	.lesson-content :global(hr) {
		border: none;
		border-top: 1px solid rgba(189, 147, 249, 0.2);
		margin: 40px 0;
	}

	.lesson-nav {
		display: flex;
		justify-content: space-between;
		gap: 20px;
		margin-top: 60px;
		padding-top: 40px;
		border-top: 1px solid rgba(189, 147, 249, 0.2);
	}

	.nav-prev,
	.nav-next {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 20px;
		background: rgba(61, 61, 71, 0.4);
		border: 1px solid rgba(189, 147, 249, 0.2);
		border-radius: 12px;
		text-decoration: none;
		transition: all 0.3s ease;
		max-width: 45%;
	}

	.nav-prev:hover,
	.nav-next:hover {
		background: rgba(61, 61, 71, 0.6);
		border-color: rgba(189, 147, 249, 0.4);
		transform: translateY(-2px);
	}

	.nav-prev span,
	.nav-next span {
		display: flex;
		flex-direction: column;
	}

	.nav-prev small,
	.nav-next small {
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: rgba(255, 255, 255, 0.5);
		margin-bottom: 4px;
	}

	.nav-prev span,
	.nav-next span {
		font-size: 14px;
		color: var(--white);
	}

	.nav-next {
		text-align: right;
		margin-left: auto;
	}

	.nav-next span {
		align-items: flex-end;
	}

	@media (max-width: 768px) {
		.lesson {
			padding: 60px 20px;
		}

		.lesson-nav {
			flex-direction: column;
		}

		.nav-prev,
		.nav-next {
			max-width: 100%;
		}

		.nav-next {
			flex-direction: row-reverse;
			text-align: left;
		}

		.nav-next span {
			align-items: flex-start;
		}
	}
</style>
