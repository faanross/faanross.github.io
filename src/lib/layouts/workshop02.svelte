<script>
	import ScrollProgress from '$lib/components/ScrollProgress.svelte';
	export let title = '';
</script>

<ScrollProgress contentSelector=".lesson-content" />

<svelte:head>
	<title>{title} | Let's Build a Multi-Modal C2</title>
</svelte:head>

<article class="lesson">
	<div class="lesson-header">
		<a href="/courses/workshop02" class="back-link">
			<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<polyline points="15 18 9 12 15 6"></polyline>
			</svg>
			Back to Course
		</a>
		<h1>{title}</h1>
	</div>

	<div class="lesson-content">
		<slot />
	</div>
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

	@media (max-width: 768px) {
		.lesson {
			padding: 60px 20px;
		}
	}
</style>
