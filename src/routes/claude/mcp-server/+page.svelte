<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import ScrollProgress from '$lib/components/ScrollProgress.svelte';

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
	<title>I Gave My AI Access to Its Own Memory | Faan Rossouw</title>
	<meta name="description" content="Building an MCP server so Claude can search its own conversation history. From manual queries to autonomous memory access." />
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
				<span class="date">2026-01-25</span>
				<h1>I Gave My AI Access to Its Own Memory</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				<figure class="article-image">
					<img src="/images/claude/mcp-server/IMG-001-HERO.png" alt="Claude accessing its own memory through MCP" />
				</figure>

				<p>The semantic search was working. I could find conversations by meaning, not just keywords. "Retry logic when things fail" found discussions about exponential backoff even though I'd never used that exact phrase.</p>

				<p>But there was still friction.</p>

				<p>Every time I wanted to check if we'd discussed something before, I had to open a terminal, activate a virtual environment, run a Python script, read the output, then copy relevant bits back to Claude. I was the middleman between my AI assistant and its own history.</p>

				<p>The data was accessible. The search was powerful. But Claude couldn't use any of it directly.</p>

				<p>That's the leak I set out to fix.</p>

				<hr />

				<h2>The Gap</h2>

				<p>Phase 4 gave me semantic search. I could run queries like:</p>

				<pre><code>python search.py "handling errors gracefully" --limit 5</code></pre>

				<p>And get back conceptually related conversations - retry logic, circuit breakers, exponential backoff. Powerful stuff.</p>

				<p>But the workflow looked like this:</p>

				<pre><code>Me: "Have we discussed authentication rate limiting before?"
Claude: "I don't have access to our previous conversations."
Me: *opens terminal*
Me: *runs search*
Me: *copies results*
Me: "Here's what we discussed last week..."
Claude: "Ah yes, based on that context..."</code></pre>

				<p>Every lookup added minutes of friction. I was the translation layer between Claude and its own memory.</p>

				<p>The obvious solution: give Claude direct access.</p>

				<hr />

				<h2>Why MCP</h2>

				<p>MCP - Model Context Protocol - is how Claude Code talks to external tools. When Claude needs to read a file, it calls the Read tool. When it needs to run a command, it calls Bash. These aren't magic; they're tools exposed through MCP servers.</p>

				<p>I needed to build my own MCP server. One that exposes my memory system as tools Claude can call directly.</p>

				<p>The alternative would be simpler: bash scripts.</p>

				<pre><code># Claude could just run:
~/tools/memory-search.sh "authentication"</code></pre>

				<p>Mario Zechner made this argument in his article <a href="https://mariozechner.at/posts/2025-11-02-what-if-you-dont-need-mcp/" target="_blank" rel="noopener">"What if you don't need MCP at all?"</a> - and he's not wrong. MCP adds complexity. Some popular MCP servers expose 18,000 tokens worth of tool definitions. That's overhead.</p>

				<p>But my server would be minimal. Five tools. Maybe 500 tokens of definitions. And MCP gives me structured input/output, automatic discovery, native integration with Claude Code's architecture.</p>

				<p>I decided to build it properly. If it became painful, I could always fall back to bash.</p>

				<figure class="article-image">
					<img src="/images/claude/mcp-server/IMG-002-ARCHITECTURE.png" alt="MCP server architecture diagram" />
				</figure>

				<hr />

				<h2>Building the Server</h2>

				<p>I chose Go. The planned backend will be in Go and it's the language I'm most experienced in, so it just made sense. Plus <a href="https://github.com/marcboeker/go-duckdb" target="_blank" rel="noopener">go-duckdb</a> is mature enough for production use, and a single binary is easier to deploy than a Node runtime.</p>

				<p>The official MCP SDK had just hit v1.2.0 (as of January 2026) - stable, maintained with Google, cleaner API with generics. Good timing.</p>

				<pre><code>mkdir -p ~/repos/claude-memory-mcp
cd ~/repos/claude-memory-mcp
go mod init github.com/faanross/claude-memory-mcp</code></pre>

				<p>Project structure:</p>

				<pre><code>claude-memory-mcp/
