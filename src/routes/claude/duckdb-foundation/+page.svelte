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
	<title>I Turned Claude's Hidden Memory Into a Queryable Database | Faan Rossouw</title>
	<meta name="description" content="Building the analytics foundation for Claude's conversation history. 500MB of JSONL compressed to 69MB in DuckDB, with millisecond queries." />
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
				<span class="date">2026-01-14</span>
				<h1>I Turned Claude's Hidden Memory Into a Queryable Database</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				<p>A week ago, I discovered that Claude Code stores every conversation locally. Complete transcripts. Tool calls. Timestamps. Hundreds of megabytes of institutional knowledge, just sitting there in <code>~/.claude/projects/</code>.</p>

				<p>I wrote about that discovery. Felt like I'd found buried treasure.</p>

				<p>Then I actually tried to use it.</p>

				<p>"What was that approach we used for the authentication flow?" I knew we'd solved it. I remembered the shape of the solution. Grep returned 47 matches. I scrolled through walls of JSON, looking for the one conversation that mattered.</p>

				<p>Twenty minutes later, I gave up and re-explained the whole context to Claude from scratch.</p>

				<p>That's when the frustration crystallized into something specific: having the data isn't enough. I needed to make it <em>accessible</em>.</p>

				<figure class="article-image">
					<img src="/images/claude/duckdb-foundation/hero.png" alt="Converting raw conversation data into a queryable database" />
				</figure>

				<hr />

				<h2>The Friction I'm Fighting</h2>

				<p>My mantra has been simple: <strong>reduce friction to the minimum required to fully manifest an idea. Every unnecessary step between thought and creation is a leak in the system.</strong></p>

				<p>That's what this whole series is about. Every article. Every build. It all comes back to this: the gap between "I want X" and "X exists" is still too wide.</p>

				<p>Claude Code has done more to narrow that gap than any tool I've used. But there's a leak in the system. A big one.</p>

				<p>Every session starts fresh. Claude doesn't remember what we built yesterday. Doesn't know the decisions we made last week or the patterns that emerged from a month of collaboration. All of that context - hundreds of hours of shared work - evaporates between sessions.</p>

				<p>So I re-explain. Re-establish context. Re-discover solutions we've already found.</p>

				<p>That's friction. That's the leak.</p>

				<p>The conversation history exists. It's all there on disk. But it's locked in flat files - raw, unindexed, full of untapped potential. A few additional layers could transform it from data I technically have into knowledge I can actually use.</p>

				<p>This project is about fixing that. This article is step one.</p>

				<hr />

				<h2>Starting Simple: Just Make It Queryable</h2>

				<p>The temptation was to build everything at once. Semantic search. Natural language queries. A beautiful dashboard. The whole vision.</p>

				<p>I've learned to resist that.</p>

				<p>Step one: can I even query this data with SQL? Get the structure right. Prove the foundation works. Everything else builds on that.</p>

				<p>I needed a database that fit the constraints:</p>

				<ul>
					<li>Local (no cloud, no servers to maintain)</li>
					<li>Analytical (good at aggregations and filtering, not transactions)</li>
					<li>Simple (one file, no infrastructure)</li>
				</ul>

				<p>DuckDB kept coming up. Embedded analytical database. Think SQLite but optimized for the kind of queries I wanted to run. I'd heard good things. Time to actually try it.</p>

				<pre><code>brew install duckdb</code></pre>

				<p>That was the easy part.</p>

				<hr />

				<h2>Wrestling With the Data</h2>

				<p>The JSONL files looked simple enough when I first opened them. One JSON object per line. How hard could it be?</p>

				<p>Harder than I expected.</p>

				<p>First issue: the folder structure encodes the working directory with dashes.</p>

				<pre><code>-Users-faanross-Desktop-obs_vault-faanross-project</code></pre>

				<p>That needs to decode back to:</p>

				<pre><code>/Users/faanross/Desktop/obs_vault/faanross/project</code></pre>

				<p>Straightforward string manipulation. Fine.</p>

				<p>Second issue: the message format varies. I wrote my first parser expecting <code>message</code> to be a simple string.</p>

				<pre><code>{`content = record['message']  # This will work, right?`}</code></pre>

				<p>It crashed immediately. Some messages are strings. Some are arrays. Some are nested objects with a <code>content</code> array inside. Claude Code uses different formats depending on whether it's user input, a tool result, or an API response.</p>

				<pre><code>{`// Sometimes this:
{"message": "hello"}

// Sometimes this:
{"message": [{"type": "text", "text": "..."}]}

// Sometimes this:
{"message": {"content": [{"type": "tool_use", "name": "Bash", ...}]}}`}</code></pre>

				<p>I spent longer than I'd like to admit handling all the edge cases. Recursive flattening. Type checking. Null handling. The kind of grunt work that doesn't feel like progress until suddenly everything works.</p>

				<figure class="article-image">
					<img src="/images/claude/duckdb-foundation/dataflow.png" alt="Data flow from JSONL files through parsing to DuckDB" />
				</figure>

				<hr />

				<h2>The Schema That Emerged</h2>

				<p>After a few iterations, I landed on three tables:</p>

				<p><strong>Sessions</strong> - one row per conversation</p>

				<pre><code>{`CREATE TABLE sessions (
    session_id VARCHAR PRIMARY KEY,
    project_path VARCHAR,
    project_name VARCHAR,
    first_message_at TIMESTAMP,
    last_message_at TIMESTAMP,
    message_count INTEGER
);`}</code></pre>

				<p><strong>Messages</strong> - every message in every session</p>

				<pre><code>{`CREATE TABLE messages (
    id VARCHAR PRIMARY KEY,
    session_id VARCHAR,
    type VARCHAR,           -- 'user' or 'assistant'
    timestamp TIMESTAMP,
    content TEXT,
    tool_name VARCHAR,
    cwd VARCHAR
);`}</code></pre>

				<p><strong>Tool Calls</strong> - every tool Claude used</p>

				<pre><code>{`CREATE TABLE tool_calls (
    id VARCHAR PRIMARY KEY,
    session_id VARCHAR,
    tool_name VARCHAR,
    input_json TEXT,
    timestamp TIMESTAMP
);`}</code></pre>

				<p>Three levels of granularity. Sessions for patterns. Messages for content. Tool calls for understanding how I actually work with Claude.</p>

				<hr />

				<h2>The First Query That Worked</h2>

				<p>The moment everything came together:</p>

				<pre><code>python ingest.py</code></pre>

				<p>Thirty seconds of processing. Then:</p>

				<pre><code>{`Sessions ingested: 367
Messages ingested: 52,764
Tool calls ingested: 15,825`}</code></pre>

				<p>I opened the DuckDB CLI and ran my first real query:</p>

				<pre><code>{`SELECT project_name, COUNT(*) as sessions
FROM sessions
GROUP BY project_name
ORDER BY sessions DESC
LIMIT 5;`}</code></pre>

				<p>Results appeared instantly. My projects, ranked by how much time I'd spent in each. Data I'd never seen before, from conversations I'd already forgotten having.</p>

				<p>The raw JSONL files weighed around 500MB. The DuckDB database? 69MB. Seven-to-one compression, and queries that would have taken grep minutes now returned in milliseconds.</p>

				<figure class="article-image">
					<img src="/images/claude/duckdb-foundation/compression.png" alt="Compression comparison: 500MB JSONL to 69MB DuckDB" />
				</figure>

				<hr />

				<h2>What the Data Showed Me</h2>

				<p>Here's where it got interesting. I thought I knew how I worked. The data told a different story.</p>

				<p><strong>My peak productivity hours:</strong></p>

				<pre><code>{`SELECT EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as messages
FROM messages WHERE type = 'user'
GROUP BY hour ORDER BY messages DESC LIMIT 3;`}</code></pre>

				<figure class="article-image">
					<img src="/images/claude/duckdb-foundation/scr-hours.png" alt="Terminal showing peak productivity hours query results" />
				</figure>

				<p>9pm and 11am. Two distinct windows - late morning and late evening. Not what I would have guessed.</p>

				<p><strong>How I actually use Claude:</strong></p>

				<pre><code>{`SELECT tool_name, COUNT(*) as uses,
       ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as pct
FROM tool_calls
GROUP BY tool_name
ORDER BY uses DESC LIMIT 5;`}</code></pre>

				<figure class="article-image">
					<img src="/images/claude/duckdb-foundation/scr-tools.png" alt="Terminal showing tool usage query results" />
				</figure>

				<div class="data-table">
					<table>
						<thead>
							<tr>
								<th>tool_name</th>
								<th>uses</th>
								<th>pct</th>
							</tr>
						</thead>
						<tbody>
							<tr><td>Edit</td><td>3726</td><td>23.5%</td></tr>
							<tr><td>Bash</td><td>3700</td><td>23.4%</td></tr>
							<tr><td>Read</td><td>3352</td><td>21.2%</td></tr>
							<tr><td>Write</td><td>1764</td><td>11.1%</td></tr>
							<tr><td>TodoWrite</td><td>1320</td><td>8.3%</td></tr>
						</tbody>
					</table>
				</div>

				<p>Edit and Read dominate. I'm modifying existing code twice as often as writing new files. The iterative refinement pattern I felt in practice, now confirmed by data.</p>

				<p><strong>Where the deep work happens:</strong></p>

				<pre><code>{`SELECT project_name, message_count, DATE(first_message_at) as date
FROM sessions ORDER BY message_count DESC LIMIT 3;`}</code></pre>

				<figure class="article-image">
					<img src="/images/claude/duckdb-foundation/scr-projects.png" alt="Terminal showing deep work sessions query results" />
				</figure>

				<p>One session hit nearly 10,000 messages. A single conversation, sustained over hours. I remember that day (working on my C2 framework Numinon) - deep in a complex build, completely in flow. The data captured the trace of what that felt like.</p>

				<hr />

				<h2>The Pattern Recognition Problem</h2>

				<p>But here's what the data also showed me: the foundation isn't enough.</p>

				<p>I searched for "authentication":</p>

				<pre><code>{`SELECT * FROM messages WHERE content LIKE '%authentication%';`}</code></pre>

				<p>47 results. Equal weight. No ranking. The important conversation buried among casual mentions.</p>

				<p>I searched for "login" hoping to find related discussions. Nothing. Different word, no match. The conceptual connection between "authentication" and "login" is obvious to me - invisible to a LIKE clause.</p>

				<p>This is exactly the limitation I'm building toward fixing.</p>

				<p>Phase 2 adds full-text search with ranking. Phase 3 adds semantic embeddings - search by meaning, not keywords. Phase 4 gives Claude direct access to query its own history.</p>

				<p>Each phase removes a friction point. Each phase makes the path from question to answer shorter.</p>

				<p>But they all need this foundation first. You can't build search without something to search. You can't build analytics without structured data.</p>

				<hr />

				<h2>Automation: Keeping It Current</h2>

				<p>The database needs to stay fresh. New conversations happen constantly.</p>

				<p>I set up a cron job - hourly ingestion:</p>

				<pre><code>{`0 * * * * /path/to/venv/bin/python /path/to/ingest.py >> /path/to/cron.log 2>&1`}</code></pre>

				<p>It's crude. Eventually, I'll have a proper backend with file watching and incremental updates. But right now, hourly is enough. The database stays current. Queries reflect reality.</p>

				<p>I checked the cron log the next morning. Ingestion had run eight times overnight. New sessions picked up automatically. Zero manual work.</p>

				<p>This is what reducing friction looks like. Not grand gestures - small automations that compound. Every time I don't have to think about refreshing the database, that's cognitive load I can spend elsewhere.</p>

				<hr />

				<h2>What I Learned</h2>

				<p><strong>The data structure is messier than it looks.</strong> Three different message formats in the same dataset. Edge cases everywhere. Handle them all or watch the parser crash on production data.</p>

				<p><strong>Compression is dramatic.</strong> 7:1 wasn't what I expected. DuckDB's columnar storage handles repetitive strings (session IDs, tool names, working directories) extremely well.</p>

				<p><strong>Objective measurement beats intuition.</strong> I was wrong about when I work best, which tools I use most, which projects get my deepest attention. The data corrected my assumptions.</p>

				<p><strong>Foundation work isn't glamorous, but it's essential.</strong> No flashy demos here. Just structured data in a database. But everything that follows depends on this.</p>

				<hr />

				<h2>The Bigger Picture</h2>

				<p>This is Part 1 of a larger build. The database is the foundation. What comes next:</p>

				<ul>
					<li><strong>Phase 2:</strong> Full-text search with BM25 ranking</li>
					<li><strong>Phase 3:</strong> Local LLM setup to keep data private and minimize costs</li>
					<li><strong>Phase 4:</strong> Semantic embeddings with LanceDB for meaning-based queries</li>
					<li><strong>Phase 5:</strong> MCP server so Claude can query its own history</li>
					<li><strong>Phase 6:</strong> Visual dashboard for pattern exploration</li>
					<li><strong>Phase 7:</strong> Voice control and search</li>
					<li><strong>Phase 8:</strong> Custom Go backend for production-grade performance</li>
					<li><strong>Phase 9:</strong> Security hardening across the entire infrastructure</li>
				</ul>

				<p>Each phase removes friction. Each phase makes the collaboration more cumulative, less amnesiac.</p>

				<p>The goal isn't just "search my conversations." It's this: <strong>every session with Claude should build on everything that came before.</strong> Context that persists. Knowledge that compounds. An AI partner that remembers what we've learned together.</p>

				<p>We're not there yet. But we're one step closer.</p>

				<p>The data is structured. The queries work. The foundation is laid.</p>

				<p>Time to build on it.</p>

				<hr />

				<h2>Quick Reference</h2>

				<p><strong>Location:</strong> <code>~/repos/claude-memory/</code></p>

				<p><strong>Run ingestion:</strong></p>

				<pre><code>{`cd ~/repos/claude-memory
source venv/bin/activate
python ingest.py`}</code></pre>

				<p><strong>Query interactively:</strong></p>

				<pre><code>duckdb ~/repos/claude-memory/memory.duckdb</code></pre>

				<p><strong>Useful queries:</strong></p>

				<pre><code>{`-- Overview stats
SELECT
    (SELECT COUNT(*) FROM sessions) as sessions,
    (SELECT COUNT(*) FROM messages) as messages,
    (SELECT COUNT(*) FROM tool_calls) as tool_calls;

-- Search for a term
SELECT * FROM messages WHERE content LIKE '%pattern%' LIMIT 20;

-- Recent sessions
SELECT project_name, first_message_at, message_count
FROM sessions ORDER BY first_message_at DESC LIMIT 10;

-- Productivity by hour
SELECT EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as messages
FROM messages WHERE type = 'user'
GROUP BY hour ORDER BY hour;`}</code></pre>

				<hr />

				<p><em>Part 1 of the Claude Memory Project. Next: adding full-text search with BM25 ranking.</em></p>

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
		margin: 32px 0 16px;
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

	.data-table {
		margin: 24px 0;
		overflow-x: auto;
	}

	.data-table table {
		width: 100%;
		border-collapse: collapse;
		font-size: 15px;
	}

	.data-table th,
	.data-table td {
		padding: 12px 16px;
		text-align: left;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.data-table th {
		color: var(--white);
		font-weight: 600;
		background: rgba(189, 147, 249, 0.1);
	}

	.data-table td {
		color: rgba(255, 255, 255, 0.85);
	}

	.data-table tr:hover td {
		background: rgba(255, 255, 255, 0.03);
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

		.data-table {
			font-size: 14px;
		}

		.data-table th,
		.data-table td {
			padding: 10px 12px;
		}
	}
</style>
