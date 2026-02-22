<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="The Article That Made Me Realize I Didn't Create a Memory System After All"
	date="2026-02-22"
	description="I built a memory system for Claude Code. Then I read an article that made me realize it was just a well-organized chat log. Here's the audit, the gap, and the upgrade path."
>
	<figure class="article-image">
		<img src="/images/claude/memory-upgrade/IMG-034-MEMORY-AUDIT-HERO.png" alt="Memory system audit hero" />
	</figure>

	<p>Six weeks ago I built a memory system for my Claude Code sessions. DuckDB for structured storage, FTS5 for keyword search, LanceDB with nomic-embed-text for semantic search, a Go MCP server to expose it all. I wrote about every phase as I built it - the database design, the embedding pipeline, the dashboard.</p>

	<p>And I've been using it daily since. It's genuinely helpful in two ways. First, when I forget something specific, I can search back and find it. No huge surprise - that's the entire reason I built it. But the second use case I didn't anticipate: recovering crashed conversation threads. Say I'm deep in a session with Claude, working on something, and the terminal hangs, or I get pounded with API 500 errors, or my parallel agents drain the context window dry before compaction can kick in. Previously that meant starting over. Now I just open a new chat and say "go find the conversation we just had about X, I want to continue it." That alone has saved me a lot of friction.</p>

	<p>Then a few days ago I read <a href="https://x.com/rohit4verse/status/2012925228159295810">@rohit4verse's article on how to build an agent that never forgets</a> and realized that despite being useful, what I'd built was a well-organized chat log. Not a memory system. There was so much untapped potential.</p>

	<hr />

	<h2>The Sentence That Really Hit Me</h2>

	<p>Their framing hit immediately:</p>

	<blockquote>
		<p>"Here is what I thought memory meant: Keeping the conversation history and stuffing it into the context window. That works for about 10 exchanges. Then the context window fills up. So you truncate old messages. Now your agent forgets the user is vegan and recommends a steakhouse. You realize conversation history isn't memory - it's just a chat log."</p>
	</blockquote>

	<p>That's my system. I have 90,000 messages indexed across 472 sessions. I can search them. I can embed them. But when I ask "what components make up Artifex?" - a threat hunting system I'm developing and which I've discussed in hundreds of sessions - the search returns individual messages that each mention one or two components. Never the full picture. So even though the search works, the memory, that is a higher level of abstraction that distills and integrates events, does not.</p>

	<p>They go further with an example that maps directly to my situation:</p>

	<blockquote>
		<p>"After two weeks, the vector database had 500 entries. When the user asked, 'What did I tell you about my work situation?' the retrieval system returned fragments from 12 different conversations. The agent saw: 'I love my job' (Week 1), 'I'm thinking about quitting' (Week 2)... Which one is true? The agent had no idea."</p>
	</blockquote>

	<p>Reading this, I realized I have exactly this problem. When I renamed a component in Artifex from "Intuition" to "Thelema," both names coexist in my database. A search for the intuition engine returns the old name alongside the new one with no way to know which is current. Embeddings measure similarity, not "truth".</p>

	<hr />

	<h2>The Audit</h2>

	<figure class="article-image">
		<img src="/images/claude/memory-upgrade/IMG-035-THE-AUDIT.png" alt="The audit process" />
	</figure>

	<p>Before doing anything about it, I needed to understand exactly what I had. Not what I remembered building - what actually exists in the codebase. Memory systems should probably start with honest self-assessment, after all we know human memory ain't exactly perfect either.</p>

	<p>So I went through every phase of my original project plan, checked the documentation against the code, and produced a line-by-line audit.</p>

	<p><strong>What's actually working:</strong></p>
	<ul>
		<li>DuckDB with three tables (sessions, messages, tool_calls) - rebuilt hourly via cron</li>
		<li>FTS5 keyword search with BM25 scoring, Porter stemming, English stopwords</li>
		<li>Ollama running on my Mac Mini with nomic-embed-text (768-dim vectors)</li>
		<li>LanceDB semantic search with incremental embedding</li>
		<li>Go MCP server with 7 tools - the agent-facing interface</li>
		<li>A SvelteKit dashboard with heatmaps, charts, and a session browser</li>
	</ul>

	<p><strong>What I thought was working but isn't:</strong></p>
	<ul>
		<li>Incremental ingestion - nope, it does a full rebuild every hour. Drops all tables, re-parses everything.</li>
		<li>Hybrid search - never built. Keyword and semantic are completely separate paths.</li>
		<li>Duplicate handling - duplicate message IDs cause constraint errors in the cron log. I'd never checked.</li>
	</ul>

	<p><strong>What was planned but never started:</strong></p>
	<ul>
		<li>Voice control (Phase 7) - a design document exists, zero code</li>
		<li>Go backend server (Phase 8) - mentioned in the project overview, no code, no repo</li>
		<li>Security hardening (Phase 9) - a planning document with "considerations," no decisions made</li>
	</ul>

	<p>The audit was uncomfortable. I'd mentally filed some of these as "kinda done" when they were really stuck in the purgatory of ideasies. Writing systems for tracking knowledge when you're not tracking your own accurately - oh sweet irony.</p>

	<hr />

	<h2>The Gap</h2>

	<p>The article describes three layers:</p>

	<blockquote>
		<p>"Layer 1: Resources (Raw Data). The source of truth. Unprocessed logs, uploads, transcripts. Immutable and timestamped.<br />
		Layer 2: Items (Atomic Facts). Discrete facts extracted from resources.<br />
		Layer 3: Categories (Evolving Summaries). The high-level context."</p>
	</blockquote>

	<p>I have Layer 1 only. The JSONL files and DuckDB messages - that's it. Raw resources. No atomic facts extracted. No evolving summaries. No intelligence layer between storage and retrieval.</p>

	<p>Here's the gap in a table:</p>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Capability</th>
					<th>Rohan's Architecture</th>
					<th>My System</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>Fact extraction</td>
					<td>LLM extracts atomic facts</td>
					<td>Stores raw messages as-is</td>
				</tr>
				<tr>
					<td>Category summaries</td>
					<td>Evolving markdown profiles per topic</td>
					<td>Nothing</td>
				</tr>
				<tr>
					<td>Conflict resolution</td>
					<td>New facts overwrite outdated ones</td>
					<td>Contradictions coexist silently</td>
				</tr>
				<tr>
					<td>Time-decay scoring</td>
					<td>Recency boosts relevance</td>
					<td>All messages weighted equally</td>
				</tr>
				<tr>
					<td>Tiered retrieval</td>
					<td>Summaries first, drill down if needed</td>
					<td>Flat: keyword OR semantic, both return raw messages</td>
				</tr>
				<tr>
					<td>Memory maintenance</td>
					<td>Nightly consolidation, weekly summarization</td>
					<td>Nothing is ever pruned</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>The system I built stores conversations. The system Rohan describes manages knowledge. The difference is the intelligence layer in between - fact extraction, summarization, conflict resolution, decay.</p>

	<hr />

	<h2>The Decisions</h2>

	<figure class="article-image">
		<img src="/images/claude/memory-upgrade/IMG-036-THE-DECISIONS.png" alt="The decisions" />
	</figure>

	<p>With the audit in hand, I had to decide what to do about the original plan's unfinished items before starting the upgrade.</p>

	<p><strong>Voice Control: Dropped.</strong> The design document was ambitious - Web Speech API, pattern matching, NL-to-SQL via Llama. But in practice I barely use the dashboard UI. I interact with memory almost entirely through the terminal via MCP tools, and I already have Whisper running for speech-to-text. Building another voice interface doesn't make the memory itself smarter. It might be a fun learning project someday, but it shouldn't block pragmatic improvements to the system. Cut it.</p>

	<p><strong>Security Hardening: Deferred.</strong> The system is localhost-only with no network exposure. Hardening matters, but hardening a system that's about to gain new tables, pipelines, and tools probably means hardening it twice. Do it after the architecture stabilizes.</p>

	<p><strong>Dashboard Polish: Deferred.</strong> As I mentioned, I barely use the dashboard as it is. Polishing something I don't use based on a data model that's about to change would be a fanciful contrivance. If the upgrade creates real needs for visualization, I'll build to those needs. Not before.</p>

	<p><strong>Go Backend: Deferred, Evaluate Later.</strong> The Go backend was supposed to replace cron with a file-watching daemon and serve a dashboard API. But cron works for the pipelines I have. If five cron jobs running in sequence proves fragile, that's when a daemon becomes justified - and I'll know exactly what it needs to do. Premature abstraction is the enemy.</p>

	<hr />

	<h2>The Upgrade Path</h2>

	<p>The upgrade is seven phases, each independently deployable and measurable:</p>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Phase</th>
					<th>What We Build</th>
					<th>Key Question It Answers</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>0</td>
					<td>Benchmark suite</td>
					<td>How bad is it, exactly?</td>
				</tr>
				<tr>
					<td>1</td>
					<td>Time-decay scoring</td>
					<td>Can a single formula improve recency?</td>
				</tr>
				<tr>
					<td>2</td>
					<td>Fact extraction pipeline</td>
					<td>Can we distill messages into atomic knowledge?</td>
				</tr>
				<tr>
					<td>3</td>
					<td>Category summaries</td>
					<td>Can we synthesize facts into coherent profiles?</td>
				</tr>
				<tr>
					<td>4</td>
					<td>Tiered retrieval</td>
					<td>Can we answer most queries without touching raw messages?</td>
				</tr>
				<tr>
					<td>5</td>
					<td>Conflict resolution</td>
					<td>Can we detect and resolve contradictions automatically?</td>
				</tr>
				<tr>
					<td>6</td>
					<td>Memory maintenance</td>
					<td>Can we keep the system healthy with automated decay?</td>
				</tr>
				<tr>
					<td>7</td>
					<td>Knowledge graph (optional)</td>
					<td>Can entity relationships improve retrieval beyond flat search?</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>Phase 7 is exploratory - I'm not sure a knowledge graph is justified for a personal dev memory system, but I want to find out. Everything else is sequential, each phase building on the previous one.</p>

	<p>The critical design principle: everything is additive. No existing tables are modified. No existing tools are removed. The upgrade layers intelligence on top of a working foundation. If any phase fails or isn't worth the complexity, the system still works exactly as it does today.</p>

	<hr />

	<h2>What I Learned From the Audit Itself</h2>

	<figure class="article-image">
		<img src="/images/claude/memory-upgrade/IMG-037-WHAT-I-LEARNED.png" alt="What I learned" />
	</figure>

	<p>The most valuable part wasn't the gap analysis - it was being honest about what I'd actually built versus what I thought I'd built. Three lessons:</p>

	<p><strong>Document what exists, not what you planned.</strong> My project overview described eight phases. Five were done. One was partial. Three had design docs and nothing else. But until I checked the code, I'd have told you "six of eight phases are done." The design docs felt like progress. They weren't.</p>

	<p><strong>Auditing is design.</strong> The audit wasn't wasted time before the "real work" of building. The audit IS the work that makes building productive. I now know exactly which cron jobs run, which tables exist, which tools are registered, which features are vaporware. Every decision in the upgrade plan is grounded in that reality.</p>

	<p><strong>Read other people's architectures after you build, not before.</strong> If I'd read Rohan's article before building anything, I might have tried to implement the three-layer hierarchy from scratch. Instead, I built a solid storage-and-retrieval layer first, used it for six weeks, felt its limitations firsthand, and THEN read about the intelligence layer I was missing. The gap analysis was meaningful because I understood both sides - my system's strengths and its blind spots.</p>

	<p>The next step is Phase 0: building a benchmark suite to put numbers on exactly how the current system performs. Before upgrading anything, measure what you have. That's the subject of the next article.</p>
</ArticleLayout>