├── cmd/server/main.go       # Entry point
├── internal/
│   ├── db/duckdb.go         # Database connection
│   └── tools/
│       ├── search.go        # Keyword search
│       ├── semantic.go      # Semantic search
│       └── analytics.go     # Stats, sessions
└── claude-memory-mcp        # Built binary (7.5MB)</code></pre>

				<p>The first real lesson came immediately: <strong>log to stderr</strong>. MCP uses stdout for the protocol. If your debug logs go to stdout, you corrupt the JSON-RPC stream. Your server "works" but Claude gets garbage. I lost time here before checking the MCP spec.</p>

				<hr />

				<h2>The Tools</h2>

				<p>I decided to implement five tools, each serving a different need:</p>

				<div class="data-table">
					<table>
						<thead>
							<tr>
								<th>Tool</th>
								<th>Purpose</th>
								<th>Backend</th>
							</tr>
						</thead>
						<tbody>
							<tr>
								<td><code>search_memory</code></td>
								<td>Keyword search with BM25 ranking</td>
								<td>DuckDB FTS5</td>
							</tr>
							<tr>
								<td><code>semantic_search</code></td>
								<td>Find by meaning</td>
								<td>LanceDB vectors</td>
							</tr>
							<tr>
								<td><code>get_project_stats</code></td>
								<td>Analytics for a project</td>
								<td>DuckDB</td>
							</tr>
							<tr>
								<td><code>recent_sessions</code></td>
								<td>List recent conversations</td>
								<td>DuckDB</td>
							</tr>
							<tr>
								<td><code>get_session</code></td>
								<td>Retrieve full session content</td>
								<td>DuckDB</td>
							</tr>
						</tbody>
					</table>
				</div>

				<p>The keyword search was straightforward - just wrapping the FTS queries I'd already built. (Note: I keep calling it "FTS5" out of habit from SQLite, but DuckDB has its own FTS extension with similar functionality - not actually SQLite's FTS5.) Connect to DuckDB, run the query, return formatted results.</p>

				<p>But one decision shaped everything: read-only access.</p>

				<pre><code class="language-go">db, err := sql.Open("duckdb", dbPath+"?access_mode=read_only")
if err != nil {'{'}
    return nil, fmt.Errorf("failed to open database: %w", err)
{'}'}</code></pre>

				<p>This wasn't just a technical choice - it was a trust boundary. I was building a tool that Claude would use autonomously, without me reviewing every action. The MCP server needed to be safe by design.</p>

				<p>Read-only means Claude can query anything: search conversations, pull analytics, retrieve sessions. But it can never modify the history. No accidental deletions. No corrupted data. No "I was trying to help and deleted everything."</p>

				<p>This constraint is what makes autonomous use possible. I don't need to approve every tool call because the worst case is a query that returns nothing useful. The history remains intact no matter what.</p>

				<p>There's a broader principle here worth noting: when designing agents, what the agent <em>cannot</em> do matters as much as what it <em>can</em>. We focus on capabilities - what tools to expose, what goals to set. But guardrails are equally critical. They're not limitations; they're what make autonomy safe to grant.</p>

				<hr />

				<h2>The LanceDB Problem</h2>

				<p>Semantic search was trickier - LanceDB doesn't have official Go bindings.</p>

				<p>I'd built the semantic search in Python during Phase 4. The embedding model runs on my Mac Mini via Ollama, the vectors live in a LanceDB directory, and a Python script ties it together. It works.</p>

				<p>Now I needed to call it from Go.</p>

				<p>I considered three approaches:</p>

				<ol>
					<li><strong>Rewrite in Go</strong> - No LanceDB bindings. Dead end.</li>
					<li><strong>HTTP wrapper</strong> - Spin up a Flask server, call it from Go. Adds a long-running service to maintain.</li>
					<li><strong>Shell out to Python</strong> - Just call the existing script.</li>
				</ol>

				<p>Option 3 felt inelegant. Shelling out to Python from a Go binary? But it reuses tested code, works now, and I can migrate later if LanceDB adds Go support.</p>

				<p>I added a <code>--json</code> flag to <code>search.py</code>:</p>

				<pre><code class="language-python">if args.json:
    print(json.dumps(results))</code></pre>

				<p>Then from Go:</p>

				<pre><code class="language-go">cmd := exec.Command(pythonPath, scriptPath, query, "--json", "--limit", strconv.Itoa(limit))
