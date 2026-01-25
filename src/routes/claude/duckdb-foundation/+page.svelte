<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="I Turned Claude's Hidden Memory Into a Queryable Database"
	date="2026-01-14"
	description="Building the analytics foundation for Claude's conversation history. 500MB of JSONL compressed to 69MB in DuckDB, with millisecond queries."
>
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

	<p>I spent longer than I'd like to admit handling all the edge cases. The core extraction logic ended up looking something like this:</p>

	<pre><code>{`def extract_content(message):
    """Handle the three different message formats."""
    if message is None:
        return ""
    if isinstance(message, str):
        return message
    if isinstance(message, list):
        # Array of content blocks
        return " ".join(
            block.get("text", "")
            for block in message
            if isinstance(block, dict)
        )
    if isinstance(message, dict):
        # Nested structure - recurse into 'content'
        return extract_content(message.get("content", ""))
    return str(message)`}</code></pre>

	<p>Recursive flattening. Type checking at every level. The kind of grunt work that doesn't feel like progress until suddenly everything works.</p>

	<figure class="article-image">
		<img src="/images/claude/duckdb-foundation/dataflow.png" alt="Data flow from JSONL files through parsing to DuckDB" />
	</figure>

	<hr />

	<h2>The Schema That Emerged</h2>

	<p>After a few iterations, I landed on three tables. This is a simplified schema - the raw JSONL contains more fields (<code>uuid</code>, <code>parentUuid</code>, <code>gitBranch</code>, <code>version</code>, etc.) but I focused on what I actually wanted to query.</p>

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
    type VARCHAR,           -- 'user', 'assistant', 'progress', 'system', or 'file-history-snapshot'
    timestamp TIMESTAMP,
    content TEXT,
    tool_name VARCHAR,      -- populated for tool_use content blocks
    cwd VARCHAR
);`}</code></pre>

	<p>Note: I <a href="/claude/hidden-memory">discovered earlier</a> that JSONL files contain five record types, not just user/assistant. For this foundation, I'm primarily interested in <code>user</code> and <code>assistant</code> messages, but the schema captures all types for completeness.</p>

	<p><strong>Tool Calls</strong> - every tool Claude used</p>

	<pre><code>{`CREATE TABLE tool_calls (
    id VARCHAR PRIMARY KEY,
    session_id VARCHAR,
    message_id VARCHAR,     -- links back to the assistant message
    tool_name VARCHAR,
    input_json TEXT,        -- the full tool input as JSON string
    timestamp TIMESTAMP
);`}</code></pre>

	<p>Three levels of granularity. Sessions for high-level patterns. Messages for content search. Tool calls for understanding how I actually work with Claude - which tools get used, in what combinations, for which projects.</p>

	<hr />

	<h2>The First Query That Worked</h2>

	<p>The moment everything came together:</p>

	<pre><code>python ingest.py</code></pre>

	<!-- TODO: Replace with actual screenshot of ingest.py terminal output -->
	<figure class="article-image placeholder">
		<img src="/images/claude/duckdb-foundation/scr-ingest-output.png" alt="Terminal showing ingest.py output with session, message, and tool call counts" />
		<figcaption class="placeholder-note">📸 TODO: Actual terminal screenshot of ingest.py running</figcaption>
	</figure>

	<p>I opened the DuckDB CLI and ran my first real query:</p>

	<pre><code>{`SELECT project_name, COUNT(*) as sessions
FROM sessions
GROUP BY project_name
ORDER BY sessions DESC
LIMIT 5;`}</code></pre>

	<p>Results appeared instantly. My projects, ranked by how much time I'd spent in each. Data I'd never seen before, from conversations I'd already forgotten having.</p>

	<p>The compression was dramatic. To verify:</p>

	<pre><code>{`# Check raw JSONL size
du -sh ~/.claude/projects/

# Check DuckDB size
du -sh ~/repos/claude-memory/memory.duckdb`}</code></pre>

	<!-- TODO: Replace with actual screenshot showing file sizes -->
	<figure class="article-image placeholder">
		<img src="/images/claude/duckdb-foundation/scr-compression.png" alt="Terminal showing du -sh output comparing JSONL and DuckDB sizes" />
		<figcaption class="placeholder-note">📸 TODO: Actual terminal screenshot of du -sh commands</figcaption>
	</figure>

	<p>In my case: roughly 7:1 compression. DuckDB's columnar storage handles repetitive strings (session IDs, tool names, working directories) extremely efficiently.</p>

	<hr />

	<h2>What the Data Showed Me</h2>

	<p>Here's where it got interesting. I thought I knew how I worked. The data told a different story.</p>

	<p><em>Note: Run these queries against your own data - your patterns will be different. That's the point.</em></p>

	<p><strong>My peak productivity hours:</strong></p>

	<pre><code>{`SELECT EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as messages
