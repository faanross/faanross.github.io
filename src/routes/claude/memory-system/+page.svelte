<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(async () => {
		mounted = true;
		await tick();

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
	<title>Building a Memory System for Claude Code | Faan Rossouw</title>
	<meta name="description" content="An 8-phase project to make Claude Code conversation history searchable, queryable, and useful for optimization." />
</svelte:head>

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
				<span class="date">2026-01-13</span>
				<h1>I'm Building a Memory System for My AI Pair Programmer</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				<p>Claude Code stores every conversation locally. Full transcripts. Tool calls. Timestamps. Everything. I <a href="/claude/hidden-memory">recently discovered</a> this growing treasure trove sitting in <code>~/.claude/projects/</code>.</p>

				<p>This is gold. But only if I can actually use it.</p>

				<p>Right now? I can ask Claude Code to grep through the files, but it's slow, limited, and doesn't scale. No way to run analytics, find patterns, or give Claude direct access to query its own history.</p>

				<p>That's about to change.</p>

				<figure class="article-image">
					<img src="/images/IMG-001-HERO.png" alt="AI Memory System visualization" />
				</figure>

				<hr />

				<h2>The Opportunity</h2>

				<p>This isn't just about asking "what did we discuss last week?" - though that would be useful.</p>

				<p>It's about <em>optimization</em>. Mining patterns from my own usage to make the workflow faster and smoother.</p>

				<p>One example: Claude Code stops and asks permission for certain commands. Every stop is friction. But which commands am I approving over and over? Which ones are clearly non-destructive and could be pre-approved? I can't answer that without querying my history.</p>

				<p>That's one insight among many. Tool usage patterns. Project time allocation. Recurring problems that might need permanent solutions. Context that keeps getting re-explained.</p>

				<p>The data is there. I just need to make it queryable - for myself, and for Claude directly.</p>

				<hr />

				<h2>What I'm Building</h2>

				<p>A complete memory system. Not just search - a full stack for understanding and retrieving my conversation history.</p>

				<p>Here's the architecture:</p>

				<figure class="article-image">
					<img src="/images/IMG-002-ARCHITECTURE.png" alt="8-phase memory system architecture" />
				</figure>

				<p>Eight phases. Each one adds a capability layer. Each one is independently useful.</p>

				<hr />

				<h2>The Phases</h2>

				<h3>Phase 1: Analytics Foundation (DuckDB)</h3>

				<figure class="article-image">
					<img src="/images/IMG-003-PHASE1-DUCKDB.png" alt="DuckDB analytics foundation" />
				</figure>

				<p>Parse the JSONL files into a proper database. Enable SQL queries.</p>

				<p><strong>What it unlocks:</strong></p>
				<ul>
					<li>"What projects have I worked on most?"</li>
					<li>"When am I most productive?"</li>
					<li>"Which tools do I use most?"</li>
				</ul>

				<p>DuckDB is embedded, file-based, ridiculously fast. No server to run. Just a single file on disk.</p>

				<p>This is the foundation everything else builds on.</p>

				<hr />

				<h3>Phase 2: Keyword Search (FTS5)</h3>

				<figure class="article-image">
					<img src="/images/IMG-004-PHASE2-FTS5.png" alt="Full-text search with ranking" />
				</figure>

				<p>Add full-text search with BM25 ranking.</p>

				<p><strong>What it unlocks:</strong></p>
				<ul>
					<li>Boolean queries: <code>authentication AND NOT oauth</code></li>
					<li>Phrase matching: <code>"rate limiting"</code></li>
					<li>Ranked results instead of "here's 500 matches, good luck"</li>
				</ul>

				<p>DuckDB has a built-in FTS extension. Porter stemming means "running" matches "run", "runs", "runner". Case-insensitive by default.</p>

				<p>The limitation: you need to know the exact keywords. "auth" won't find "login".</p>

				<hr />

				<h3>Phase 3: Local LLM Setup (Mac Mini + Llama)</h3>

				<figure class="article-image">
					<img src="/images/IMG-005-PHASE3-LOCALLLM.png" alt="Local LLM setup" />
				</figure>

				<p>Set up Ollama on my Mac Mini for two things:</p>
				<ol>
					<li><strong>Embeddings</strong> - Convert text to vectors for semantic search</li>
					<li><strong>NL→SQL</strong> - Generate database queries from natural language</li>
				</ol>

				<p><strong>Why local:</strong></p>
				<ul>
					<li>Gigabytes of conversation history = significant embedding cost if using cloud APIs</li>
					<li>Everything stays on my network</li>
					<li>Zero ongoing cost</li>
				</ul>

				<p>The Mac Mini runs the models. My main Mac calls the API remotely. Inference stays off my workstation.</p>

				<hr />

				<h3>Phase 4: Semantic Search (LanceDB)</h3>

				<figure class="article-image">
					<img src="/images/IMG-006-PHASE4-LANCEDB.png" alt="Semantic vector search" />
				</figure>

				<p>This is where it gets interesting.</p>

				<p>Embed every message as a vector. Store in LanceDB. Now search by <em>meaning</em>, not keywords.</p>

				<p><strong>What it unlocks:</strong></p>
				<ul>
					<li>"Find conversations about handling failures gracefully" → finds "retry logic", "circuit breaker", "exponential backoff"</li>
					<li>"What have I learned about concurrency?" → surfaces relevant discussions even if I never used that word</li>
					<li>Cross-project pattern matching</li>
				</ul>

				<p>Think of it as having a librarian who's read everything. Describe what you're looking for. They find it even if you used different words.</p>

				<hr />

				<h3>Phase 5: Agent Access (MCP Server)</h3>

				<figure class="article-image">
					<img src="/images/IMG-007-PHASE5-MCP.png" alt="MCP server for agent access" />
				</figure>

				<p>Build an MCP server so Claude can query its own history.</p>

				<p><strong>What it unlocks:</strong></p>
				<ul>
					<li>Claude checks for relevant past conversations automatically</li>
					<li>"Let me see if we've discussed this before..." → actual search happens</li>
					<li>Self-aware agent that can reference prior context</li>
				</ul>

				<p>The dashboard (Phase 6) is for me to explore visually. The MCP server is for Claude to access programmatically.</p>

				<p>This is where the system starts feeling less like a tool and more like augmented memory.</p>

				<hr />

				<h3>Phase 6: Visual Dashboard (SvelteKit + Layercake)</h3>

				<figure class="article-image">
					<img src="/images/IMG-008-PHASE6-DASHBOARD.png" alt="Visual dashboard" />
				</figure>

				<p>A proper UI for exploring the data.</p>

				<p><strong>Planned visualizations:</strong></p>
				<ul>
					<li>Activity heatmap (when am I most productive?)</li>
					<li>Project breakdown (where does my time go?)</li>
					<li>Topic clusters (what themes emerge?)</li>
					<li>Trend lines (how has my usage evolved?)</li>
				</ul>

				<p>DuckDB has a WASM build that runs in the browser. No backend needed for basic queries. Point the dashboard at a Parquet export and it just works.</p>

				<hr />

				<h3>Phase 7: Voice Control</h3>

				<figure class="article-image">
					<img src="/images/IMG-009-PHASE7-VOICE.png" alt="Voice control interface" />
				</figure>

				<p>Ask questions out loud. Get visual answers.</p>

				<p>"What did I work on last Friday?" → Bar chart of projects with session counts.</p>

				<p>Web Speech API for voice input. Pattern matching for common queries. Llama fallback for anything complex.</p>

				<p>The dream: verbal queries while I'm thinking through a problem, answers appearing on screen without touching the keyboard.</p>

				<hr />

				<h3>Phase 8: Go Backend</h3>

				<figure class="article-image">
					<img src="/images/IMG-010-PHASE8-GOBACKEND.png" alt="Go backend infrastructure" />
				</figure>

				<p>At my current usage rate, I'm generating over a gigabyte of conversation history per month.</p>

				<p>Browser-only won't scale. Phase 8 adds a proper Go backend:</p>
				<ul>
					<li>Serves both the dashboard and the MCP server</li>
					<li>Background ingestion watches for new sessions</li>
					<li>Handles queries that would choke a browser</li>
				</ul>

				<hr />

				<h2>Why This Order</h2>

				<p>Each phase is independently useful. You could stop at any point and have a working system.</p>

				<pre><code>Phase 1 alone: "I can query my history with SQL"
Phase 1+2: "I can search with keywords"
Phase 1+2+4: "I can search by meaning"
Phase 1+2+4+5: "Claude can search for me"</code></pre>

				<p>But they're ordered for dependencies too:</p>
				<ul>
					<li>Semantic search (Phase 4) needs embeddings (Phase 3)</li>
					<li>MCP server (Phase 5) needs all search types available</li>
					<li>Dashboard (Phase 6) needs data to visualize</li>
					<li>Voice control (Phase 7) needs both the dashboard and Llama</li>
				</ul>

				<p>Phase 8 could technically come earlier - it's about scale, not features. But I'd rather prove the concept with simpler architecture first.</p>

				<hr />

				<h2>What I Expect to Learn</h2>

				<p>This is a build log, not a tutorial. I'm writing it as I go.</p>

				<p>Some things will work as planned. Others won't. I'll hit walls I can't predict right now. Make tradeoffs that seem obvious in hindsight.</p>

				<p>A few things I'm already uncertain about:</p>

				<p><strong>Embedding strategy</strong> - Do I embed full messages? Chunk long ones? Summarize first? I'll figure it out when I have real vectors to query.</p>

				<p><strong>MCP tool design</strong> - What granularity makes sense? One <code>search</code> tool with modes, or separate <code>keyword_search</code> and <code>semantic_search</code>? Depends how Claude actually uses them.</p>

				<p><strong>Update frequency</strong> - Hourly cron job? Real-time file watcher? Depends on how often I need fresh data vs how much I care about resource usage.</p>

				<p>The plan is a starting point. Reality will edit it.</p>

				<hr />

				<h2>The Bigger Picture</h2>

				<p>My thesis: Claude Code isn't a tool I use. It's becoming how I think.</p>

				<p>Every deep work session leaves traces. Solutions discovered. Approaches rejected. Patterns that worked. Context accumulated over hundreds of conversations.</p>

				<p>Right now, all of that evaporates between sessions.</p>

				<p>Building this memory system is about preserving it. Surfacing it when relevant. Making the collaboration cumulative instead of amnesiac.</p>

				<p>If this resonates, follow along. I'll publish each phase as I complete it - the decisions, the debugging, the actual code.</p>

				<p>This is article 0. The foundation. The "what" and "why" before the "how".</p>

				<p>Let's build.</p>

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

	.article-content h3 {
		font-size: 20px;
		font-weight: 600;
		color: var(--white);
		margin: 36px 0 16px;
	}

	.article-content hr {
		border: none;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
		margin: 48px 0;
	}

	.article-content ul,
	.article-content ol {
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