output, err := cmd.Output()
if err != nil {'{'}
    return nil, fmt.Errorf("semantic search failed: %w", err)
{'}'}</code></pre>

				<p><em>(Security note: <code>exec.Command</code> passes arguments separately - no shell interpolation. This is intentionally safe from injection attacks. If I'd used <code>exec.Command("bash", "-c", "python " + query)</code>, that would be vulnerable. The pattern above is the secure approach.)</em></p>

				<p>Not beautiful, a bit of a hack TBH. But functional. Ships today, not someday.</p>

				<hr />

				<h2>Registration Confusion</h2>

				<p>With tools implemented, I added the server to Claude Code's config:</p>

				<pre><code class="language-json">"claude-memory": {'{'}
  "type": "stdio",
  "command": "/Users/faanross/repos/claude-memory-mcp/claude-memory-mcp",
  "args": []
{'}'}</code></pre>

				<p>Restarted Claude Code. Ran <code>/mcp</code> to verify.</p>

				<p>The server wasn't listed.</p>

				<p>I double-checked the JSON syntax. Rebuilt the binary. Restarted again. Nothing.</p>

				<p>Then I found the problem: I was editing the wrong file.</p>

				<p>Two config files exist:</p>
				<ul>
					<li><code>~/.claude/settings.json</code> - Where I added the MCP</li>
					<li><code>~/.claude.json</code> - Where Claude Code actually reads user-level MCPs</li>
				</ul>

				<p>The naming is... unfortunate. I'd been editing <code>settings.json</code> because that's where other config lived. But MCP servers need to be in <code>~/.claude.json</code>.</p>

				<pre><code># Check which servers are actually registered
cat ~/.claude.json | jq '.mcpServers | keys'</code></pre>

				<p>Moved the config to the right file. Restarted. The server appeared.</p>

				<figure class="article-image">
					<img src="/images/claude/mcp-server/IMG-003-REGISTRATION.png" alt="MCP server registration in Claude Code" />
				</figure>

				<hr />

				<h2>First Real Test</h2>

				<p>With everything wired up, I asked Claude a simple question:</p>

				<p>"What are my memory stats?"</p>

				<p><em>(The outputs below are formatted for readability. Claude receives JSON from the MCP tools and formats it for display - the exact formatting varies by session.)</em></p>

				<pre><code>┌────────────────────┬─────────────────────────────────────┐
│       Metric       │                Value                │
├────────────────────┼─────────────────────────────────────┤
│ Total Size         │ 556 MB                              │
│ Conversation Files │ 371                                 │
│ Unique Projects    │ 45                                  │
│ Date Range         │ Dec 30, 2025 → Jan 13, 2026         │
└────────────────────┴─────────────────────────────────────┘</code></pre>

				<p>Claude called the tool, got structured JSON back, formatted it into this table. No terminal. No copy-paste. Just a question and an answer.</p>

				<p>Then I pushed harder: "Search for past conversations about DuckDB."</p>

				<pre><code>Found 77 files with DuckDB mentions.
Primary context: AionSec architecture work, historical telemetry system.</code></pre>

				<p>"Show me analytics for the AION-PAI project."</p>

				<pre><code>┌───────────────────────────────────┬───────────────────────────────────┐
│              Metric               │               Value               │
├───────────────────────────────────┼───────────────────────────────────┤
│ Total Sessions                    │ 51 (across 7 working directories) │
│ Total Size                        │ 77 MB                             │
│ Date Range                        │ Jan 3 – Jan 13, 2026 (11 days)    │
└───────────────────────────────────┴───────────────────────────────────┘

Skill Focus Areas:
- Image Generation: 7,680 mentions
- Remotion (Video): 2,102 mentions
- Keynote Gen: 1,938 mentions</code></pre>

				<p>Then the semantic search: "Find conversations about handling errors gracefully."</p>

				<pre><code>┌──────────┬─────────┬──────────────────────────────────────────────────────┐
│ Project  │  Date   │                        Topic                         │
├──────────┼─────────┼──────────────────────────────────────────────────────┤
│ numinon  │ Jan 8-9 │ Advanced error handling - retryable vs non-retryable │
│ Learning │ Jan 8   │ Go Error Handling - "Handle failures gracefully"     │
└──────────┴─────────┴──────────────────────────────────────────────────────┘</code></pre>

				<p>I hadn't searched for "retryable" or "exponential backoff." But Claude found them. Searched by meaning, not keywords. The semantic search infrastructure, now accessible directly.</p>

				<hr />

				<h2>The Unexpected Behavior</h2>

				<p>Here's what I didn't anticipate: Claude started using the memory tools proactively.</p>

				<p>I was working on something unrelated - debugging a webhook issue. Mid-conversation, Claude said: "Let me check if we've encountered this before..." and called <code>search_memory</code> without me asking.</p>

				<p>It found a relevant session from four days earlier. Same error pattern, different project. The fix applied here too.</p>

				<p>Claude had just searched its own memory to provide context I hadn't thought to look for.</p>

				<p>This is the shift. Not "a tool I can ask to search" but "an assistant that remembers."</p>

				<p>Admittedly, this doesn't always work smoothly. Sometimes Claude forgets the MCP tools exist and says "I don't have access to previous conversations" - even though it does. When this happens, I either remind it explicitly ("use your memory tools") or reference the tools by name ("call search_memory"). It's not perfect. The proactive behavior emerges inconsistently. But when it works, it's genuinely useful.</p>

				<figure class="article-image">
					<img src="/images/claude/mcp-server/IMG-004-UNEXPECTED.png" alt="Claude proactively searching its memory" />
				</figure>

				<hr />

				<h2>What I Learned</h2>

				<p><strong>Config file locations matter.</strong> <code>~/.claude/settings.json</code> and <code>~/.claude.json</code> are different files with different purposes. MCP servers go in the latter. I wasted time debugging a "broken" server that was just registered in the wrong place.</p>

				<p><strong>Claude falls back gracefully.</strong> When I tested before the MCP was properly registered, Claude still answered my questions - by shelling out to bash and grepping the JSONL files directly. It used the manual commands documented in CLAUDE.md. This validates that bash scripts would have worked. But MCP is cleaner when it works.</p>

				<p><strong>MCP servers are spawned, not long-running.</strong> Unlike my Telegram webhook (which needs ngrok and a server running constantly), Claude Code spawns MCP servers on demand. No health checks needed. No "is it running?" questions. Claude Code handles the lifecycle.</p>

				<p><strong>Pragmatic beats pure.</strong> Shelling out to Python from Go feels wrong. But it ships. If LanceDB adds Go bindings, I can refactor. Until then, the working solution beats the elegant one that doesn't exist.</p>

				<p><strong>Read-only enables autonomy.</strong> The trust boundary I mentioned earlier - it's what makes the "unexpected behavior" possible. Claude can use the tools proactively because the worst case is a useless query, not corrupted data.</p>

				<hr />

				<h2>The Bigger Picture</h2>

				<p>The memory system has three access modes:</p>

				<div class="data-table">
					<table>
						<thead>
							<tr>
								<th>Mode</th>
								<th>For</th>
								<th>How</th>
							</tr>
						</thead>
						<tbody>
							<tr>
								<td>Manual queries</td>
								<td>Deep exploration</td>
								<td>Terminal + Python scripts</td>
							</tr>
							<tr>
								<td>MCP tools</td>
								<td>Claude-assisted lookup</td>
								<td>Automatic via MCP</td>
							</tr>
							<tr>
								<td>Dashboard</td>
								<td>Visual exploration</td>
								<td>Phase 6 (coming)</td>
							</tr>
						</tbody>
					</table>
				</div>

				<p>The MCP server bridges the gap. I can still run manual queries when I want control. But for routine lookups - "have we discussed this?" - Claude handles it directly.</p>

				<p>The goal was never just "searchable conversations." It was this: an AI partner that builds on everything we've done together. Context that persists. Knowledge that compounds.</p>

				<p>With the MCP server, Claude can now search all our sessions and messages autonomously. It doesn't just respond to my questions - it can proactively check whether we've solved similar problems before.</p>

				<p>That's the difference between a tool and a partner.</p>

				<p>Next phase: a visual dashboard to explore patterns I can't see in text results. When am I most productive? How has my usage evolved? How can I reduce friction?</p>

				<p>The memory system continues.</p>

				<figure class="article-image">
					<img src="/images/claude/mcp-server/IMG-005-CLOSING.png" alt="The complete memory system architecture" />
				</figure>

				<hr />

				<h2>Quick Reference</h2>

				<p><strong>MCP server location:</strong> <code>~/repos/claude-memory-mcp/</code></p>

				<p><strong>Build the server:</strong></p>
				<pre><code>cd ~/repos/claude-memory-mcp
go build -o claude-memory-mcp ./cmd/server</code></pre>

				<p><strong>Register in Claude Code:</strong> Add to <code>~/.claude.json</code>:</p>
				<pre><code class="language-json">"mcpServers": {'{'}
  "claude-memory": {'{'}
    "type": "stdio",
    "command": "/Users/faanross/repos/claude-memory-mcp/claude-memory-mcp",
    "args": []
  {'}'}
{'}'}</code></pre>

				<p><strong>Available tools:</strong></p>
				<ul>
					<li><code>search_memory</code> - Keyword search (FTS5)</li>
					<li><code>semantic_search</code> - Meaning-based search (LanceDB)</li>
					<li><code>get_project_stats</code> - Project analytics</li>
					<li><code>recent_sessions</code> - List recent conversations</li>
					<li><code>get_session</code> - Full session content</li>
				</ul>

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

	.article-content a {
		color: var(--aion-purple-light);
		text-decoration: none;
	}

	.article-content a:hover {
		text-decoration: underline;
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

	.series-note {
		color: rgba(255, 255, 255, 0.6);
		font-style: italic;
		text-align: center;
		margin-top: 24px;
	}

	:global(.code-block) {
		position: relative;
		margin-bottom: 24px;
	}

	:global(.code-block pre) {
		margin-bottom: 0;
		padding-right: 80px;
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
