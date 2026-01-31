<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { fly } from 'svelte/transition';
	import ScrollProgress from './ScrollProgress.svelte';

	interface Props {
		title: string;
		date: string;
		description: string;
		backLink?: string;
		backText?: string;
		children?: import('svelte').Snippet;
	}

	let {
		title,
		date,
		description,
		backLink = '/claude',
		backText = 'Back to Claude',
		children
	}: Props = $props();

	let mounted = $state(false);

	onMount(async () => {
		mounted = true;
		await tick();

		// Add copy buttons to code blocks
		setTimeout(() => {
			document.querySelectorAll('.article-content pre').forEach((pre) => {
				// Skip if already wrapped (by ArticleLayout or global)
				const parent = pre.parentElement;
				if (parent?.classList.contains('code-block') || parent?.classList.contains('code-block-global')) return;

				const wrapper = document.createElement('div');
				wrapper.className = 'code-block';

				const button = document.createElement('button');
				button.className = 'copy-btn';
				button.textContent = 'Copy';
				button.addEventListener('click', async () => {
					const code = pre.querySelector('code')?.textContent || pre.textContent || '';
					await navigator.clipboard.writeText(code);
					button.textContent = 'Copied!';
					setTimeout(() => (button.textContent = 'Copy'), 2000);
				});

				pre.parentNode?.insertBefore(wrapper, pre);
				wrapper.appendChild(button);
				wrapper.appendChild(pre);
			});
		}, 100);
	});
</script>

<svelte:head>
	<title>{title} | Faan Rossouw</title>
	<meta name="description" content={description} />
</svelte:head>

<ScrollProgress />

