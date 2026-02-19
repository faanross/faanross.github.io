<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="I'm Building a Memory System for Claude Code"
	date="2026-01-13"
	description="An 8-phase project to make Claude Code conversation history searchable, queryable, and useful for optimization."
>
	<p>Claude Code stores every conversation locally. Full transcripts. Tool calls. Timestamps. Everything. I <a href="/claude/hidden-memory">recently discovered</a> this growing treasure trove sitting in <code>~/.claude/projects/</code>.</p>

	<p>This is gold. But only if I can actually use it.</p>

	<p>Right now? I can ask Claude Code to grep through the files, but it's slow, limited, and doesn't scale. No way to run analytics, find patterns, or give Claude direct access to query its own history.</p>

	<p>That's about to change.</p>

	<figure class="article-image">
		<img src="/images/claude/memory-system/hero.png" alt="AI Memory System visualization" />
	</figure>

	<hr />

	<h2>The Opportunity</h2>

	<p>This isn't just about asking "what did we discuss last week?" - though that would be useful.</p>

	<p>It's about <em>optimization</em>. Mining patterns from my own usage to make the workflow faster and smoother.</p>

	<p>One example: Claude Code stops and asks permission for certain commands. Every stop is friction. But which commands am I approving over and over? Which ones are clearly non-destructive and could be pre-approved? I can't answer that without querying my history.</p>

	<p>That's one insight among many. Tool usage patterns. Project time allocation. Recurring problems that might need permanent solutions. Context that keeps getting re-explained.</p>

	<p>The data is there. I just need to make it queryable - for myself, and for Claude directly.</p>

	<hr />

	<h2>Why Not Just Use Grep?</h2>

	<p>Fair question. I could add instructions to my CLAUDE.md telling Claude to grep through <code>~/.claude/projects/</code> whenever I ask about past conversations. It would work. So why build something more complex?</p>

	<p><strong>Token efficiency.</strong> Every time Claude greps through gigabytes of JSONL files, those results flow through the conversation. That's tokens - and cost. A dedicated query system returns only what's needed, keeping the context window lean.</p>

	<p><strong>Privacy.</strong> When Claude searches via grep, my conversation history travels through Anthropic's servers as part of the prompt/response cycle. With a local MCP server, queries execute entirely on my machine. The search happens locally; only the results enter the conversation.</p>

	<p><strong>Speed.</strong> Grep is O(n) - it reads every byte of every file. An indexed database query is O(log n) or better. As my history grows past hundreds of megabytes, this difference becomes visceral. Milliseconds vs seconds.</p>

	<p><strong>Structured queries.</strong> Grep finds text matches. SQL lets me ask: "What projects did I work on between 10pm and 2am last month, sorted by total tokens used?" Aggregations, date ranges, joins across tables - things grep simply can't do.</p>

	<p><strong>Semantic search.</strong> "Find conversations about handling failures gracefully" can't be done with keyword grep. It requires embeddings and vector similarity - a fundamentally different capability.</p>

	<p>And honestly? <strong>I want to learn this stuff.</strong> Building MCP servers, embedding pipelines, local LLM infrastructure, analytical dashboards. Even if the practical benefits were marginal, the learning isn't. This is an excuse to build something real while exploring technologies I want to understand better.</p>

	<hr />

	<h2>What I'm Building</h2>

	<p>A complete memory system. Not just search - a full stack for understanding and retrieving my conversation history.</p>

	<p>Here's the architecture:</p>

	<figure class="article-image">
		<img src="/images/claude/memory-system/architecture.png" alt="8-phase memory system architecture" />
	</figure>

	<p>Eight phases. Each one adds a capability layer. Each one is independently useful.</p>

	<hr />

	<h2>The Phases</h2>

	<h3>Phase 1: Analytics Foundation (DuckDB)</h3>

	<figure class="article-image">
		<img src="/images/claude/memory-system/phase1-duckdb.png" alt="DuckDB analytics foundation" />
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
		<img src="/images/claude/memory-system/phase2-fts5.png" alt="Full-text search with ranking" />
	</figure>

	<p>Add full-text search with BM25 ranking.</p>

	<p><strong>What it unlocks:</strong></p>
	<ul>
		<li>Boolean queries: <code>authentication AND NOT oauth</code></li>
		<li>Phrase matching: <code>"rate limiting"</code></li>
		<li>Ranked results instead of "here's 500 matches, good luck"</li>
	</ul>

	<p>DuckDB supports full-text search via a loadable extension (<code>LOAD fts;</code>). Porter stemming means "running" matches "run", "runs", "runner". Case-insensitive by default.</p>

	<p>The limitation: you need to know the exact keywords. "auth" won't find "login".</p>

	<hr />

	<h3>Phase 3: Local LLM Setup (Mac Mini + Llama)</h3>

	<figure class="article-image">
		<img src="/images/claude/memory-system/phase3-localllm.png" alt="Local LLM setup" />
	</figure>

	<p>Set up Ollama on my Mac Mini for embeddings - converting text to vectors for semantic search.</p>

	<p>I'm planning to use <code>nomic-embed-text</code> or a similar embedding model. The choice will depend on quality vs speed tradeoffs I'll discover during implementation.</p>

	<p><strong>Why local instead of cloud APIs:</strong></p>
	<ul>
		<li>Embedding hundreds of megabytes of conversation history via OpenAI's API would cost real money. Local is free after setup.</li>
		<li>Everything stays on my network - no conversation data leaving my machines.</li>
		<li>No rate limits, no API quotas, no ongoing costs.</li>
	</ul>

	<p>The Mac Mini runs the models. My main Mac calls the Ollama API over the local network. Inference stays off my workstation, keeping it responsive.</p>

	<hr />

	<h3>Phase 4: Semantic Search (LanceDB)</h3>

	<figure class="article-image">
		<img src="/images/claude/memory-system/phase4-lancedb.png" alt="Semantic vector search" />
	</figure>

	<p>This is where it gets interesting.</p>

	<p>LanceDB is an embedded vector database - same philosophy as DuckDB (file-based, no server, just works) but designed for vector similarity search rather than SQL queries.</p>

	<p>The pipeline: extract messages from DuckDB → generate embeddings via Ollama → store vectors in LanceDB. DuckDB remains the source of truth for structured data; LanceDB handles the semantic layer.</p>

	<p>Now I can search by <em>meaning</em>, not keywords.</p>

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
		<img src="/images/claude/memory-system/phase5-mcp.png" alt="MCP server for agent access" />
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
		<img src="/images/claude/memory-system/phase6-dashboard.png" alt="Visual dashboard" />
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
		<img src="/images/claude/memory-system/phase7-voice.png" alt="Voice control interface" />
	</figure>

	<p>Ask questions out loud. Get visual answers.</p>

	<p>"What did I work on last Friday?" → Bar chart of projects with session counts.</p>

	<p>Web Speech API for voice input. Pattern matching for common queries. Llama fallback for anything complex.</p>

	<p>The dream: verbal queries while I'm thinking through a problem, answers appearing on screen without touching the keyboard.</p>

	<hr />

	<h3>Phase 8: Go Backend</h3>

	<figure class="article-image">
		<img src="/images/claude/memory-system/phase8-gobackend.png" alt="Go backend infrastructure" />
	</figure>

	<p>At my current heavy usage rate - multiple long sessions daily - I'm generating several hundred megabytes of conversation history per month, trending toward a gigabyte as usage increases.</p>

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
</ArticleLayout>
