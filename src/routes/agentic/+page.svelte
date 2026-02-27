<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});

	const featured = {
		title: "The Obvious Approach Was Wrong",
		description: "I expected time-decay scoring to be a straightforward improvement. Apply decay, watch numbers improve, ship it. Instead I was staring at a 35% degradation.",
		url: "/agentic/time-decay",
		date: "2026-02-27",
		image: "/images/claude/time-decay/IMG-001-HERO.png"
	};

	const articles = [
		{
			title: "Before You Upgrade Anything, Measure What You Have",
			description: "My memory system finds the correct answer less than a third of the time. Here are the 28-query benchmark results across five categories that prove it.",
			url: "/agentic/benchmarks",
			date: "2026-02-24"
		},
		{
			title: "The Article That Made Me Realize I Didn't Create a Memory System After All",
			description: "I built a memory system for Claude Code. Then I read an article that made me realize it was just a well-organized chat log. Here's the audit, the gap, and the upgrade path.",
			url: "/agentic/memory-upgrade",
			date: "2026-02-22"
		},
		{
			title: "Agentic Coding as a Learning Accelerator: The Mindset and Systems That Make It Work",
			description: "The critics say AI agents will atrophy your skills. They're half right. Here's the mindset and systems that turn agentic coding into a learning accelerator.",
			url: "/agentic/agentic-coding-learning-accelerator",
			date: "2026-02-19"
		},
		{
			title: "The Bug That Proved the Point",
			description: "My Zeek parser looked right. The tests said otherwise. How verification catches what specifications miss.",
			url: "/agentic/verification-catches-misinterpretations",
			date: "2026-02-04"
		},
		{
			title: "Speedrunning Jevon's Paradox In My Own Life",
			description: "I expected AI tools to mean working less. My time logs say otherwise. A reflection on efficiency, expanded ambition, and the trap of Jevons Paradox.",
			url: "/agentic/jevons-paradox",
			date: "2026-01-29"
		},
		{
			title: "I Gave My AI Access to Its Own Memory",
			description: "Building an MCP server so Claude can search its own conversation history. From manual queries to autonomous memory access.",
			url: "/agentic/mcp-server",
			date: "2026-01-25"
		},
		{
			title: "Quality Over Volume",
			description: "The trap of mistaking volume for value when working with Claude Code. Why incremental validation beats autopilot generation.",
			url: "/agentic/quality-over-volume",
			date: "2026-01-23"
		},
		{
			title: "I Built Search That Finds What I Mean, Not What I Type",
			description: "Adding semantic search to my Claude memory system with LanceDB. Find conversations by concept, not just keywords.",
			url: "/agentic/semantic-search",
			date: "2026-01-19"
		},
		{
			title: "Your CLAUDE.md Is Probably Too Big",
			description: "How I reduced my CLAUDE.md from 570 lines to 187 - same functionality, 67% less context usage.",
			url: "/agentic/claude-md-optimization",
			date: "2026-01-17"
		},
		{
			title: "I Set Up a Local LLM and Learned to Design Around Its Limitations",
			description: "Setting up Ollama on Mac Mini for local embeddings and SQL generation. Why a 3B model with validation beats 8B alone.",
			url: "/agentic/local-llm",
			date: "2026-01-16"
		},
		{
			title: "I Gave My Memory System Ranked Search",
			description: "Adding full-text search with BM25 ranking. From 47 equal results to scored relevance with boolean queries and phrase matching.",
			url: "/agentic/ranked-search",
			date: "2026-01-15"
		},
		{
			title: "I Turned Claude's Hidden Memory Into a Queryable Database",
			description: "Building the analytics foundation for Claude's conversation history. 500MB of JSONL compressed to 69MB in DuckDB, with millisecond queries.",
			url: "/agentic/duckdb-foundation",
			date: "2026-01-14"
		},
		{
			title: "I'm Building a Memory System for Claude Code",
			description: "An 8-phase system to turn conversation transcripts into queryable, searchable, actionable memory. From DuckDB analytics to semantic search to a self-aware agent.",
			url: "/agentic/memory-system",
			date: "2026-01-13"
		},
		{
			title: "I Discovered Claude Code Has a Hidden Memory",
			description: "Claude Code stores complete conversation transcripts locally. Here's how to unlock persistent memory across sessions with one addition to your CLAUDE.md.",
			url: "/agentic/hidden-memory",
			date: "2025-01-11"
		},
		{
			title: "I Taught Claude to Text Me When It Needs Help",
			description: "Setting up async notifications and bidirectional control for Claude Code via Telegram.",
			url: "/agentic/iphone-notifications",
			date: "2025-01-10"
		}
	];
</script>

<svelte:head>
	<title>Agentic Design | Faan Rossouw</title>
	<meta name="description" content="Documenting my journey optimizing how I work with AI." />
</svelte:head>

