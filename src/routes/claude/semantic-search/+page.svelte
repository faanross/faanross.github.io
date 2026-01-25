<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="I Built Search That Finds What I Mean, Not What I Type"
	date="2026-01-19"
	description="Adding semantic search to my Claude memory system with LanceDB. Find conversations by concept, not just keywords."
>
	<p>One week into this project, my keyword search was working nicely. I could find exact error messages, specific code patterns, anything I remembered the words for. This saved me a number of times from having to repeat certain info, or when I needed a specific solution we stumbled on in the past that I'd forgotten.</p>

	<p>Then I tried something different: a friction analysis. I wanted to surface interactions where Claude and I hit walls - moments of frustration, dead ends, workflows that felt painful. The goal was to identify patterns that might reveal opportunities for new skills or better workflows.</p>

	<p>But how do you search for frustration?</p>

	<p>I tried "frustrated." "Problem." "Didn't work." "Failed." Heck, I even dropped a f-bomb in there (this did produce a few results). I got some results, but couldn't shake the feeling that many of my least favourite interactions were still not surfacing. And like an alchemist searching for lead to transmute, I was ready to layer on the next component of my memory system - semantic search.</p>

	<p>Keyword search requires you to guess which exact words past-you happened to use. Semantic search allows you to find info based on sentiment - you can search for a concept, or an analogous term, and still find the result you want.</p>

	<figure class="article-image">
		<img src="/images/claude/semantic-search/IMG-001-HERO.png" alt="Semantic search concept - finding meaning, not just words" />
	</figure>

	<hr />

	<h2>How Semantic Search Works</h2>

	<p>Semantic search is powered by embedding models - neural networks trained on massive amounts of text. These models learn that certain words and phrases tend to appear in similar contexts, and from that they develop an understanding of meaning that goes beyond literal string matching.</p>

	<p>When you pass text through an embedding model, it outputs a vector - a list of hundreds of numbers that represent the "meaning" of that text in high-dimensional space. The key insight: texts with similar meanings produce similar vectors, even when they use completely different words.</p>

	<p><em>(Technical note: "similar vectors" means vectors that are close together when measured by a distance metric. LanceDB uses L2 (Euclidean) distance by default - literally the straight-line distance between two points in 768-dimensional space. Lower distance = more similar meaning.)</em></p>

	<p>This is what makes it possible to search for "frustration" and find "this is getting tedious." The embedding model understands these express related concepts.</p>

	<pre><code>"This is getting tedious"           →  [0.31, 0.42, -3.98, ..., 0.57]
