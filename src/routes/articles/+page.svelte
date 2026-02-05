<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { articleSeries } from '$lib/data/articles';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});
</script>

<svelte:head>
	<title>Articles | Faan Rossouw</title>
	<meta name="description" content="Research articles on C2 techniques, threat hunting methodology, and malware analysis published on Active Countermeasures." />
</svelte:head>

<section class="articles-hero">
	<div class="container">
		{#if mounted}
			<h1 in:fly={{ y: 30, duration: 800, delay: 200 }}>Articles</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Research and analysis published on <a href="https://www.activecountermeasures.com" target="_blank" rel="noopener noreferrer">Active Countermeasures</a>
			</p>
		{/if}
	</div>
</section>

{#each articleSeries as series, seriesIndex}
	<section class="article-series">
		<div class="container">
			{#if mounted}
				<div class="series-header" in:fly={{ y: 30, duration: 600, delay: 500 + seriesIndex * 100 }}>
					<h2>{series.name}</h2>
					<p>{series.description}</p>
				</div>
				<div class="articles-grid">
					{#each series.articles as article, articleIndex}
						<a
							href={article.url}
							target="_blank"
							rel="noopener noreferrer"
							class="article-card glass-card"
							in:fly={{ y: 30, duration: 500, delay: 600 + seriesIndex * 100 + articleIndex * 50 }}
						>
							<span class="series-badge">{series.name}</span>
							<h3>{article.title}</h3>
							<span class="external-link">
								Read on Active Countermeasures
								<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
									<polyline points="15 3 21 3 21 9"></polyline>
									<line x1="10" y1="14" x2="21" y2="3"></line>
								</svg>
							</span>
						</a>
					{/each}
				</div>
			{/if}
		</div>
	</section>
{/each}

<section class="aionsec-cta">
	<div class="container">
		{#if mounted}
			<div class="cta-card glass-card" in:fly={{ y: 30, duration: 600, delay: 800 }}>
				<h2>From Research to Practice</h2>
				<p>These articles explore threats in the wild. AionSec teaches you how to build the agentic systems that detect them</p>
				<a href="https://aionsec.ai" target="_blank" rel="noopener noreferrer" class="btn-primary">
					Visit AionSec
				</a>
			</div>
		{/if}
	</div>
</section>

<style>
	.articles-hero {
		padding: 80px 0 60px;
		text-align: center;
	}

	.articles-hero h1 {
		margin-bottom: 16px;
	}

	.lead {
		font-size: clamp(16px, 2vw, 20px);
		color: rgba(255, 255, 255, 0.7);
	}

	.lead a {
		color: var(--aion-purple);
		text-decoration: none;
	}

	.lead a:hover {
		color: var(--aion-purple-light);
	}

	.article-series {
		padding: 40px 0 60px;
	}

	.series-header {
		margin-bottom: 32px;
	}

	.series-header h2 {
		font-size: clamp(24px, 3vw, 32px);
		margin-bottom: 12px;
		background: linear-gradient(135deg, var(--aion-purple), var(--aion-purple-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.series-header p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.6);
		max-width: 600px;
	}

	.articles-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 20px;
	}

	.article-card {
		display: flex;
		flex-direction: column;
		padding: 24px;
		text-decoration: none;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.article-card:hover {
		transform: translateY(-3px);
		border-color: rgba(189, 147, 249, 0.5);
	}

	.series-badge {
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--aion-purple);
		margin-bottom: 12px;
	}

	.article-card h3 {
		font-size: 16px;
		font-weight: 600;
		color: var(--white);
		line-height: 1.4;
		flex-grow: 1;
		margin-bottom: 16px;
	}

	.external-link {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		font-weight: 500;
		color: var(--aion-yellow);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.article-card:hover .external-link {
		color: var(--aion-yellow-light);
	}

	.external-link svg {
		flex-shrink: 0;
	}

	.aionsec-cta {
		padding: 20px 0 100px;
	}

	.cta-card {
		max-width: 500px;
		margin: 0 auto;
		text-align: center;
		padding: 40px;
		animation: ctaGlow 5s ease-in-out infinite;
	}

	@keyframes ctaGlow {
		0%, 100% {
			box-shadow: 0 0 15px rgba(245, 230, 99, 0.1), 0 0 30px rgba(245, 230, 99, 0.05);
		}
		50% {
			box-shadow: 0 0 20px rgba(245, 230, 99, 0.18), 0 0 40px rgba(245, 230, 99, 0.08);
		}
	}

	.cta-card h2 {
		font-size: 24px;
		margin-bottom: 12px;
	}

	.cta-card p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.7);
		margin-bottom: 24px;
	}

	@media (max-width: 768px) {
		.articles-grid {
			grid-template-columns: 1fr;
		}

		.article-card {
			padding: 20px;
		}

		.cta-card {
			padding: 32px 24px;
		}
	}
</style>