<section class="claude-hero">
	<div class="container">
		{#if mounted}
			<h1 in:fly={{ y: 30, duration: 800, delay: 200 }}>Agentic Design</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Documenting my journey optimizing how I work with AI
			</p>
			<div class="hero-image" in:fade={{ duration: 600, delay: 500 }}>
				<img src="/images/claude/hero.png" alt="Claude Code and Obsidian integration" />
			</div>
		{/if}
	</div>
</section>

<section class="featured-section">
	<div class="container">
		{#if mounted}
			<a href={featured.url} class="featured-card glass-card" in:fly={{ y: 40, duration: 600, delay: 500 }}>
				<div class="featured-badge">
					<span class="badge">Latest</span>
				</div>
				<div class="featured-content">
					<span class="featured-date">{featured.date}</span>
					<h2 class="featured-title">{featured.title}</h2>
					<p class="featured-description">{featured.description}</p>
					<span class="featured-link">
						Read article
						<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<line x1="5" y1="12" x2="19" y2="12"></line>
							<polyline points="12 5 19 12 12 19"></polyline>
						</svg>
					</span>
				</div>
				<div class="featured-visual">
					<img src={featured.image} alt={featured.title} />
				</div>
			</a>
		{/if}
	</div>
</section>

<section class="claude-articles">
	<div class="container">
		{#if mounted}
			<div class="articles-grid">
				{#each articles as article, index}
					<a
						href={article.url}
						class="article-card glass-card"
						in:fly={{ y: 30, duration: 500, delay: 600 + index * 50 }}
					>
						<span class="date">{article.date}</span>
						<h3>{article.title}</h3>
						<p>{article.description}</p>
						<span class="read-more">
							Read article
							<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<line x1="5" y1="12" x2="19" y2="12"></line>
								<polyline points="12 5 19 12 12 19"></polyline>
							</svg>
						</span>
					</a>
				{/each}
			</div>
		{/if}
	</div>
</section>

<style>
	.claude-hero {
		padding: 80px 0 40px;
		text-align: center;
	}

	.claude-hero h1 {
		margin-bottom: 16px;
		background: linear-gradient(135deg, var(--aion-purple), var(--aion-purple-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.lead {
		font-size: clamp(16px, 2vw, 20px);
		color: rgba(255, 255, 255, 0.7);
	}

	.hero-image {
		margin-top: 32px;
		max-width: 625px;
		margin-left: auto;
		margin-right: auto;
	}

	.hero-image img {
		width: 100%;
		height: auto;
		border-radius: 8px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
	}

	/* Featured Section */
	.featured-section {
		padding: 40px 0;
	}

	.featured-section .container {
		max-width: 900px;
		margin: 0 auto;
	}

	.featured-card {
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: 32px;
		padding: 36px;
		text-decoration: none;
		position: relative;
		overflow: hidden;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.featured-card:hover {
		transform: translateY(-4px);
		border-color: rgba(189, 147, 249, 0.5);
	}

	.featured-badge {
		position: absolute;
		top: 20px;
		right: 20px;
	}

	.badge {
		display: inline-block;
		padding: 4px 12px;
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		border-radius: 12px;
		background: rgba(245, 230, 99, 0.15);
		color: var(--aion-yellow);
		border: 1px solid rgba(245, 230, 99, 0.25);
	}

	.featured-content {
		display: flex;
		flex-direction: column;
		justify-content: center;
	}

	.featured-date {
		font-size: 12px;
		font-weight: 500;
		color: var(--aion-purple);
		margin-bottom: 12px;
	}

	.featured-title {
		font-size: clamp(22px, 2.5vw, 28px);
		font-weight: 700;
		line-height: 1.3;
		margin-bottom: 14px;
		background: linear-gradient(135deg, #ffffff 0%, rgba(255, 255, 255, 0.8) 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.featured-description {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.7);
		line-height: 1.65;
		margin-bottom: 20px;
	}

	.featured-link {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: var(--aion-yellow);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		transition: all 0.3s ease;
	}

	.featured-card:hover .featured-link {
		color: var(--aion-yellow-light);
	}

	.featured-visual {
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 10px;
		overflow: hidden;
		border: 1px solid rgba(189, 147, 249, 0.15);
	}

	.featured-visual img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.claude-articles {
		padding: 0 0 80px;
	}

	.articles-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 24px;
		max-width: 900px;
		margin: 0 auto;
	}

	.article-card {
		display: flex;
		flex-direction: column;
		padding: 28px;
		text-decoration: none;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.article-card:hover {
		transform: translateY(-3px);
		border-color: rgba(189, 147, 249, 0.5);
	}

	.date {
		font-size: 12px;
		font-weight: 500;
		color: var(--aion-purple);
		margin-bottom: 12px;
	}

	.article-card h3 {
		font-size: 18px;
		font-weight: 600;
		color: var(--white);
		line-height: 1.4;
		margin-bottom: 12px;
	}

	.article-card p {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.6);
		line-height: 1.5;
		flex-grow: 1;
		margin-bottom: 16px;
	}

	.read-more {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		font-weight: 500;
		color: var(--aion-yellow);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.article-card:hover .read-more {
		color: var(--aion-yellow-light);
	}

	@media (max-width: 900px) {
		.featured-card {
			grid-template-columns: 1fr;
			padding: 28px;
		}

		.featured-visual {
			order: -1;
			max-height: 200px;
		}

		.featured-badge {
			position: static;
			margin-bottom: 12px;
		}
	}

	@media (max-width: 768px) {
		.articles-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
