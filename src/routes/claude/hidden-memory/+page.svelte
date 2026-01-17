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
	<title>I Discovered Claude Code Has a Hidden Memory | Faan Rossouw</title>
	<meta name="description" content="Claude Code stores complete conversation transcripts locally. Here's how to unlock persistent memory across sessions with one addition to your CLAUDE.md." />
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
				<span class="date">2025-01-11</span>
				<h1>I Discovered Claude Code Has a Hidden Memory</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				<p>I wasn't looking for this. I was trying to build something else entirely.</p>

				<p>I'd been setting up a Stop Hook - a script that fires when a Claude session ends. The idea was to automatically log session summaries to my daily notes. Keep a record of what Claude and I worked on each day.</p>

				<p>I couldn't quite get it to do what I intended.</p>

				<p>But in debugging why it wasn't working, I stumbled onto something better.</p>

				<figure class="article-image">
					<img src="/images/claude/hidden-memory/001.png" alt="Hidden treasure chest revealing glowing data" />
				</figure>

				<hr />

				<h2>The accidental discovery</h2>

				<p>I was digging through Claude Code's internals, trying to figure out what data the hooks actually receive, when I noticed a folder I'd never paid attention to:</p>

				<pre><code>~/.claude/projects/</code></pre>

				<p>I opened it expecting config files. Instead, I found something else entirely.</p>

				<p>JSONL files. Hundreds of them. Organized by working directory. Each one containing... everything.</p>

				<p>Every message I'd sent. Every response Claude gave. Every tool call, every file read, every edit. Timestamps. Token counts. The complete record of every conversation I'd ever had with Claude Code.</p>

				<p>I checked the dates. Files going back to the very first conversation I had with Claude Code. Nothing deleted. All of it still there.</p>

				<p>Permanent. Local. Mine.</p>

				<hr />

				<h2>What's actually in there</h2>

				<p>Each session is a JSONL file - one JSON object per line. Here's what gets stored:</p>

				<p><strong>For every message you send:</strong></p>

				<pre><code>{`{
  "type": "user",
  "message": { "role": "user", "content": "your message here" },
  "timestamp": "2026-01-07T22:26:33.425Z",
  "sessionId": "6a704ab6-...",
  "cwd": "/path/to/your/project"
}`}</code></pre>

				<p><strong>For every Claude response:</strong></p>

				<pre><code>{`{
  "type": "assistant",
  "message": { "role": "assistant", "content": [...] },
  "timestamp": "2026-01-07T22:26:37.102Z",
  "usage": { "input_tokens": 9500, "output_tokens": 1200 }
}`}</code></pre>

				<p>Voice or typed - doesn't matter. Once transcribed, it's all text, all stored.</p>

				<p>The files are organized by the directory you're working in. So all your sessions in <code>/Users/you/project-a/</code> live in one folder, sessions in <code>/Users/you/project-b/</code> in another.</p>

				<hr />

				<h2>Why this matters</h2>

				<p>I'd been wanting something like this since I started using AI assistants. A way to go back and find "that thing Claude explained last week." A searchable record of decisions and solutions.</p>

				<p>But I assumed I'd have to build it. Export conversations manually. Set up logging. Create some elaborate capture system.</p>

				<p>Turns out it already existed. I just didn't know where to look.</p>

				<p>And it's better than what I would have built:</p>

				<ul>
					<li><strong>Complete</strong> - both sides of every conversation, not just my input</li>
					<li><strong>Automatic</strong> - no export step, no manual logging</li>
					<li><strong>Permanent</strong> - local disk storage, no cloud retention policy eating your history after 30 days</li>
					<li><strong>Parseable</strong> - standard JSONL format, easy to query with basic tools</li>
				</ul>

				<hr />

				<h2>The comparison that made me realize this is gold</h2>

				<p>I'd seen <a href="https://www.linkedin.com/in/artemxtech" target="_blank" rel="noopener noreferrer">Artem Zhutov</a> posting about analyzing his Claude conversations. He uses Wispr Flow - a voice dictation tool that captures everything he speaks into any app. 956K words across all applications. 147K words to Claude Desktop alone.</p>

				<p>Impressive stats. But here's what I realized:</p>

				<figure class="article-image">
					<img src="/images/claude/hidden-memory/003.png" alt="Comparison between partial and complete conversation capture" />
				</figure>

				<div class="comparison-table">
					<table>
						<thead>
							<tr>
								<th>Aspect</th>
								<th>Wispr Flow</th>
								<th>Claude Code Built-in</th>
							</tr>
						</thead>
						<tbody>
							<tr>
								<td>What's captured</td>
								<td>Your input only</td>
								<td>Full conversations (both sides)</td>
							</tr>
							<tr>
								<td>Claude's responses</td>
								<td>No</td>
								<td>Yes</td>
							</tr>
							<tr>
								<td>Tool calls</td>
								<td>No</td>
								<td>Yes</td>
							</tr>
							<tr>
								<td>Token usage</td>
								<td>No</td>
								<td>Yes</td>
							</tr>
							<tr>
								<td>Storage</td>
								<td>Cloud, 30-day default</td>
								<td>Local, permanent</td>
							</tr>
							<tr>
								<td>Requires setup</td>
								<td>Yes (subscription)</td>
								<td>Already there</td>
							</tr>
						</tbody>
					</table>
				</div>

				<p>Wispr Flow captures what you <em>say</em>. Claude Code captures everything that <em>happened</em>.</p>

				<p>I already had richer data than I thought possible. I just needed to know where to find it.</p>

				<hr />

				<h2>How to access your conversation history</h2>

				<p>Here's the practical guide. Everything you need to query your past sessions.</p>

				<h3>Find your history</h3>

				<pre><code>ls -la ~/.claude/projects/</code></pre>

				<p>You'll see folders named after your working directories, with dashes replacing slashes:</p>

				<pre><code>{`-Users-yourname-project-a/
-Users-yourname-project-b/
-Users-yourname-Documents-work/`}</code></pre>

				<h3>List sessions for a specific project</h3>

				<pre><code>ls -la ~/.claude/projects/-Users-yourname-project-a/</code></pre>

				<p>Each <code>.jsonl</code> file is a session. Timestamps in the file names.</p>

				<h3>Find sessions by date</h3>

				<p>Looking for what you worked on Tuesday?</p>

				<pre><code>{`find ~/.claude/projects -name "*.jsonl" -type f ! -path "*/subagents/*" -newermt "2026-01-07" ! -newermt "2026-01-08" -exec ls -la {} \\;`}</code></pre>

				<h3>Search across all conversations</h3>

				<p>That solution you can't quite remember? Grep it:</p>

				<pre><code>{`grep -r "MITRE" ~/.claude/projects/ --include="*.jsonl" | head -20`}</code></pre>

				<h3>Parse a session into readable format</h3>

				<pre><code>{`cat session-file.jsonl | jq -s '[.[] | select(.type == "user" or .type == "assistant")] | .[] | {type, content: .message.content, time: .timestamp}'`}</code></pre>

				<h3>Quick stats</h3>

				<p>How many sessions total?</p>

				<pre><code>{`find ~/.claude/projects -name "*.jsonl" -type f ! -path "*/subagents/*" | wc -l`}</code></pre>

				<p>Total size of your history?</p>

				<pre><code>du -sh ~/.claude/projects/</code></pre>

				<hr />

				<h2>Making Claude aware of its own memory</h2>

				<p>Here's the real unlock: you can tell Claude Code to access this history.</p>

				<p>I added a section to my <code>CLAUDE.md</code> file (the project instructions Claude reads on startup):</p>

				<pre><code>{`## Conversation History Access

Claude Code stores complete conversation transcripts at ~/.claude/projects/

When user asks about:
- "What did we work on last week?"
- "Remember when we discussed X?"
- "What was that solution for Y?"

Check the JSONL files in that directory. Files are organized by working directory.
Use grep for keyword searches, jq for parsing specific sessions.`}</code></pre>

				<p>Now when I ask "What did we discuss about authentication last week?" - Claude knows where to look.</p>

				<p>Want to test it? After adding this to your CLAUDE.md, start a new session and ask: "Claude, what was the very first conversation we ever had about?" It's a simple way to verify the memory is working.</p>

				<p>The snippet above is the minimum to get started. For a more complete version with trigger phrases, query examples, and use cases, see the <a href="#full-snippet">full CLAUDE.md snippet</a> at the bottom of this article.</p>

				<p>It's not perfect retrieval—at least not yet. Claude has to search and parse like any other file operation. But it works, and I have <a href="#building-next">plans to make it better</a>. The memory exists and is accessible.</p>

				<figure class="article-image">
					<img src="/images/claude/hidden-memory/005.png" alt="Data flow showing conversation history being accessed" />
				</figure>

				<hr />

				<h2>What you can do with this</h2>

				<p>Beyond simple lookups, this data enables:</p>

				<p><strong>Personal analytics</strong></p>
				<ul>
					<li>Which projects get most of your attention?</li>
					<li>What times are you most productive with Claude?</li>
					<li>How much are you spending on tokens?</li>
				</ul>

				<p><strong>Pattern recognition</strong></p>
				<ul>
					<li>Topics that keep coming up</li>
					<li>Questions you ask repeatedly (maybe worth documenting)</li>
					<li>How your usage has evolved over time</li>
				</ul>

				<p><strong>Knowledge extraction</strong></p>
				<ul>
					<li>Export important conversations as permanent notes</li>
					<li>Build a searchable knowledge base from your sessions</li>
					<li>Create summaries of project work</li>
				</ul>

				<p><strong>Continuity</strong></p>
				<ul>
					<li>Pick up exactly where you left off, even weeks later</li>
					<li>Reference specific past decisions</li>
					<li>Find that code snippet Claude wrote that you didn't save</li>
				</ul>

				<hr />

				<h2>The irony</h2>

				<p>I spent way too long trying to build session logging via Stop Hooks. It failed because Claude Code doesn't pass conversation content to hooks.</p>

				<p>Then I discovered that Claude Code was already logging everything, automatically, in a better format than I would have designed, with more data than I would have captured.</p>

				<p>Sometimes the feature you want already exists.</p>

				<figure class="article-image">
					<img src="/images/claude/hidden-memory/006.png" alt="Analytics dashboard visualization" />
				</figure>

				<hr />

				<h2 id="building-next">What I'm building next</h2>

				<p>The raw data is there, but grep and jq only get you so far. I'm building two things:</p>

				<h3>Smarter retrieval</h3>

				<p>Right now Claude searches my history with basic text matching. I'm moving the data into <strong>DuckDB</strong> - a fast analytical database that can slice through hundreds of megabytes in milliseconds. On top of that, I'm adding <strong>semantic search</strong> via vector embeddings. The goal: find past conversations by <em>meaning</em>, not just keywords. "Find when I solved something like this before" - even if I used completely different words.</p>

				<h3>Visual dashboard</h3>

				<p>A Svelte-based interface for exploring my conversation history, using <a href="https://layercake.graphics/" target="_blank" rel="noopener noreferrer">Layercake</a> for visualizations. Activity heatmaps, project breakdowns, topic patterns over time. The kind of insights that are invisible when everything lives in flat files. Think: "Which projects get most of my attention?" or "When am I most productive with Claude?"</p>

				<figure class="article-image">
					<img src="/images/claude/hidden-memory/007.png" alt="Memory to insights flow: raw JSONL files to DuckDB with semantic search to visual dashboard" />
				</figure>

				<p>The data is a goldmine. Now I'm building the tools to actually mine it.</p>

				<hr />

				<h2>The bigger picture</h2>

				<p>My mantra: reduce friction to the minimum required to fully manifest an idea. Every unnecessary step between thought and creation is a leak in the system.</p>

				<p>Starting every session from scratch is friction. Re-explaining context. Repeating decisions. Losing the thread of what you solved last week. That's cognitive overhead that doesn't serve the work.</p>

				<p>This discovery - that the memory already exists, just waiting to be accessed - removes that friction. Claude can now remember. Not perfectly, not magically. But the data is there. The continuity is possible.</p>

				<p>If you're using Claude Code, you have this too. Check <code>~/.claude/projects/</code>. Your history is waiting.</p>

				<hr />

				<h2 id="full-snippet">Bonus: Full CLAUDE.md snippet</h2>

				<p>Here's the complete conversation history section from my CLAUDE.md. Copy, paste, and adapt for your own setup:</p>

				<pre><code>{`## Conversation History Access (Long-Term Memory)

Claude Code automatically stores **complete conversation transcripts** locally. This gives you persistent memory across sessions.

### Location

\`\`\`
~/.claude/projects/
\`\`\`

Files are organized by working directory path. Each session creates a JSONL file containing:
- All user messages (typed and voice-transcribed)
- All Claude responses
- Tool calls and results
- Timestamps
- Token usage

### Data Retention

**PERMANENT** - stored locally on disk, no cloud retention policy. Files stay until manually deleted.

### When to Access

If user asks about:
- "What did we work on last week?"
- "What did we discuss on [date]?"
- "Remember when we talked about X?"
- "What was that solution for Y?"
- "Show me our conversation history"
- "Search our past sessions for [topic]"

### How to Access

1. **List sessions for a project:**
   \`\`\`bash
   ls -la ~/.claude/projects/-Users-yourname-project/
   \`\`\`

2. **Find sessions by date:**
   \`\`\`bash
   find ~/.claude/projects -name "*.jsonl" -type f ! -path "*/subagents/*" -newermt "2026-01-10" -exec ls -la {} \\;
   \`\`\`

3. **Search for keywords across all sessions:**
   \`\`\`bash
   grep -r "keyword" ~/.claude/projects/ --include="*.jsonl"
   \`\`\`

4. **Parse a specific session:**
   \`\`\`bash
   cat [session-file].jsonl | jq -s '[.[] | select(.type == "user" or .type == "assistant")]'
   \`\`\`

### JSONL Structure

Each line is a JSON object with:
\`\`\`json
{
  "type": "user" | "assistant",
  "message": { "role": "...", "content": "..." },
  "timestamp": "2026-01-07T22:26:33.425Z",
  "sessionId": "6a704ab6-...",
  "cwd": "/path/to/working/directory"
}
\`\`\`

### Use Cases

- Recall specific decisions or explanations from past sessions
- Find code snippets or solutions discussed previously
- Track patterns in what topics you work on
- Search for that "thing we talked about" without remembering exactly when`}</code></pre>

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

	.comparison-table {
		margin: 24px 0;
		overflow-x: auto;
	}

	.comparison-table table {
		width: 100%;
		border-collapse: collapse;
		font-size: 15px;
	}

	.comparison-table th,
	.comparison-table td {
		padding: 12px 16px;
		text-align: left;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.comparison-table th {
		color: var(--white);
		font-weight: 600;
		background: rgba(189, 147, 249, 0.1);
	}

	.comparison-table td {
		color: rgba(255, 255, 255, 0.85);
	}

	.comparison-table tr:hover td {
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

		.comparison-table {
			font-size: 14px;
		}

		.comparison-table th,
		.comparison-table td {
			padding: 10px 12px;
		}
	}
</style>