"Why isn't this working?"           →  [0.29, 0.45, -3.85, ..., 0.52]
"Everything runs smoothly now"      →  [-0.65, 0.12, 0.33, ..., -0.21]</code></pre>

	<p>The first two vectors are close together in that 768-dimensional space - both express frustration even though they share almost no words. The third is far away. Different emotional territory.</p>

	<hr />

	<h2>Choosing LanceDB</h2>

	<p>I needed a vector database. Some options include Pinecone, Weaviate, Milvus - all cloud-hosted or requiring infrastructure. But I'd built this entire project on a specific philosophy: embedded, file-based, no servers. DuckDB for analytics, FTS5 for keyword search, Ollama on my Mac Mini for embeddings. Everything runs locally.</p>

	<p>LanceDB fit the pattern:</p>
	<ul>
		<li>File-based storage (just a directory on disk, using the Lance columnar format)</li>
		<li>Embedded (runs in your process, no server to manage)</li>
		<li>Python SDK (same as my ingestion scripts)</li>
		<li>Optimized for vector similarity search with metadata filtering</li>
	</ul>

	<p>Same philosophy as DuckDB. Point it at a directory, it works. The Lance format stores vectors and metadata together efficiently - no need for a separate metadata store.</p>

	<p>I didn't evaluate alternatives deeply. The architecture matched what I was already building. Sometimes that's reason enough.</p>

	<figure class="article-image">
		<img src="/images/claude/semantic-search/IMG-002-LANCEDB.png" alt="LanceDB as embedded vector database" />
	</figure>

	<hr />

	<h2>The Embedding Process</h2>

	<p>The plan was simple: take every message from my DuckDB database, send it to the Mac Mini's embedding API, store the vector in LanceDB.</p>

	<p><em>(A note on nomic-embed-text: this model technically expects prefixes for optimal results - <code>"search_document: "</code> for indexing and <code>"search_query: "</code> for queries. I didn't use them initially and got decent results anyway. Adding them later improved relevance slightly. If you're implementing this, use the prefixes from the start.)</em></p>

	<p>52,279 messages.</p>

	<p>I wrote the first version of <code>embed.py</code> in about twenty minutes. Loop through messages, call the API, insert into LanceDB. Clean, obvious.</p>

	<p>Then I ran it.</p>

	<p>Nothing happened for two minutes. No output, no progress, no indication anything was working. Was it stuck? Was the API failing? Was I rate-limited?</p>

	<p>I added a print statement every 100 messages just to see movement. That helped, but watching "100... 200... 300..." tick up slowly toward 52,000 wasn't much better. I killed the script and added <code>tqdm</code>.</p>

	<pre><code class="language-python">{`from tqdm import tqdm

errors = []
for msg in tqdm(messages, desc="Embedding messages"):
    try:
        vector = get_embedding(msg["content"])
        # ... store in LanceDB
    except Exception as e:
        errors.append({"id": msg["id"], "error": str(e)})
        continue  # Skip this message, keep going

print(f"Completed with {len(errors)} errors")`}</code></pre>

	<pre><code>Embedding messages: 100%|████████████████████| 52279/52279 [20:14&lt;00:00, 43.1msg/s]</code></pre>

	<p>Something as simple as progress bar can dramatically improve an experience. Same 20 minutes, but now I could see it was working, estimate when it would finish, walk away and come back.</p>

	<p>At 43 messages per second, the entire history took just over 20 minutes to embed. 67 errors out of 52,279 - about 0.1%. The error handling logs failures and continues - I'd rather have 99.9% of my messages searchable than crash on edge cases. Most errors were unusual content (binary data that somehow ended up in messages, extremely long tool outputs). I logged them for later investigation and moved on.</p>

	<hr />

	<h2>The Truncation Problem</h2>

	<p>Some Claude responses are enormous. We're talking code blocks, explanations, multi-file diffs - thousands of characters. The embedding model has a context limit (nomic-embed-text supports 8192 <em>tokens</em>, which is roughly 32,000 characters since tokens average ~4 characters). And even within that limit, does embedding 30,000 characters produce a meaningfully better vector than embedding the first 8,000?</p>

	<p>I didn't know the answer, so I made practical choices:</p>
	<ul>
		<li><strong>Embedding input:</strong> Truncate to 8,000 characters (well within token limit, captures the gist)</li>
		<li><strong>Stored content:</strong> Truncate to 2,000 characters (for display in search results)</li>
	</ul>

	<p>This means if someone buries the most relevant content at the end of a long message, I'll miss it. That's a real limitation. I decided to accept it for now rather than over-engineer chunking strategies before seeing if the basic approach worked.</p>

	<figure class="article-image">
		<img src="/images/claude/semantic-search/IMG-003-TRUNCATION.png" alt="Truncation trade-offs in embedding" />
	</figure>

	<hr />

	<h2>First Search</h2>

	<p>I ran my first semantic search.</p>

	<pre><code>python search.py "retry logic when things fail" --limit 5</code></pre>

	<p>Not too sure what I was expecting tbh, but this is what I got:</p>

	<p><em>(Output below is formatted for readability - actual CLI output is similar but less pretty-printed)</em></p>

	<pre><code>Searching: "retry logic when things fail"

Found 5 results:

