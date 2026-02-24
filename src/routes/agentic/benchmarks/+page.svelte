<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="Before You Upgrade Anything, Measure What You Have"
	date="2026-02-24"
	description="My memory system finds the correct answer less than a third of the time. Here are the 28-query benchmark results across five categories that prove it."
>
	<figure class="article-image">
		<img src="/images/claude/benchmarks/IMG-001-HERO.png" alt="Memory system benchmark hero" />
	</figure>

	<p>In the last article I audited my memory system against <a href="https://x.com/rohit4verse">@rohit4verse's</a> architecture and found a ginormous gap: I'd built a solid storage-and-retrieval layer but had zero intelligence between storage and retrieval - fact extraction, summaries, conflict resolution, decay, all missing. A well-indexed chat log, not actual memory.</p>

	<p>In that audit I formulated a seven-phase upgrade plan. But before cracking my knuckles and getting to work, I needed to answer a simpler question: how bad is the gap, exactly? Not measured in vibes, but in numbers.</p>

	<hr />

	<h2>The Benchmark</h2>

	<p>I wrote 28 test queries across five categories, each with a known correct answer that I verified exists in my 90,000-message history:</p>

	<ul>
		<li><strong>Factual</strong> (8 queries): Basic recall. "What embedding model do we use?" Expected: nomic-embed-text.</li>
		<li><strong>Temporal</strong> (5 queries): Time-sensitive. "When was the MCP server first built?" Expected: January 13, 2026.</li>
		<li><strong>Contradiction</strong> (5 queries): Facts that changed. "How many layers does Grimoire have?" Was 4, now 3.</li>
		<li><strong>Cross-session</strong> (5 queries): Requires aggregation. "What components make up Artifex?" Spread across dozens of sessions.</li>
		<li><strong>Specificity</strong> (5 queries): Precise answers. "What port does Ollama listen on?" Expected: 11434.</li>
	</ul>

	<p>Each query runs against both search engines. I score on five dimensions: keyword hit rate, verbatim answer found, top-result quality, recency appropriateness, and token cost.</p>

	<p>I used deterministic heuristics instead of LLM-as-judge. Keyword presence, string matching. I want a baseline I can reproduce exactly.</p>

	<hr />

	<h2>The Results</h2>

	<p>Overall answer-found rate: <strong>28.6%</strong> for keyword search, <strong>25.0%</strong> for semantic.</p>

	<p>Whuuuut? My memory system can find the verbatim correct answer less than a third of the time.</p>

	<p>Here's the breakdown by category:</p>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Category</th>
					<th>Keyword (answer found)</th>
					<th>Semantic (answer found)</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>Factual</td>
					<td>50%</td>
					<td>62.5%</td>
				</tr>
				<tr>
					<td>Temporal</td>
					<td>20%</td>
					<td>0%</td>
				</tr>
				<tr>
					<td>Contradiction</td>
					<td>20%</td>
					<td>0%</td>
				</tr>
				<tr>
					<td>Cross-session</td>
					<td>0%</td>
					<td>0%</td>
				</tr>
				<tr>
					<td>Specificity</td>
					<td>40%</td>
					<td>40%</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>The factual category is decent, mind you, and interestingly semantic search actually beats keyword here because embeddings catch conceptual matches that BM25 misses. If you ask about the embedding model using different words than appear in the data, semantic search still finds it.</p>

	<p>Everything else is pretty rough ngl.</p>

	<p><strong>Temporal: 20% / 0%.</strong> Neither engine understands time. They treat a message from six weeks ago the same as one from yesterday. When I ask "what was the most recently discussed project?" I get results from random dates.</p>

	<p><strong>Contradiction: 20% / 0%.</strong> This is the one that stings. Grimoire went from four layers to three. Intuition was renamed to Thelema. Agent Hub became Pneuma. The system happily returns the old answers alongside the new ones with no way to know which is current. It's actively misleading.</p>

	<p><strong>Cross-session: 0% / 0%.</strong> Total failure. Zero queries answered. "What components make up Artifex?" requires synthesizing information from dozens of conversations. Each individual message mentions one or two components. Neither engine can aggregate. This is the fundamental limitation of message-level retrieval - you get fragments, never the whole picture.</p>

	<hr />

	<h2>What I Learned</h2>

	<figure class="article-image">
		<img src="/images/claude/benchmarks/IMG-003-WHAT-I-LEARNED.png" alt="What I learned from the benchmark results" />
	</figure>

	<p><strong>Keyword search is still king for precision.</strong> It wins 11 queries to semantic's 7. When you're looking for a specific port number, IP address, or model name, BM25 with Porter stemming is hard to beat. It's also twice as fast (50ms vs 92ms per query).</p>

	<p><strong>Semantic search wins on conceptual flexibility.</strong> It finds the embedding model discussion even when the query doesn't use the word "nomic." It connects "coding language preferences" to conversations about Python and Go. But this flexibility comes with noise - semantically related isn't the same as factually correct.</p>

	<p><strong>The problem is what's being searched, not search quality.</strong> Individual messages are the wrong unit of knowledge. A message that says "let's use three layers" only makes sense if you know there was previously a four-layer design. A message mentioning "Pneuma" only makes sense if you know it used to be called Agent Hub. The search engines are doing their job - finding relevant messages. But messages aren't knowledge.</p>

	<hr />

	<h2>What's Next</h2>

	<p>This benchmark becomes the scorecard for every upgrade:</p>

	<ol>
		<li><strong>Time-decay scoring</strong> - weight recent results higher, especially for temporal and contradiction queries</li>
		<li><strong>Fact extraction</strong> - distill messages into versioned facts that supersede older ones</li>
		<li><strong>Entity graph</strong> - connect components, decisions, and preferences across sessions</li>
		<li><strong>Query expansion</strong> - automatically include old names when searching for renamed things</li>
		<li><strong>Content filtering</strong> - stop indexing tool output and thinking blocks that add noise</li>
	</ol>

	<p>Each phase reruns these same 28 queries. The numbers either go up or they don't. If they don't, the upgrade didn't work.</p>

	<p>The system I have today stores conversations. The system I want manages knowledge. This benchmark is the gap between those two things, measured in five dimensions across 28 queries.</p>

	<p>Time to close the gap - let's do this.</p>
</ArticleLayout>
