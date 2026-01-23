<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import ScrollProgress from '$lib/components/ScrollProgress.svelte';

	let mounted = $state(false);

	onMount(async () => {
		mounted = true;
		await tick();

		// Small delay to ensure transitions complete
		setTimeout(() => {
			document.querySelectorAll('pre').forEach((pre) => {
				const wrapper = document.createElement('div');
				wrapper.className = 'code-block';

				const button = document.createElement('button');
				button.className = 'copy-btn';
				button.textContent = 'Copy';
				button.addEventListener('click', async () => {
					const code = pre.querySelector('code')?.textContent || pre.textContent || '';
					await navigator.clipboard.writeText(code);
					button.textContent = 'Copied!';
					setTimeout(() => button.textContent = 'Copy', 2000);
				});

				pre.parentNode?.insertBefore(wrapper, pre);
				wrapper.appendChild(button);
				wrapper.appendChild(pre);
			});
		}, 100);
	});
</script>

<svelte:head>
	<title>Quality Over Volume | Faan Rossouw</title>
	<meta name="description" content="The trap of mistaking volume for value when working with Claude Code. Why incremental validation beats autopilot generation." />
</svelte:head>

<ScrollProgress />

<article class="article">
	<div class="container">
		{#if mounted}
			<header class="article-header" in:fly={{ y: 30, duration: 800, delay: 200 }}>
				<a href="/claude" class="back-link">
					<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<line x1="19" y1="12" x2="5" y2="12"></line>
						<polyline points="12 19 5 12 12 5"></polyline>
					</svg>
					Back to Claude
				</a>
				<span class="date">2026-01-23</span>
				<h1>Quality Over Volume</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				<figure class="article-image hero-image">
					<img src="/images/claude/quality-over-volume/hero.png" alt="Scholar overwhelmed by avalanche of scrolls" />
				</figure>

				<p>I had the idea for a course. A big one. The outline came together quickly - modules, lessons, a logical flow from basics to advanced. I handed it to Claude and said: go.</p>

				<p>Within an hour, I had dozens of lessons drafted. Content everywhere. It felt productive.</p>

				<p>Then I started reviewing.</p>

				<p>The first module had a fundamental flaw in how it explained a core concept. Not wrong exactly, but not what I meant. I clarified, asked Claude to fix it and propagate the change across everything.</p>

				<p>It did. And I found another issue. And another.</p>

				<p>Each fix rippled through the material. Each ripple introduced new inconsistencies. The whole thing started losing shape - reworked and refactored so many times that the original vision got diluted. I had a lot of content. I didn't have a course.</p>

				<hr />

				<h2>The Temptation</h2>

				<p>The temptation is always there when you're caffeinated and inspired: just go. Produce. Generate. You can fix things along the way.</p>

				<p>And there's some truth to that. Sometimes you need momentum. Sometimes the only way to know what you're building is to build it.</p>

				<p>But with Claude, the ability to generate at scale makes this temptation dangerous. You can produce a thousand articles with almost no effort. A thousand articles will sit on your hard drive. But so what? What value do they have? Do they serve anyone? Or are they just bytes organized in a certain structure, making you feel like you accomplished something?</p>

				<p>I've learned this lesson multiple times now. The hard way, every time.</p>

				<hr />

				<h2>The Pattern That Works</h2>

				<p>What actually works: incremental validation.</p>

				<p>Start with the first module. The first few lessons. Review them carefully. Make sure there's no misunderstanding - about the style, the tone, the technical details, whatever matters for that project.</p>

				<p>If something's off, catch it early. Update your instructions. Update your global rules. Fix the foundation before you build on top of it.</p>

				<p>Then do the next few lessons. Review again. By now you should be seeing consistent quality. The chapters coming out should express exactly what you intended.</p>

				<p>Only then do you put it on autopilot.</p>

				<p>This takes longer upfront. But it takes far less time overall than generating everything first and discovering foundational problems when you're already buried in content.</p>

				<figure class="article-image">
					<img src="/images/claude/quality-over-volume/pattern.png" alt="Craftsman carefully examining work before scaling" />
				</figure>

				<hr />

				<h2>The Bigger Point</h2>

				<p>This isn't really about Claude. It's about output and the illusion of progress. Mistaking volume for value is a trap that exists in all knowledge work. We've always been susceptible to feeling productive just because something appeared on our screen.</p>

				<p>But Claude amplifies this. The sheer volume of output now possible makes the trap deeper. You can generate more in an afternoon than you could have written in a month. That capability is extraordinary - and dangerous if you're not careful about what you're actually optimizing for.</p>

				<p>The goal of working with Claude isn't just more output. It's more output <em>and</em> better output. Both. At the same time. That's the real unlock - improving volume and quality simultaneously, not trading one for the other.</p>

				<p>When I rush and generate everything at once, I optimize for volume at the cost of quality. The fix is simple: slow down early, validate the foundation, then scale with confidence. That's how you get both.</p>

				<p>Plan more. Review earlier. Trust the process less than you trust your judgment about what you're actually trying to create.</p>

				<p>The mountains of content will still be there when you're ready for them. But they'll be the right mountains.</p>

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

	.article-content {
		font-size: 17px;
		line-height: 1.8;
		color: rgba(255, 255, 255, 0.85);
	}

	.article-content p {
		margin-bottom: 24px;
	}

	.article-content h2 {
		font-size: 24px;
		font-weight: 600;
		color: var(--white);
		margin: 48px 0 24px;
	}

	.article-content hr {
		border: none;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
		margin: 48px 0;
	}

	.article-content ul {
		margin-bottom: 24px;
		padding-left: 24px;
	}

	.article-content li {
		margin-bottom: 8px;
	}

	.article-content strong {
		color: var(--white);
	}

	.article-content em {
		font-style: italic;
	}

	.article-content code {
		background: rgba(189, 147, 249, 0.15);
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'SF Mono', 'Fira Code', monospace;
		font-size: 0.9em;
		color: var(--aion-purple-light);
	}

	.article-content pre {
		background: rgba(0, 0, 0, 0.4);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 8px;
		padding: 20px;
		overflow-x: auto;
		margin-bottom: 24px;
	}

	.article-content pre code {
		background: none;
		padding: 0;
		font-size: 14px;
		color: rgba(255, 255, 255, 0.9);
		line-height: 1.6;
	}

	.article-image {
		margin: 32px 0;
	}

	.article-image img {
		width: 100%;
		height: auto;
		border-radius: 8px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
	}

	:global(.code-block) {
		position: relative;
		margin-bottom: 24px;
	}

	:global(.code-block pre) {
		margin-bottom: 0;
	}

	:global(.copy-btn) {
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

	:global(.copy-btn:hover) {
		color: var(--white);
		background: rgba(255, 255, 255, 0.15);
		border-color: rgba(255, 255, 255, 0.3);
	}

	@media (max-width: 768px) {
		.article {
			padding: 40px 0 80px;
		}

		.article-content {
			font-size: 16px;
		}

		.article-content pre {
			padding: 16px;
			font-size: 13px;
		}
	}
</style>
