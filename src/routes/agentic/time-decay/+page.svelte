<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="The Obvious Approach Was Wrong"
	date="2026-02-27"
	description="I expected time-decay scoring to be a straightforward improvement. Apply decay, watch numbers improve, ship it. Instead I was staring at a 35% degradation."
>
	<figure class="article-image">
		<img src="/images/claude/time-decay/IMG-001-HERO.png" alt="Time decay scoring hero" />
	</figure>

	<p>In the last article I built a benchmark suite and put numbers on my memory system's failures. The headline: 50% answer-found rate for keyword search, 25% for semantic. Cross-session queries scored zero. Contradiction queries were actively misleading - returning old answers alongside new ones with no way to tell which was current.</p>

	<p>That benchmark gave me a scorecard. Time-decay scoring was first on the upgrade list - the idea that recent messages should rank higher than older ones, so when I ask "how many layers does Grimoire have?" I get the current answer (3) instead of the outdated one (4).</p>

	<p>Seems pretty straight-forward imo - multiply the relevance score by an exponential decay factor. Recent messages keep their weight, old messages fade. I figured I'd implement it, rerun the 28 queries, and watch the numbers go up.</p>

	<p>The numbers went down.</p>

	<p>Womp-womp.</p>

	<hr />

	<h2>The Setup</h2>

	<p>Before writing any decay code, I profiled the data to understand what I was working with. The age distribution told me something important:</p>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Age Bucket</th>
					<th>Messages</th>
					<th>Percentage</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>0-7 days</td>
					<td>41,197</td>
					<td>25.7%</td>
				</tr>
				<tr>
					<td>8-14 days</td>
					<td>33,725</td>
					<td>21.0%</td>
				</tr>
				<tr>
					<td>15-30 days</td>
					<td>84,275</td>
					<td>52.5%</td>
				</tr>
				<tr>
					<td>31-60 days</td>
					<td>1,209</td>
					<td>0.8%</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>The system spans only 31 days of data. No messages older than 60 days. More than half the messages are 15-30 days old.</p>

	<p>I've been using Claude Code daily since late December - about two months at this point. So where's the first month? Turns out Claude Code by default deletes session JSONL files older than 30 days. So, I found out my December and early January conversations were wiped before I even knew there was a retention policy. Now I have ways of restoring that, but more importantly if you wanna make sure that does not happen to you, add this to <code>~/.claude/settings.json</code>:</p>

	<pre><code>{`"cleanupPeriodDays": 99999`}</code></pre>

	<p>Just FYI - Don't set it to 0. There's a known bug where 0 disables transcript writing entirely instead of disabling cleanup. Use a large number.</p>

	<p>Anyway - back to the data I do have. The decay function's entire operating range is compressed into a single month. A 30-day half-life would barely differentiate anything, while a 7-day half-life would create a cliff within the last two weeks. Not a lot of room to work with, but enough to test the idea.</p>

	<p>So I implemented an exponential decay function: <code>decay(age) = exp(-ln(2) / half_life * age_days)</code>. At the half-life, a message retains 50% of its original score. At twice the half-life, 25%. Standard radioactive decay curve. I applied it to both search engines - multiplying BM25 scores for keyword search, converting L2 distance to similarity and multiplying for semantic search. To compensate for the re-ranking, I overfetched 3x results from each engine, applied decay, re-sorted, and returned the top-K.</p>

	<p>Then I ran a sweep across three half-lives: 7 days, 14 days, and 30 days.</p>

	<hr />

	<h2>Pure Decay: A Disaster</h2>

	<figure class="article-image">
		<img src="/images/claude/time-decay/IMG-002-PURE-DECAY-DISASTER.png" alt="Pure decay results showing degradation" />
	</figure>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Half-life</th>
					<th>KW answer-found</th>
					<th>KW delta</th>
					<th>KW recency</th>
					<th>SM answer-found</th>
					<th>SM delta</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>baseline</td>
					<td>0.500</td>
					<td>-</td>
					<td>0.100</td>
					<td>0.250</td>
					<td>-</td>
				</tr>
				<tr>
					<td>7d</td>
					<td>0.321</td>
					<td><strong>-0.179</strong></td>
					<td>1.000</td>
					<td>0.107</td>
					<td>-0.143</td>
				</tr>
				<tr>
					<td>14d</td>
					<td>0.321</td>
					<td><strong>-0.179</strong></td>
					<td>1.000</td>
					<td>0.107</td>
					<td>-0.143</td>
				</tr>
				<tr>
					<td>30d</td>
					<td>0.429</td>
					<td>-0.071</td>
					<td>0.800</td>
					<td>0.179</td>
					<td>-0.071</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>This was not what I expected to see. Keyword answer-found dropped from 50% to 32.1% at 7-day half-life - a 35.8% degradation. Semantic search went from 25% to 10.7%. The feature I built to make search better had made it meaningfully worse across the board.</p>

	<p>But then I looked at the recency column and had one of those moments where the data tells two stories at once. Recency went from 0.100 to 1.000. Perfect. The most recent relevant result was in the top 3 every single time. The decay function was doing exactly what I asked it to - promoting recent results. It just turns out that promoting recent results and finding correct answers are not the same thing.</p>

	<p>Think about it this way: a 30-day-old message explaining that Numinon is written in Go has its BM25 score cut by 94% at 7d half-life. Doesn't matter that the message perfectly answers the question. A message from yesterday that happens to mention Go in passing now outranks it. The intuition that "recent equals relevant" only holds when you're asking about something that changed recently. For stable facts - what language something is written in, what port a service runs on, what embedding model the system uses - the original answer is just as valid months later. Pure decay can't tell the difference. It treats every query as if it's asking about something that might have changed.</p>

	<p>This was humbling. I'd gone in confident that this was a straightforward improvement. Apply decay, watch numbers improve, ship it. Instead I was staring at a 35% degradation and wondering if the whole approach was fundamentally flawed.</p>

	<hr />

	<h2>The Floor</h2>

	<p>After contemplating this for awhile and doing some research, it struck me that the problem wasn't the decay curve itself - it was the asymptote. Pure exponential decay approaches zero. A 30-day-old message at 7d half-life gets a factor of 0.06. That's not "slightly less relevant" - that's functionally discarded. I was essentially deleting old results from the ranking while pretending I was gently re-ordering them.</p>

	<p>What if the decay function had a minimum? A floor below which scores can never drop?</p>

	<pre><code>{`decay(age) = floor + (1 - floor) * exp(-ln(2) / half_life * age_days)`}</code></pre>

	<p>The idea is simple: with a floor of 0.5, the oldest message in the system still retains at least 50% of its original relevance score. The decay still boosts recent results - a brand new message gets factor 1.0, a two-week-old message gets 0.75 - but nothing gets buried. The curve gently nudges instead of cliff-dropping.</p>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Age (days)</th>
					<th>Pure Decay (14d)</th>
					<th>Soft Decay (14d, floor=0.5)</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>0</td>
					<td>1.000</td>
					<td>1.000</td>
				</tr>
				<tr>
					<td>7</td>
					<td>0.707</td>
					<td>0.854</td>
				</tr>
				<tr>
					<td>14</td>
					<td>0.500</td>
					<td>0.750</td>
				</tr>
				<tr>
					<td>21</td>
					<td>0.354</td>
					<td>0.677</td>
				</tr>
				<tr>
					<td>30</td>
					<td>0.228</td>
					<td>0.614</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>Look at the difference at 30 days. That gap is the difference between "this old message is basically invisible" and "this old message still has a fighting chance if it's actually relevant." I ran the full sweep again with floor=0.5 across all three half-lives, genuinely curious whether this would be enough to save the approach.</p>

	<hr />

	<h2>Soft Decay: The Sweet Spot</h2>

	<figure class="article-image">
		<img src="/images/claude/time-decay/IMG-003-WHERE-IT-WORKED.png" alt="Soft decay results showing improvement" />
	</figure>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Half-life</th>
					<th>KW answer-found</th>
					<th>KW delta</th>
					<th>KW recency</th>
					<th>Contradiction</th>
					<th>Factual</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>baseline</td>
					<td>0.500</td>
					<td>-</td>
					<td>0.100</td>
					<td>0.800</td>
					<td>0.750</td>
				</tr>
				<tr>
					<td>7d</td>
					<td>0.464</td>
					<td>-0.036</td>
					<td>0.900</td>
					<td>1.000</td>
					<td>0.625</td>
				</tr>
				<tr>
					<td><strong>14d</strong></td>
					<td><strong>0.500</strong></td>
					<td><strong>0.000</strong></td>
					<td><strong>0.800</strong></td>
					<td><strong>1.000</strong></td>
					<td>0.625</td>
				</tr>
				<tr>
					<td>30d</td>
					<td>0.464</td>
					<td>-0.036</td>
					<td>0.600</td>
					<td>1.000</td>
					<td>0.500</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>The 14-day half-life with floor=0.5 row seemed a wee bit befuddling... Zero degradation in overall answer-found rate - the same 50% as baseline. OK. But recency jumped from 10% to 80%, meaning the system now consistently surfaces the most recent relevant result near the top, so some wins in the mix.</p>

	<p>But the real interesting part is the contradiction column, which went from 80 to 100%. As a reminder, this represents the category where the system was returning outdated information that actively misled, where you could receive the wrong answer because old messages had the potential to outrank current ones. Well, that category now scores 100% - meaning every single contradiction query returns the correct, current answer.</p>

	<p>But it ain't all rosy, and its important to be aware of it. No cherry-picking over here, mind you...</p>

	<p>Factual queries dropped from 75% to 62.5% - one query that used to be found is now missed. Now sample size is too small to be significant here, and it possible, perhaps even likely this kinda variation will smooth out as my dataset grows. But for the moment, it's worth noting, but imo even if some drop persists it's likely adequately compensated for by the contradiction gains.</p>

	<p>Here's the full per-category breakdown for the winning config:</p>

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					<th>Category</th>
					<th>Baseline</th>
					<th>With Decay</th>
					<th>Delta</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>Factual (8 queries)</td>
					<td>0.750</td>
					<td>0.625</td>
					<td>-0.125</td>
				</tr>
				<tr>
					<td>Temporal (5 queries)</td>
					<td>0.400</td>
					<td>0.400</td>
					<td>0.000</td>
				</tr>
				<tr>
					<td>Contradiction (5 queries)</td>
					<td>0.800</td>
					<td><strong>1.000</strong></td>
					<td><strong>+0.200</strong></td>
				</tr>
				<tr>
					<td>Cross-session (5 queries)</td>
					<td>0.000</td>
					<td>0.000</td>
					<td>0.000</td>
				</tr>
				<tr>
					<td>Specificity (5 queries)</td>
					<td>0.400</td>
					<td>0.400</td>
					<td>0.000</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>You'll also notice that three categories didn't move at all, and it makes sense when you think about what decay actually does. Time-decay doesn't help find when something was first built - that requires finding OLD messages, and decay actively works against it. It doesn't help synthesize across sessions, which requires aggregation, not re-ranking. And it doesn't help find specific values that were already findable or unfindable for other reasons.</p>

	<p>What it does help, is promoting the current version of a fact when multiple versions exist in the database. Now again, this is a narrower win than I expected going in, but I'll remeasure these results monthly as my dataset grows to get a more accurate sense of the trend. For now, this feels good enough.</p>

	<hr />

	<h2>Semantic Search: Don't Bother</h2>

	<figure class="article-image">
		<img src="/images/claude/time-decay/IMG-004-SEMANTIC-NOISE.png" alt="Semantic search decay results" />
	</figure>

	<p>I also tested floor=0.3 across all half-lives to look for a middle ground between pure- and soft decay. Between all 9 configurations plus the baseline, the results told a clear story for semantic search: nothing helps. Not a single configuration improved it. The best case (7d/floor=0.5) only dropped answer-found by 0.036, but it didn't improve anything either.</p>

	<p>I'm no expert, in fact I'm largely doing all of this not only to get a useful tool but to learn, but if I had to venture a wager I think it probably comes down to signal quality. L2 vector distance is already a noisy signal. When two messages have distances of 250 and 255, that 5-unit gap is insignificant - it's within the noise floor of what the embedding model can distinguish. Multiplying both by slightly different decay factors doesn't add useful information; it amplifies noise that was already there.</p>

	<p>BM25, on the other hand, produces scores with real separation. A message that mentions "Numinon" and "Go" in the same sentence gets a score meaningfully higher than one that just mentions "Go" in passing. That signal is strong enough to absorb a gentle multiplier without losing its ranking power. Vectors don't have that luxury.</p>

	<p>Decision: Don't apply time-decay to semantic search, keyword search only.</p>

	<hr />

	<h2>Deploying It</h2>

	<p>The production change was small - maybe 30 lines of Go. In the MCP server that handles keyword search, I added a <code>computeDecayFactor</code> function implementing the soft decay formula, modified the search handler to overfetch 3x results, apply decay factors, re-sort, and trim to the requested limit.</p>

	<p>Controlled by three constants at the top of the file: half-life 14 days, floor 0.5, overfetch factor 3. Nothing fancy, nothing clever.</p>

	<p>Though not perceptible, it's always worth examining performance impact, in this case keyword search went from ~56ms to ~62ms per query, so 3x overfetch adds 6 milliseconds. That's a rounding error, nothing to fret about.</p>

	<hr />

	<h2>What I Learned</h2>

	<figure class="article-image">
		<img src="/images/claude/time-decay/IMG-005-WHAT-I-LEARNED.png" alt="Key takeaways from time decay implementation" />
	</figure>

	<p><strong>Pure decay is a trap.</strong> The intuition seems airtight - recent messages are more relevant, just apply a time penalty to old ones. But relevance and recency are partially correlated, not identical. BM25 already encodes relevance. Blindly discounting it by age throws away good information.</p>

	<p><strong>The floor is the insight, not the curve.</strong> What makes time-decay viable isn't the exponential function - it's the floor parameter that prevents old results from being destroyed. Without a floor, you're choosing between "good recency but bad answers" and "good answers but bad recency." With a floor of 0.5, you get both. I keep coming back to this because it's counterintuitive: the curve shape matters less than the guarantee that no result loses more than half its relevance score.</p>

	<p><strong>Different engines need different treatment.</strong> I assumed time-decay would help both search paths equally. It helps keyword search, where BM25 scores are a strong, stable signal that can absorb a gentle multiplier. It hurts semantic search, where vector distances are noisy and decay amplifies the noise. If I hadn't tested both separately, I would have degraded semantic search and called the whole feature an improvement because keyword search got better. That's the kind of mistake you only catch by measuring carefully.</p>

	<p><strong>The benchmark caught what I wouldn't have.</strong> This is the one that really sticks with me. Without the Phase 0 benchmark suite, I would have deployed pure decay and subjectively felt that search was "about the same" - a 35.8% degradation, completely invisible in day-to-day use. The whole reason I found the floor parameter is because the pure decay numbers were so obviously bad that I couldn't ignore them.</p>

	<hr />

	<h2>What's Next</h2>

	<p>Time-decay is a re-ranking strategy. It can reorder existing results but can't surface information that wasn't in the candidate set to begin with. Cross-session synthesis is still at 0%. Temporal queries about the past are still weak. Content noise - tool output and thinking blocks crowding out actual answers - is untouched.</p>

	<p>The next phase is fact extraction: using an LLM to distill raw messages into atomic, timestamped facts that can be searched independently. "Grimoire has 3 layers" as a fact, not as a sentence buried in a 500-word message about project architecture. That's where the real improvement should come from - not better ranking of the same raw messages, but better data to rank in the first place.</p>
</ArticleLayout>