─── Result 1 ───
Project: numinon | Type: assistant | Date: 2026-01-08
Distance: 208.88 (lower = more similar)
Content: Advanced Error Handling is about implementing smarter retry
behavior in the agent communication layer. Looking at the description:
- Retryable errors: timeouts, 503s, network errors → retry with backoff
- Non-retryable errors: 401, 404 → don't retry, report error...

─── Result 2 ───
Project: numinon | Type: assistant | Date: 2026-01-08
Distance: 220.88 (lower = more similar)
Content: Let me explore the current agent communication code to
understand the retry logic needs.

─── Result 3 ───
Project: numinon | Type: assistant | Date: 2026-01-08
Distance: 223.55 (lower = more similar)
Content: There's some basic retry behavior in the runloops. Let me
explore the agent communication logic more thoroughly to understand
the current error handling patterns.</code></pre>

	<p>The distance scores (208.88, 220.88, 223.55) are L2 distances in 768-dimensional space. Lower means closer, meaning more semantically similar. The absolute numbers aren't intuitive - what matters is the relative ordering and that smaller is better.</p>

	<p>Cool. I hadn't searched for "backoff" or "retryable errors" or "503s." But weeks ago, when I was implementing error handling, those were the conversations we had. The embedding model understood that "retry logic when things fail" is conceptually related to "implementing smarter retry behavior" even though the phrasing is completely different.</p>

	<p>And so I tried another:</p>

	<pre><code>python search.py "setting up automated testing" --limit 5</code></pre>

	<pre><code>─── Result 1 ───
Project: AIONSEC | Type: assistant | Date: 2026-01-06
Content: The user wants to set up remote testing capability so I can:
1. Run server on their Mac (current system)
2. Deploy agent to Windows victim via SCP
3. Run agent remotely from this system...

─── Result 2 ───
Project: numinon | Type: assistant | Date: 2026-01-07
Content: Good, now I have all the information I need to write the tests.
I'll create a test file for the agentstatemanager package.

