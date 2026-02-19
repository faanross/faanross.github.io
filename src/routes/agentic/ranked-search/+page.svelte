<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="I Gave My Memory System Ranked Search"
	date="2026-01-15"
	description="Adding full-text search with BM25 ranking to Claude's conversation history. From 47 equal results to scored relevance."
>
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

	<p><strong>Option B: DuckDB's FTS extension.</strong> Use DuckDB for everything. Single database, simpler stack. (It's a loadable extension, not built-in - you need <code>INSTALL fts; LOAD fts;</code> - but it integrates seamlessly once loaded.)</p>

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

	<p>The parameters explained:</p>

	<ul>
		<li><strong><code>'messages', 'id', 'content'</code></strong> - Table name, primary key column, column(s) to index</li>
		<li><strong><code>stemmer='porter'</code></strong> - Porter stemming algorithm. Handles word variations: "running" → "run", "authentication" → "authent"</li>
		<li><strong><code>stopwords='english'</code></strong> - Ignores common words like "the", "is", "at" that add noise without meaning</li>
		<li><strong><code>ignore='(\\.|[^a-z])+'</code></strong> - Regex to skip during indexing. This skips punctuation and non-letter characters</li>
		<li><strong><code>strip_accents=1</code></strong> - Treats "café" and "cafe" as equivalent</li>
		<li><strong><code>lower=1</code></strong> - Case-insensitive: "DuckDB" matches "duckdb"</li>
	</ul>

	<p>The index creates a naming convention: <code>fts_main_&lt;tablename&gt;</code>. So for my <code>messages</code> table, the index is accessed via <code>fts_main_messages</code>.</p>

	<p>The database grew with the index - you can check the size difference with <code>du -sh</code> before and after. The overhead is the price of ranked search.</p>

	<hr />

	<h2>The First Ranked Search</h2>

	<p>The moment I ran my first BM25 query, I knew this was the right call.</p>

	<p><strong>What is BM25?</strong> It stands for "Best Match 25" - a ranking algorithm that scores documents based on how relevant they are to a search query. It considers:</p>
	<ul>
		<li><strong>Term frequency</strong> - How often does the search term appear in this message?</li>
		<li><strong>Document length</strong> - Longer documents get slightly penalized (a term appearing once in 10 words is more significant than once in 1000)</li>
		<li><strong>Inverse document frequency</strong> - Rare terms across the corpus score higher than common ones</li>
	</ul>

	<p>The query uses DuckDB's <code>match_bm25</code> function:</p>

	<pre><code>{`SELECT
    s.project_name,
    fts_main_messages.match_bm25(m.id, 'authentication') as score,
    SUBSTRING(m.content, 1, 100) as preview
FROM messages m
JOIN sessions s ON m.session_id = s.session_id
WHERE fts_main_messages.match_bm25(m.id, 'authentication') IS NOT NULL
ORDER BY score DESC
LIMIT 5;`}</code></pre>

	<!-- TODO: Replace with actual screenshot of this query result -->
	<figure class="article-image placeholder">
		<img src="/images/claude/ranked-search/scr-bm25-auth.png" alt="Terminal showing BM25 ranked search results for authentication" />
		<figcaption class="placeholder-note">📸 TODO: Actual DuckDB CLI screenshot of authentication BM25 search</figcaption>
	</figure>

	<p>The scores differentiate results. Higher scores mean the message is more <em>about</em> the search term, not just mentioning it in passing. A focused discussion about authentication outranks a casual aside.</p>

	<p>That's BM25 working. Term frequency, document length, rarity across the corpus - all factored into a single relevance score. No parameter tuning needed; it just works out of the box.</p>

	<figure class="article-image">
		<img src="/images/claude/ranked-search/ranking.png" alt="Ranked search results with clear hierarchy" />
	</figure>

	<hr />

	<h2>Boolean Queries in Action</h2>

	<p>The boolean operators opened up queries I couldn't do before.</p>

	<p><strong>Find DuckDB discussions that aren't about SQLite:</strong></p>

	<pre><code>{`WHERE fts_main_messages.match_bm25(m.id, 'duckdb AND NOT sqlite') IS NOT NULL`}</code></pre>

	<p>When I'm looking for DuckDB-specific insights, I don't want comparison discussions. This filters them out.</p>

	<!-- TODO: Replace with actual screenshot of AND NOT query -->
	<figure class="article-image placeholder">
		<img src="/images/claude/ranked-search/scr-boolean-andnot.png" alt="Terminal showing DuckDB AND NOT sqlite query results" />
		<figcaption class="placeholder-note">📸 TODO: Actual DuckDB CLI screenshot of AND NOT query</figcaption>
	</figure>

	<p><strong>Find either authentication or authorization:</strong></p>

	<pre><code>{`WHERE fts_main_messages.match_bm25(m.id, 'authentication OR authorization') IS NOT NULL`}</code></pre>

	<p>Related concepts, both relevant. One query catches both.</p>

	<!-- TODO: Replace with actual screenshot of OR query -->
	<figure class="article-image placeholder">
		<img src="/images/claude/ranked-search/scr-boolean-or.png" alt="Terminal showing authentication OR authorization query results" />
		<figcaption class="placeholder-note">📸 TODO: Actual DuckDB CLI screenshot of OR query</figcaption>
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

	<!-- TODO: Replace with actual screenshot of phrase search -->
	<figure class="article-image placeholder">
		<img src="/images/claude/ranked-search/scr-phrase-search.png" alt="Terminal showing phrase search results" />
		<figcaption class="placeholder-note">📸 TODO: Actual DuckDB CLI screenshot of phrase search</figcaption>
	</figure>

	<hr />

	<h2>Integrating With Ingestion</h2>

	<p>The index needs to stay current. New conversations happen constantly.</p>

	<p>I updated my ingestion script to rebuild the FTS index after each data load:</p>

	<pre><code>{`def create_fts_index(con):
    con.execute("INSTALL fts")
    con.execute("LOAD fts")

    # Drop existing index if it exists
    # The try/except handles the case where the index doesn't exist yet
    try:
        con.execute("PRAGMA drop_fts_index('messages')")
    except duckdb.CatalogException:
        pass  # Index doesn't exist yet, that's fine

    # Rebuild from scratch
    con.execute("""
        PRAGMA create_fts_index(
            'messages', 'id', 'content',
            stemmer='porter',
            stopwords='english'
        )
    """)`}</code></pre>

	<p>Every hourly ingestion now rebuilds the index automatically. Full refresh each time.</p>

	<p><strong>Why full rebuild instead of incremental?</strong> Simplicity. DuckDB's FTS doesn't support incremental index updates - you either rebuild the whole thing or manage a more complex sync. For my data size (tens of thousands of messages), a full rebuild takes seconds. If this grew to millions of rows, I'd need a different approach - possibly a separate search service like Meilisearch or Typesense. For now, simple wins.</p>

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
</style>
