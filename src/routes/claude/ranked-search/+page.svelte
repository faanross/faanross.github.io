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
	<title>I Gave My Memory System Ranked Search | Faan Rossouw</title>
	<meta name="description" content="Adding full-text search with BM25 ranking to Claude's conversation history. From 47 equal results to scored relevance." />
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
				<span class="date">2026-01-15</span>
				<h1>I Gave My Memory System Ranked Search</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				<p>The database was working. Queries returned results. But there was a problem.</p>

				<p>"Authentication" returned 47 matches. All equal weight. The conversation that actually solved my problem sat somewhere in the middle, indistinguishable from casual mentions.</p>

				<p>SQL <code>LIKE</code> queries find things. They don't rank them.</p>

				<figure class="article-image">
					<img src="/images/claude/ranked-search/hero.png" alt="The chaos of unranked search results - all equal weight" />
				</figure>

				<hr />

				<h2>The Ranking Problem</h2>

				<p>Phase 1 gave me SQL queries. That's powerful - I could find every message containing "authentication" instantly. But <code>LIKE '%authentication%'</code> treats every match the same.</p>

				<p>A message that's <em>about</em> authentication gets the same weight as one that mentions it in passing. A focused discussion about token refresh logic scores identically to a random aside about auth being annoying.</p>

				<p>That's not how search should work.</p>

				<p>When I search Google, I don't get "here's 10,000 pages containing your keyword, good luck." I get ranked results. Most relevant first. The algorithm understands that some matches matter more than others.</p>

				<p>My memory system needed the same thing.</p>

				<hr />

				<h2>What I Needed</h2>

				<p>The gap between Phase 1 and useful search came down to a few capabilities:</p>

				<p><strong>Ranking</strong> - Higher scores for more relevant matches. A message focused on "authentication" should beat one that mentions it once.</p>

				<p><strong>Boolean queries</strong> - <code>duckdb AND NOT sqlite</code> to find discussions about DuckDB that aren't comparisons with SQLite.</p>

				<p><strong>Phrase matching</strong> - <code>"voice mode"</code> as an exact phrase, not just messages containing both words somewhere.</p>

				<p><strong>Stemming</strong> - "running" should match "run", "runs", "runner". Don't make me guess which form I used.</p>

				<p>Standard SQL <code>LIKE</code> can't do any of this. I needed full-text search.</p>

				<hr />

				<h2>The Architecture Choice</h2>

				<p>Two options emerged:</p>

				<p><strong>Option A: Separate SQLite database.</strong> Keep DuckDB for analytics, add SQLite with FTS5 for search. Clean separation - each database does what it's best at.</p>

				<p><strong>Option B: DuckDB's built-in FTS extension.</strong> Use DuckDB for everything. Single database, simpler stack.</p>

				<p>I spent some time researching both. SQLite's FTS5 is mature, battle-tested, widely documented. DuckDB's FTS extension is newer, less commonly used.</p>

				<p>But Option A meant maintaining two databases, syncing data between them, managing two connections. More moving parts. More ways for things to drift out of sync.</p>

				<p>I went with Option B. DuckDB FTS might be less mature, but simplicity usually wins. If it proves insufficient, I can always add SQLite later.</p>

				<hr />

				<h2>Setting It Up</h2>

				<p>The setup was surprisingly straightforward.</p>

				<pre><code>{`-- Enable the extension
INSTALL fts;
LOAD fts;

-- Create the index
PRAGMA create_fts_index(
    'messages', 'id', 'content',
    stemmer='porter',
    stopwords='english',
    ignore='(\\.|[^a-z])+',
    strip_accents=1,
    lower=1
);`}</code></pre>

				<p>That's it. One command to install, one to create the index on my messages table.</p>

				<p>The options do the heavy lifting:</p>

				<ul>
					<li><strong>Porter stemmer</strong> - Handles word variations (running → run)</li>
					<li><strong>English stopwords</strong> - Ignores "the", "is", "at"</li>
					<li><strong>Case insensitive</strong> - "DuckDB" matches "duckdb"</li>
				</ul>

				<p>The database grew from 69MB to about 100MB. That's the price of an index - space for speed.</p>

				<hr />

				<h2>The First Ranked Search</h2>

				<p>The moment I ran my first BM25 query, I knew this was the right call.</p>

				<pre><code>{`SELECT
    s.project_name,
    fts_main_messages.match_bm25(m.id, 'authentication') as score,
    SUBSTRING(m.content, 1, 100) as preview
FROM messages m
JOIN sessions s ON m.session_id = s.session_id
WHERE fts_main_messages.match_bm25(m.id, 'authentication') IS NOT NULL
ORDER BY score DESC
LIMIT 5;`}</code></pre>

				<p>Results:</p>

				<div class="data-table">
					<table>
						<thead>
							<tr>
								<th>project_name</th>
								<th>score</th>
								<th>preview</th>
							</tr>
						</thead>
						<tbody>
							<tr><td>vault</td><td>4.32</td><td>[THINKING] They have `gh` installed but not authenticated...</td></tr>
							<tr><td>vault</td><td>3.93</td><td>so I say no here Authenticate Git with your GitHub...</td></tr>
							<tr><td>vault</td><td>3.93</td><td>[THINKING] gh is already installed. Just need to authenticate...</td></tr>
							<tr><td>vault</td><td>3.89</td><td>[THINKING] The GitHub server is added but shows "Failed to connect"...</td></tr>
							<tr><td>vault</td><td>3.80</td><td>[THINKING] The user is asking if I can delete repos from GitHub...</td></tr>
						</tbody>
					</table>
				</div>

				<figure class="article-image">
					<img src="/images/claude/ranked-search/scr-bm25-auth.png" alt="Terminal showing BM25 ranked search results for authentication" />
				</figure>

				<p>Notice the scores. 4.32 vs 3.80 isn't a huge spread, but it's meaningful. The top result is <em>about</em> authentication. The bottom result just mentions it while discussing something else (deleting repos).</p>

				<p>That's BM25 working. Term frequency, document length, rarity across the corpus - all factored into a single relevance score.</p>

				<figure class="article-image">
					<img src="/images/claude/ranked-search/ranking.png" alt="Ranked search results with clear hierarchy" />
				</figure>

				<hr />

				<h2>Boolean Queries in Action</h2>

				<p>The boolean operators opened up queries I couldn't do before.</p>

				<p><strong>Find DuckDB discussions that aren't about SQLite:</strong></p>

				<pre><code>{`WHERE fts_main_messages.match_bm25(m.id, 'duckdb AND NOT sqlite') IS NOT NULL`}</code></pre>

				<p>When I'm looking for DuckDB-specific insights, I don't want comparison discussions. This filters them out.</p>

				<figure class="article-image">
					<img src="/images/claude/ranked-search/scr-boolean-andnot.png" alt="Terminal showing DuckDB AND NOT sqlite query results" />
				</figure>

				<p><strong>Find either authentication or authorization:</strong></p>

				<pre><code>{`WHERE fts_main_messages.match_bm25(m.id, 'authentication OR authorization') IS NOT NULL`}</code></pre>

				<p>Related concepts, both relevant. One query catches both.</p>

				<figure class="article-image">
					<img src="/images/claude/ranked-search/scr-boolean-or.png" alt="Terminal showing authentication OR authorization query results" />
				</figure>

				<figure class="article-image">
					<img src="/images/claude/ranked-search/boolean.png" alt="Boolean logic operations - AND, OR, NOT" />
				</figure>

				<hr />

				<h2>Phrase Search</h2>

				<p>This was the feature I didn't know I needed until I used it.</p>

				<pre><code>{`WHERE fts_main_messages.match_bm25(m.id, '"voice mode"') IS NOT NULL`}</code></pre>

				<p>The double quotes mean "exact phrase." Not messages containing "voice" somewhere and "mode" elsewhere. The phrase "voice mode" as a unit.</p>

				<p>For technical terms, feature names, error messages - phrase search is essential. <code>"connection refused"</code> finds actual connection errors. <code>connection refused</code> (without quotes) finds any message with both words, regardless of context.</p>

				<figure class="article-image">
					<img src="/images/claude/ranked-search/scr-phrase-search.png" alt="Terminal showing voice mode phrase search results" />
				</figure>

				<hr />

				<h2>Integrating With Ingestion</h2>

				<p>The index needs to stay current. New conversations happen constantly.</p>

				<p>I updated my ingestion script to rebuild the FTS index after each data load:</p>

				<pre><code>{`def create_fts_index(con):
    con.execute("INSTALL fts")
    con.execute("LOAD fts")

    # Drop existing index if it exists
    try:
        con.execute("PRAGMA drop_fts_index('messages')")
    except:
        pass

    # Rebuild
    con.execute("""
        PRAGMA create_fts_index(
            'messages', 'id', 'content',
            stemmer='porter',
            stopwords='english'
        )
    """)`}</code></pre>

				<p>Every hourly ingestion now rebuilds the index automatically. Full refresh each time - simple, reliable, no drift.</p>

				<hr />

				<h2>What I Learned</h2>

				<p><strong>Extension loading is per-session.</strong> You need <code>LOAD fts;</code> at the start of each DuckDB session before using FTS functions. Easy to forget, cryptic error when you do.</p>

				<p><strong>The index overhead is worth it.</strong> 30MB for ranked search across 50,000+ messages. That's nothing compared to the value of relevant results first.</p>

				<p><strong>BM25 just works.</strong> I didn't have to tune parameters or understand the math deeply. Out of the box, it ranks sensibly. Higher scores mean better matches.</p>

				<p><strong>Phrase search uses double quotes.</strong> <code>"voice mode"</code> not <code>'voice mode'</code>. Lost a few minutes to that one.</p>

				<hr />

				<h2>The Remaining Gap</h2>

				<p>Here's what Phase 2 still can't do:</p>

				<p>Search for "auth" and find conversations about "login."</p>

				<p>The words are different. Conceptually, they're the same topic. But FTS matches keywords, not concepts. Stemming helps with word forms (run/running), but it can't bridge semantic gaps.</p>

				<p>"How do I handle authentication errors?" won't find "debugging login failures" - even though they're discussing the same problem.</p>

				<p>That's the limitation I'm building toward fixing. Phase 3 sets up local LLM infrastructure. Phase 4 uses it for semantic search - matching by meaning, not keywords.</p>

				<p>But ranked keyword search is already a massive improvement. The 47 equal results became a scored list with the best matches at the top.</p>

				<p>The friction of finding things just dropped significantly.</p>

				<hr />

				<h2>Quick Reference</h2>

				<p><strong>Load extension (required each session):</strong></p>

				<pre><code>LOAD fts;</code></pre>

				<p><strong>Basic search template:</strong></p>

				<pre><code>{`SELECT
    s.project_name,
    DATE(m.timestamp) as date,
    fts_main_messages.match_bm25(m.id, 'YOUR SEARCH TERMS') as score,
    SUBSTRING(m.content, 1, 200) as preview
FROM messages m
JOIN sessions s ON m.session_id = s.session_id
WHERE fts_main_messages.match_bm25(m.id, 'YOUR SEARCH TERMS') IS NOT NULL
ORDER BY score DESC
LIMIT 20;`}</code></pre>

				<p><strong>Search syntax:</strong></p>

				<div class="data-table">
					<table>
						<thead>
							<tr>
								<th>Pattern</th>
								<th>Meaning</th>
							</tr>
						</thead>
						<tbody>
							<tr><td><code>word</code></td><td>Simple search</td></tr>
							<tr><td><code>word1 word2</code></td><td>Both words (implicit AND)</td></tr>
							<tr><td><code>word1 AND word2</code></td><td>Both words (explicit)</td></tr>
							<tr><td><code>word1 OR word2</code></td><td>Either word</td></tr>
							<tr><td><code>word1 AND NOT word2</code></td><td>First but not second</td></tr>
							<tr><td><code>"exact phrase"</code></td><td>Phrase match</td></tr>
						</tbody>
					</table>
				</div>

				<hr />

				<p><em>Part 2 of the Claude Memory Project. Next: setting up local LLM for embeddings.</em></p>

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