FROM messages WHERE type = 'user'
GROUP BY hour ORDER BY messages DESC LIMIT 5;`}</code></pre>

	<!-- TODO: Replace with actual screenshot of this query result -->
	<figure class="article-image placeholder">
		<img src="/images/claude/duckdb-foundation/scr-hours.png" alt="Terminal showing peak productivity hours query results" />
		<figcaption class="placeholder-note">📸 TODO: Actual DuckDB CLI screenshot of hours query</figcaption>
	</figure>

	<p>For me, two distinct windows emerged - late morning and late evening. Not what I would have guessed. Your data will reveal your own patterns.</p>

	<p><strong>How I actually use Claude:</strong></p>

	<pre><code>{`SELECT tool_name, COUNT(*) as uses,
       ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as pct
FROM tool_calls
GROUP BY tool_name
ORDER BY uses DESC LIMIT 10;`}</code></pre>

	<!-- TODO: Replace with actual screenshot of this query result -->
	<figure class="article-image placeholder">
		<img src="/images/claude/duckdb-foundation/scr-tools.png" alt="Terminal showing tool usage query results" />
		<figcaption class="placeholder-note">📸 TODO: Actual DuckDB CLI screenshot of tool usage query</figcaption>
	</figure>

	<p>In my case, Edit and Read dominated. I was modifying existing code roughly twice as often as writing new files. The iterative refinement pattern I felt in practice, now confirmed by data. Bash was close behind - lots of git operations, builds, and exploratory commands.</p>

	<p><strong>Where the deep work happens:</strong></p>

	<pre><code>{`SELECT project_name, message_count,
       DATE(first_message_at) as started,
       DATE(last_message_at) as ended
FROM sessions
ORDER BY message_count DESC
LIMIT 5;`}</code></pre>

	<!-- TODO: Replace with actual screenshot of this query result -->
	<figure class="article-image placeholder">
		<img src="/images/claude/duckdb-foundation/scr-projects.png" alt="Terminal showing deep work sessions query results" />
		<figcaption class="placeholder-note">📸 TODO: Actual DuckDB CLI screenshot of deep sessions query</figcaption>
	</figure>

	<p>My longest sessions ran into the thousands of messages - sustained conversations over hours of deep work. The data captured the trace of what flow state looks like in a conversation log.</p>

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

	<pre><code>{`# Edit crontab
crontab -e

# Add this line (adjust paths to your setup):
0 * * * * ~/repos/claude-memory/venv/bin/python ~/repos/claude-memory/ingest.py >> ~/repos/claude-memory/cron.log 2>&1`}</code></pre>

	<p>It's crude. Eventually, I'll have a proper backend with file watching and incremental updates. But right now, hourly is enough. The database stays current. Queries reflect reality.</p>

	<p>To verify it's working:</p>

	<pre><code>{`# Check recent cron executions
tail -20 ~/repos/claude-memory/cron.log

# Or check the database modification time
ls -la ~/repos/claude-memory/memory.duckdb`}</code></pre>

	<p>I checked the cron log the next morning. Ingestion had run through the night, picking up new sessions automatically. Zero manual work.</p>

	<p>This is what reducing friction looks like. Not grand gestures - small automations that compound. Every time I don't have to think about refreshing the database, that's cognitive load I can spend elsewhere.</p>

	<hr />

	<h2>What I Learned</h2>

	<p><strong>The data structure is messier than it looks.</strong> Three different message formats in the same dataset. Edge cases everywhere. Handle them all or watch the parser crash on production data.</p>

	<p><strong>Compression is dramatic.</strong> 7:1 wasn't what I expected. DuckDB's columnar storage handles repetitive strings (session IDs, tool names, working directories) extremely well.</p>

	<p><strong>Objective measurement beats intuition.</strong> I was wrong about when I work best, which tools I use most, which projects get my deepest attention. The data corrected my assumptions.</p>

	<p><strong>Foundation work isn't glamorous, but it's essential.</strong> No flashy demos here. Just structured data in a database. But everything that follows depends on this.</p>

	<hr />

	<h2>The Bigger Picture</h2>

	<p>This is Phase 1 of an <a href="/claude/memory-system">8-phase build</a>. The database is the foundation. What comes next:</p>

	<ul>
		<li><strong>Phase 2:</strong> Full-text search with BM25 ranking</li>
		<li><strong>Phase 3:</strong> Local LLM setup for embeddings (privacy + cost)</li>
		<li><strong>Phase 4:</strong> Semantic search with LanceDB for meaning-based queries</li>
		<li><strong>Phase 5:</strong> MCP server so Claude can query its own history</li>
		<li><strong>Phase 6:</strong> Visual dashboard for pattern exploration</li>
		<li><strong>Phase 7:</strong> Voice control interface</li>
		<li><strong>Phase 8:</strong> Go backend for scale</li>
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
</ArticleLayout>

<style>
	.article-image.placeholder {
		border: 2px dashed rgba(189, 147, 249, 0.4);
		border-radius: 12px;
		padding: 16px;
		background: rgba(189, 147, 249, 0.05);
	}

	.placeholder-note {
		text-align: center;
		font-size: 14px;
		color: var(--aion-purple);
		margin-top: 12px;
		font-style: italic;
	}
</style>
