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
	<title>Your CLAUDE.md Is Probably Too Big | Faan Rossouw</title>
	<meta name="description" content="How I reduced my CLAUDE.md from 570 lines to 187 - same functionality, 67% less context usage." />
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
				<span class="date">2026-01-17</span>
				<h1>Your CLAUDE.md Is Probably Too Big</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				<figure class="article-image hero-image">
					<img src="/images/claude/claude-md-optimization/hero.png" alt="CLAUDE.md optimization - routing layer vs knowledge base" />
				</figure>

				<p>My Obsidian vault root-level CLAUDE.md is prime real estate. It loads into context at the start of every session - before Claude reads a single file or runs a single command. Whatever I put there, Claude carries with it throughout our entire interaction.</p>

				<p>That makes it powerful. It also makes it expensive. Every line in CLAUDE.md is a line that isn't available for the actual work. So then the smart play is to design rules to maximize signal while minimizing footprint.</p>

				<p>My CLAUDE.md file just hit 570 lines. Session startup checks, skill routing tables, notification configs, planning system docs, memory access commands, troubleshooting guides. Everything I'd taught Claude about my workflow, accumulated over weeks.</p>

				<p><em>(Quick check: <code>wc -l CLAUDE.md</code> tells you your line count. For a rough token estimate, divide by 4 - or paste into a tokenizer. My 570 lines were roughly 2,800 tokens.)</em></p>

				<p>It worked. But every session loaded all 570 lines into context - even when most of it was conditional.</p>

				<p>That's a leak. Here's how I fixed it.</p>

				<hr />

				<h2>The Problem With Monolithic Rules</h2>

				<p>One example from my setup - a Telegram notification system that lets me control Claude remotely from my phone:</p>

				<pre><code>{`## Telegram Remote Control

Async workflow - get notifications, reply from phone.

### Components
| Service | Port | Purpose |
|---------|------|---------|
| ngrok | 4040 | Exposes local port |
| webhook | 3001 | Receives messages |

### How It Works
1. Claude finishes → hook fires → Telegram notification
2. User replies: /cmd TOKEN123 continue with X
3. Command injected back into session

### Files
- ~/tools/Claude-Code-Remote/
- ~/tools/Claude-Code-Remote/.env
- start-telegram-services.sh

### Troubleshooting
- ngrok URL changed: restart updates webhook
- 502 errors: server not running
- No notifications: check hooks config`}</code></pre>

				<p>That's 25 lines. And here's the thing: I don't need any of this by default. The components, the workflow, the file paths, the troubleshooting - all of it is only relevant when something goes wrong with Telegram. So why load it every single session?</p>

				<p>Instead, I can just tell Claude: "If there's an issue with Telegram, go here." Claude looks up the full details only when needed. Same functionality, fraction of the context.</p>

				<hr />

				<h2>The Principle: Routing Layer, Not Knowledge Base</h2>

				<p>Your CLAUDE.md should be a switchboard, not an encyclopedia.</p>

				<p><strong>What needs to be inline:</strong></p>
				<ul>
					<li>Trigger phrases (so Claude recognizes when to act)</li>
					<li>Simple commands (one-liners that are always needed)</li>
					<li>Routing tables (which skill handles what)</li>
				</ul>

				<p><strong>What should be referenced:</strong></p>
				<ul>
					<li>Troubleshooting guides</li>
					<li>Detailed how-tos</li>
					<li>Full file structures</li>
					<li>Implementation specifics</li>
				</ul>

				<p>The distinction: triggers need instant recognition. Details only matter once the trigger fires.</p>

				<hr />

				<h2>The Pattern: Conditional References</h2>

				<p>The fix is simple: keep the trigger inline, move the details elsewhere.</p>

				<p>I use Obsidian, so wikilinks (<code>[[ref_memory_health]]</code>) make this trivial. But the pattern works with any reference system - file paths, imports, even just "see X file for details". The point is separation of concerns, not the specific syntax.</p>

				<p>Here's another example - my memory system health check. The original looked like this:</p>

				<pre><code>{`## Memory System Health Check

Run these if tools fail:

### Quick Checks
stat -f "%Sm" ~/repos/claude-memory/memory.duckdb
curl -s http://192.168.2.237:11434/api/tags
tail -3 ~/repos/claude-memory/cron.log

### If Database Stale
cd ~/repos/claude-memory && python ingest.py

### Common Issues
| Issue | Fix |
|-------|-----|
| Database lock | Fixed in ingest.py |
| FTS mismatch | Upgrade go-duckdb |
...`}</code></pre>

				<p>After refactoring:</p>

				<pre><code>{`## Memory System Health Check

If tools fail or user says \`memory health\` → [[ref_memory_health]]`}</code></pre>

				<p>One line instead of twenty-five. The trigger phrase stays inline so Claude knows when to act. The actual commands and troubleshooting tables live in a separate reference doc that only loads when needed.</p>

				<hr />

				<h2>What I Extracted</h2>

				<p>I went through my entire CLAUDE.md with one question: is this always needed, or only sometimes? Anything conditional got moved to a reference doc.</p>

				<p>Here's what that audit revealed:</p>

				<div class="table-wrapper">
					<table>
						<thead>
							<tr>
								<th>Section</th>
								<th>Before</th>
								<th>After</th>
								<th>Saved</th>
							</tr>
						</thead>
						<tbody>
							<tr>
								<td>Memory access commands</td>
								<td>40 lines</td>
								<td>5 lines</td>
								<td>35</td>
							</tr>
							<tr>
								<td>Memory health checks</td>
								<td>25 lines</td>
								<td>2 lines</td>
								<td>23</td>
							</tr>
							<tr>
								<td>Telegram details</td>
								<td>35 lines</td>
								<td>3 lines</td>
								<td>32</td>
							</tr>
							<tr>
								<td>Planning system structure</td>
								<td>40 lines</td>
								<td>5 lines</td>
								<td>35</td>
							</tr>
							<tr class="total-row">
								<td><strong>Total</strong></td>
								<td>570 lines</td>
								<td>187 lines</td>
								<td><strong>67%</strong></td>
							</tr>
						</tbody>
					</table>
				</div>

				<p>Same functionality. Exact same outcomes. Two-thirds less context usage.</p>

				<hr />

				<h2>Why Context Efficiency Matters</h2>

				<p>This isn't just housekeeping. Context windows have real performance implications.</p>

				<p><strong>The capacity myth:</strong> A 200K context window doesn't mean you should use 200K tokens. Performance degrades as you approach capacity. Anecdotally, many practitioners report noticeable quality drops around 60-70% utilization - though this varies by task type and hasn't been rigorously studied.</p>

				<p><strong>Compaction helps, but has costs:</strong> Claude Code compacts conversation history to stay within limits. But compaction loses information. The more you rely on compaction, the more context gets summarized away. Better to not need it.</p>

				<p><strong>Every token counts twice:</strong> Your CLAUDE.md loads at session start <em>and</em> persists through compaction (since it's injected fresh). A bloated rules file taxes every interaction.</p>

				<p>The goal isn't minimalism for its own sake. It's keeping the context window available for what actually matters - the current task, not static reference material that might never be needed.</p>

				<hr />

				<h2>The Bigger Picture</h2>

				<p>This is standard practice in any system where resources are constrained. You don't load your entire database into RAM. You don't bundle every possible library into your binary. You load what's needed, when it's needed.</p>

				<p>Same principle applies to global agent context design. Think of your CLAUDE.md as the routing layer, and reference docs are the knowledge base. Keep them separate to make more room for the work that matters.</p>

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

	.table-wrapper {
		margin: 24px 0;
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 15px;
	}

	th, td {
		padding: 12px 16px;
		text-align: left;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	th {
		color: var(--white);
		font-weight: 600;
		background: rgba(0, 0, 0, 0.3);
	}

	td {
		color: rgba(255, 255, 255, 0.8);
	}

	.total-row td {
		border-top: 2px solid rgba(189, 147, 249, 0.3);
		font-weight: 600;
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

		th, td {
			padding: 10px 12px;
			font-size: 14px;
		}
	}
</style>