<article class="article">
	<div class="container">
		{#if mounted}
			<header class="article-header" in:fly={{ y: 30, duration: 800, delay: 200 }}>
				<a href={backLink} class="back-link">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<line x1="19" y1="12" x2="5" y2="12"></line>
						<polyline points="12 19 5 12 12 5"></polyline>
					</svg>
					{backText}
				</a>
				<span class="date">{date}</span>
				<h1>{title}</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				{@render children?.()}
			</div>

			<!-- CTA Card -->
			<div class="cta-card" in:fly={{ y: 20, duration: 600, delay: 600 }}>
				<div class="cta-content">
					<span class="cta-label">Want to go deeper?</span>
					<h3>Learn to Build AI-Powered Security Tools</h3>
					<p>Discover how to leverage agentic AI in your security workflow. From threat hunting to automation — learn to build tools that amplify your expertise.</p>
					<a href="https://aionsec.ai/course" class="cta-button" target="_blank" rel="noopener">
						Explore the Course
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<line x1="5" y1="12" x2="19" y2="12"></line>
							<polyline points="12 5 19 12 12 19"></polyline>
						</svg>
					</a>
				</div>
			</div>
		{/if}
	</div>
</article>

<style>
	.article {
		padding: 60px 0 100px;
	}

	.container {
		max-width: 800px;
		margin: 0 auto;
		padding: 0 24px;
	}

	.article-header {
		margin-bottom: 48px;
	}

	.back-link {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		color: var(--aion-purple);
		text-decoration: none;
		margin-bottom: 24px;
		transition: opacity 0.2s;
	}

	.back-link:hover {
		opacity: 0.8;
	}

	.date {
		display: block;
		font-size: 14px;
		color: var(--aion-purple);
		margin-bottom: 16px;
	}

	h1 {
		font-size: clamp(28px, 5vw, 42px);
		font-weight: 700;
		line-height: 1.2;
		color: var(--white);
		margin: 0;
	}

	/* Content styles using :global() since content is passed as children */
	.article-content {
		font-size: 17px;
		line-height: 1.8;
		color: rgba(255, 255, 255, 0.85);
	}

	.article-content :global(p) {
		margin-bottom: 24px;
	}

	.article-content :global(h2) {
		font-size: 24px;
		font-weight: 600;
		color: var(--white);
		margin: 48px 0 24px;
	}

	.article-content :global(h3) {
		font-size: 20px;
		font-weight: 600;
		color: var(--white);
		margin: 32px 0 16px;
	}

	.article-content :global(hr) {
		border: none;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
		margin: 48px 0;
	}

	.article-content :global(ul),
	.article-content :global(ol) {
		margin-bottom: 24px;
		padding-left: 24px;
	}

	.article-content :global(li) {
		margin-bottom: 8px;
	}

	.article-content :global(strong) {
		color: var(--white);
	}

	.article-content :global(em) {
		font-style: italic;
	}

	.article-content :global(a) {
		color: var(--aion-purple);
		text-decoration: none;
		transition: opacity 0.2s;
	}

	.article-content :global(a:hover) {
		opacity: 0.8;
	}

	/* Inline code */
	.article-content :global(code) {
		background: rgba(189, 147, 249, 0.15);
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'SF Mono', 'Fira Code', monospace;
		font-size: 0.9em;
		color: var(--aion-purple-light);
	}

	/* Code blocks */
	.article-content :global(pre) {
		background: rgba(0, 0, 0, 0.4);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 8px;
		padding: 20px;
		overflow-x: auto;
		margin-bottom: 24px;
	}

	.article-content :global(pre code) {
		background: none;
		padding: 0;
		font-size: 14px;
		color: rgba(255, 255, 255, 0.9);
		line-height: 1.6;
	}

	/* Code block with copy button */
	.article-content :global(.code-block) {
		position: relative;
		margin-bottom: 24px;
	}

	.article-content :global(.code-block pre) {
		margin-bottom: 0;
		padding-top: 44px;
	}

	.article-content :global(.copy-btn) {
		position: absolute;
		top: 8px;
		right: 8px;
		padding: 4px 10px;
		font-size: 12px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.7);
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.2s;
	}

	.article-content :global(.copy-btn:hover) {
		color: var(--white);
		background: rgba(255, 255, 255, 0.15);
		border-color: rgba(255, 255, 255, 0.3);
	}

	/* Images */
	.article-content :global(figure) {
		margin: 32px 0;
	}

	.article-content :global(figure img),
	.article-content :global(.article-image img) {
		width: 100%;
		height: auto;
		border-radius: 8px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
	}

	.article-content :global(.article-image) {
		margin: 32px 0;
	}

	/* Tables */
	.article-content :global(table) {
		width: 100%;
		border-collapse: collapse;
		font-size: 15px;
		margin: 24px 0;
	}

	.article-content :global(th),
	.article-content :global(td) {
		padding: 12px 16px;
		text-align: left;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.article-content :global(th) {
		color: var(--white);
		font-weight: 600;
		background: rgba(189, 147, 249, 0.1);
	}

	.article-content :global(td) {
		color: rgba(255, 255, 255, 0.85);
	}

	.article-content :global(tr:hover td) {
		background: rgba(255, 255, 255, 0.03);
	}

	/* Table wrappers for overflow */
	.article-content :global(.comparison-table),
	.article-content :global(.data-table) {
		margin: 24px 0;
		overflow-x: auto;
	}

	/* Blockquotes */
	.article-content :global(blockquote) {
		border-left: 3px solid var(--aion-purple);
		margin: 24px 0;
		padding: 12px 20px;
		background: rgba(189, 147, 249, 0.1);
		border-radius: 0 8px 8px 0;
	}

	.article-content :global(blockquote p) {
		margin: 0;
	}

	/* CTA Card */
	.cta-card {
		margin-top: 64px;
		padding: 32px;
		background: linear-gradient(135deg, rgba(189, 147, 249, 0.1) 0%, rgba(189, 147, 249, 0.05) 100%);
		border: 1px solid rgba(189, 147, 249, 0.2);
		border-radius: 16px;
		text-align: center;
	}

	.cta-label {
		display: inline-block;
		font-size: 12px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--aion-purple);
		margin-bottom: 12px;
	}

	.cta-card h3 {
		font-size: 24px;
		font-weight: 700;
		color: var(--white);
		margin: 0 0 12px 0;
	}

	.cta-card p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.7);
		line-height: 1.6;
		margin: 0 0 24px 0;
		max-width: 500px;
		margin-left: auto;
		margin-right: auto;
	}

	.cta-button {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 12px 24px;
		background: var(--aion-yellow, #f5e663);
		color: var(--aion-grey-dark, #1a1a2e);
		font-size: 14px;
		font-weight: 600;
		text-decoration: none;
		border-radius: 8px;
		transition: all 0.2s ease;
	}

	.cta-button:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 20px rgba(245, 230, 99, 0.3);
	}

	.cta-button svg {
		transition: transform 0.2s ease;
	}

	.cta-button:hover svg {
		transform: translateX(4px);
	}

	/* Responsive */
	@media (max-width: 768px) {
		.article {
			padding: 40px 0 80px;
		}

		.article-content {
			font-size: 16px;
		}

		.article-content :global(pre) {
			padding: 16px;
		}

		.article-content :global(pre code) {
			font-size: 13px;
		}

		.article-content :global(.code-block pre) {
			padding-top: 44px;
		}

		.article-content :global(table) {
			font-size: 14px;
		}

		.article-content :global(th),
		.article-content :global(td) {
			padding: 10px 12px;
		}

		.cta-card {
			margin-top: 48px;
			padding: 24px;
		}

		.cta-card h3 {
			font-size: 20px;
		}

		.cta-card p {
			font-size: 14px;
		}
	}
</style>
