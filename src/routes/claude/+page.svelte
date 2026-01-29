<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});

	const articles = [
		{
			title: "Speedrunning Jevon's Paradox In My Own Life",
			description: "I expected AI tools to mean working less. My time logs say otherwise. A reflection on efficiency, expanded ambition, and the trap of Jevons Paradox.",
			url: "/claude/jevons-paradox",
			date: "2026-01-29"
		},
		{
			title: "I Gave My AI Access to Its Own Memory",
			description: "Building an MCP server so Claude can search its own conversation history. From manual queries to autonomous memory access.",
			url: "/claude/mcp-server",
			date: "2026-01-25"
		},
		{
			title: "Quality Over Volume",
			description: "The trap of mistaking volume for value when working with Claude Code. Why incremental validation beats autopilot generation.",
			url: "/claude/quality-over-volume",
			date: "2026-01-23"
		},
		{
			title: "I Built Search That Finds What I Mean, Not What I Type",
			description: "Adding semantic search to my Claude memory system with LanceDB. Find conversations by concept, not just keywords.",
			url: "/claude/semantic-search",
			date: "2026-01-19"
		},
		{
			title: "Your CLAUDE.md Is Probably Too Big",
			description: "How I reduced my CLAUDE.md from 570 lines to 187 - same functionality, 67% less context usage.",
			url: "/claude/claude-md-optimization",
			date: "2026-01-17"
		},
		{
			title: "I Set Up a Local LLM and Learned to Design Around Its Limitations",
			description: "Setting up Ollama on Mac Mini for local embeddings and SQL generation. Why a 3B model with validation beats 8B alone.",
			url: "/claude/local-llm",
			date: "2026-01-16"
		},
		{
			title: "I Gave My Memory System Ranked Search",
			description: "Adding full-text search with BM25 ranking. From 47 equal results to scored relevance with boolean queries and phrase matching.",
			url: "/claude/ranked-search",
			date: "2026-01-15"
		},
		{
			title: "I Turned Claude's Hidden Memory Into a Queryable Database",
			description: "Building the analytics foundation for Claude's conversation history. 500MB of JSONL compressed to 69MB in DuckDB, with millisecond queries.",
			url: "/claude/duckdb-foundation",
			date: "2026-01-14"
		},
		{
			title: "I'm Building a Memory System for Claude Code",
			description: "An 8-phase system to turn conversation transcripts into queryable, searchable, actionable memory. From DuckDB analytics to semantic search to a self-aware agent.",
			url: "/claude/memory-system",
			date: "2026-01-13"
		},
		{
			title: "I Discovered Claude Code Has a Hidden Memory",
			description: "Claude Code stores complete conversation transcripts locally. Here's how to unlock persistent memory across sessions with one addition to your CLAUDE.md.",
			url: "/claude/hidden-memory",
			date: "2025-01-11"
		},
		{
			title: "I Taught Claude to Text Me When It Needs Help",
			description: "Setting up async notifications and bidirectional control for Claude Code via Telegram.",
			url: "/claude/iphone-notifications",
			date: "2025-01-10"
		}
	];
</script>

<svelte:head>
	<title>Claude | Faan Rossouw</title>
	<meta name="description" content="Chronicling my journey with Claude Code and Obsidian - optimizations, workflows, and lessons learned." />
</svelte:head>

<section class="claude-hero">
	<div class="container">
		{#if mounted}
			<h1 in:fly={{ y: 30, duration: 800, delay: 200 }}>Claude Code + Obsidian</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Documenting my journey optimizing how I work with AI
			</p>
			<div class="hero-image" in:fade={{ duration: 600, delay: 500 }}>
				<img src="/images/claude/hero.png" alt="Claude Code and Obsidian integration" />
			</div>
		{/if}
	</div>
</section>

<section class="claude-intro">
	<div class="container">
		{#if mounted}
			<div class="intro-content" in:fly={{ y: 30, duration: 600, delay: 500 }}>
				<p class="intro-text">
					Integrating Claude Code with Obsidian into my daily workflow changed everything. Not hype - genuine transformation.
				</p>
				<blockquote class="mantra">
					"Reduce friction to the minimum required to fully manifest an idea. Every unnecessary step between thought and creation is a leak in the system."
				</blockquote>
				<p class="intro-text">
					That became my mantra. And this section is where I document the journey - every optimization, every workflow tweak, every debugging session that led somewhere useful.
				</p>
			</div>
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

	.claude-intro {
		padding: 20px 0 60px;
	}

	.intro-content {
		max-width: 700px;
		margin: 0 auto;
	}

	.intro-text {
		font-size: 16px;
		line-height: 1.7;
		color: rgba(255, 255, 255, 0.8);
		margin-bottom: 24px;
	}

	.mantra {
		margin: 32px 0;
		padding: 24px 32px;
		border-left: 3px solid var(--aion-purple);
		background: rgba(189, 147, 249, 0.05);
		border-radius: 0 8px 8px 0;
	}

	.mantra {
		font-size: 18px;
		font-style: italic;
		color: rgba(255, 255, 255, 0.9);
		line-height: 1.6;
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

	@media (max-width: 768px) {
		.articles-grid {
			grid-template-columns: 1fr;
		}

		.mantra {
			padding: 20px 24px;
		}
	}
</style>