─── Result 3 ───
Project: numinon | Type: assistant | Date: 2026-01-09
Content: The testing strategy mentions unit tests with mocking,
integration tests, and manual tests...</code></pre>

	<p>None of these used the phrase "setting up automated testing." But they're all about exactly that - test infrastructure, test files, testing strategies.</p>

	<p>Semantic search actually found things I couldn't find with keywords.</p>

	<hr />

	<h2>The Data Model</h2>

	<p>Each row in LanceDB stores everything needed for a useful search result:</p>

	<pre><code class="language-python">{'{'}
    "id": "message-uuid",
    "session_id": "session-uuid",
    "content": "original text (truncated for display)",
    "vector": [0.31, 0.42, ...],  # 768 dimensions
    "timestamp": "2026-01-13T...",
    "project_name": "vault"
{'}'}</code></pre>

	<p>I store the content and metadata alongside the vector so search results are immediately useful. No second query to DuckDB needed. When you search, you get back everything you need to understand what the result is and where it came from.</p>

	<hr />

	<h2>Keyword Search Still Matters</h2>

	<p>After building semantic search, I went back to keyword search to compare. They're genuinely different tools.</p>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Search Type</th>
					<th>Best For</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td><strong>Keyword (FTS5)</strong></td>
					<td>Exact terms, error messages, specific code, boolean logic</td>
				</tr>
				<tr>
					<td><strong>Semantic</strong></td>
					<td>Concepts, "find similar," fuzzy recall, related topics</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>"Find all mentions of <code>ECONNREFUSED</code>" - that's keyword search. I know the exact string.</p>

	<p>"Find conversations about network failures" - that's semantic. I know the concept but not the words.</p>

	<p>I didn't replace one with the other. Both live in the system. Different tools for different cognitive modes.</p>

	<hr />

	<h2>What's Not Automated Yet</h2>

	<p>The embedding runs manually. When I want to update the vector index, I run <code>python embed.py</code> and it recreates the entire LanceDB table from scratch.</p>

	<p>Eventually this should be:</p>
	<ul>
		<li>Triggered automatically when new conversations are ingested</li>
		<li>Incremental (only embed new messages, not re-embed everything)</li>
		<li>Part of the hourly cron job pipeline</li>
	</ul>

	<p>But for proving the concept, manual works. The 20-minute full rebuild is a one-time cost; incremental updates will be seconds. I'll automate it at a later planned consolidation phase.</p>

	<figure class="article-image">
		<img src="/images/claude/semantic-search/IMG-004-AUTOMATION.png" alt="Manual vs automated embedding process" />
	</figure>

	<hr />

	<h2>What I Learned</h2>

	<p><strong>Semantic search is genuinely different.</strong> I was skeptical before implementing it. Finding conceptually related content with completely different vocabulary - that's not possible with keywords. It's a qualitatively different capability.</p>

	<p><strong>Progress feedback matters.</strong> Twenty minutes is fine. Twenty minutes with no indication anything is happening is anxiety. Adding <code>tqdm</code> took thirty seconds and completely changed the experience. This applies to everything, not just scripts.</p>

	<p><strong>Store metadata alongside vectors.</strong> LanceDB returns full rows. Include everything needed for display so you don't have to make a second query. The same principle applies to any search system: think about what you need to do with results, not just how to find them.</p>

	<p><strong>Truncation is a trade-off, not a failure.</strong> Long messages get chopped. Some information is lost. I could build elaborate chunking strategies, but the simple approach works well enough for now. Ship the simple thing, improve when you hit real problems.</p>

	<p><strong>Two search types &gt; one.</strong> I thought semantic would replace keyword. It doesn't. They're complementary. Different mental modes (hunting vs. browsing vs. exact lookup) want different tools.</p>

	<hr />

	<h2>The Bigger Picture</h2>

	<p>This phase completes the search layer:</p>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Layer</th>
					<th>What It Does</th>
					<th>Limitation</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td><strong>Phase 1 (DuckDB)</strong></td>
					<td>SQL queries, analytics</td>
					<td>Exact matches only</td>
				</tr>
				<tr>
					<td><strong>Phase 2 (FTS5)</strong></td>
					<td>Ranked keyword search</td>
					<td>Must know the words</td>
				</tr>
				<tr>
					<td><strong>Phase 4 (LanceDB)</strong></td>
					<td>Semantic search</td>
					<td>Needs vectors pre-computed</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>Together, they cover the spectrum from precise lookup to fuzzy recall.</p>

	<p>But the search still requires <em>me</em> to run queries manually. Open a terminal, write the command, read the output. That's friction. The next phase gives Claude direct access to these tools through an MCP server. Instead of me searching manually, Claude searches its own memory autonomously.</p>

	<p>That's when the system starts to feel less like a tool and more like what I've been building toward: an AI partner that actually remembers.</p>

	<figure class="article-image">
		<img src="/images/claude/semantic-search/IMG-005-BIGGER-PICTURE.png" alt="The complete search layer architecture" />
	</figure>

	<hr />

	<p class="series-note"><em>Part 4 of the Claude Memory Project. Next: MCP server so Claude can search its own memory.</em></p>
</ArticleLayout>

<style>
	.table-wrapper {
		overflow-x: auto;
		margin: 1.5rem 0;
	}

	.table-wrapper table {
		width: 100%;
		border-collapse: collapse;
		font-size: 15px;
	}

	.table-wrapper th,
	.table-wrapper td {
		padding: 12px 16px;
		text-align: left;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.table-wrapper th {
		color: var(--white);
		font-weight: 600;
		background: rgba(189, 147, 249, 0.1);
	}

	.table-wrapper td {
		color: rgba(255, 255, 255, 0.85);
	}

	.table-wrapper tr:hover td {
		background: rgba(255, 255, 255, 0.03);
	}

	.series-note {
		text-align: center;
		color: rgba(255, 255, 255, 0.6);
	}
</style>
