<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="I Set Up a Local LLM and Learned to Design Around Its Limitations"
	date="2026-01-16"
	description="Setting up Ollama on Mac Mini for local embeddings and SQL generation. Why a 3B model with validation beats 8B alone."
>
	<p>The next phase of my memory system needed embeddings - converting text to vectors for semantic search. Cloud APIs charge per token. My conversation history is 500+ megabytes and growing.</p>

	<p>Do the math. That's a lot of tokens. Running embeddings through Gemini or Anthropic every time I wanted to index new conversations would add up fast.</p>

	<p>But cost wasn't the only factor. I also wanted my conversation data to stay on my machine - not shipped off to cloud providers for processing. And honestly, I'd been wanting to get hands-on with local inference for a while. Running your own models feels like a rite of passage in 2026.</p>

	<p>The Mac Mini sitting on my desk seemed like the obvious choice.</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/hero_llm.png" alt="Mac Mini running local LLM" />
	</figure>

	<p>This is what I'm working with: entry-level M4 Mac Mini. 10-core CPU, 10-core GPU, and - crucially - <strong>16GB of unified memory</strong>. That RAM constraint shaped every decision that followed.</p>

	<p><em>(For non-Apple readers: "unified memory" means the CPU and GPU share the same RAM pool. Unlike discrete GPUs with their own VRAM, Apple Silicon can load large models directly into this shared memory. The 16GB is split between system use, the OS, and whatever models you're running - so every gigabyte counts.)</em></p>

	<hr />

	<h2>The Setup Plan</h2>

	<p>For this project, I needed two different models serving two different purposes.</p>

	<p>The first is an <strong>embeddings model</strong>. Its job is to convert text into vectors - numerical representations that capture semantic meaning. This is what powers semantic search. When you want to find "conversations about authentication" even if the word "authentication" never appears, embeddings make that possible. I went with <code>nomic-embed-text</code> - it's small (274MB), fast, and optimized specifically for this task.</p>

	<p>The second is a <strong>generation model</strong> - what most people think of when they hear "LLM." This one handles natural language to SQL conversion. When I eventually ask "What did I work on Friday?" via voice, this model turns that question into a database query. For this, I needed a chat-capable model that could understand instructions and produce structured output.</p>

	<p>The embeddings model was an easy choice. But for the generation model, I had a decision to make: which one?</p>

	<hr />

	<h2>Installing Ollama</h2>

	<p>Ollama is the runtime that manages and serves local LLMs. It handles model downloads, loads them into memory, and exposes a REST API for inference.</p>

	<h3>Step 1: Install</h3>

	<p>Download from <a href="https://ollama.com/download" target="_blank" rel="noopener">ollama.com</a>. Drag to Applications. Launch it.</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/scr_shot_01.png" alt="Ollama download page" />
	</figure>

	<p>Ollama runs as a menubar app - you'll see a small llama icon. It starts a local server on port 11434.</p>

	<p>Verify it's working:</p>

	<pre><code>ollama --version</code></pre>

	<figure class="article-image">
		<img src="/images/claude/local-llm/scr_shot_02.png" alt="Ollama version output" />
	</figure>

	<h3>Step 2: Pull Models</h3>

	<pre><code>{`# Embedding model - converts text to vectors
# Small (274MB), fast, optimized for semantic similarity
ollama pull nomic-embed-text

# Generation model - for text-to-SQL conversion
# We'll discuss why 3b instead of 8b below
ollama pull llama3.2:3b`}</code></pre>

	<figure class="article-image">
		<img src="/images/claude/local-llm/scr_shot_03.png" alt="Pulling models" />
	</figure>

	<p>Check what's installed:</p>

	<pre><code>ollama list</code></pre>

	<p>You should see:</p>

	<div class="comparison-table">
		<table>
			<thead>
				<tr>
					<th>Model</th>
					<th>Size</th>
					<th>Parameters</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td><code>llama3.2:3b</code></td>
					<td>2.0 GB</td>
					<td>3.2B</td>
				</tr>
				<tr>
					<td><code>nomic-embed-text</code></td>
					<td>274 MB</td>
					<td>137M</td>
				</tr>
			</tbody>
		</table>
	</div>

	<hr />

	<h2>The 8B vs 3B Debate</h2>

	<p>The common advice is bigger = better. Tim Toole recommends <code>llama3.1:8b</code>, which makes sense for a general-purpose assistant.</p>

	<p>But my use case isn't general-purpose. I'm not having philosophical debates with this model. I'm asking it to convert sentences into SQL queries. A structured, bounded task.</p>

	<p><strong>8B model:</strong> ~4-5GB RAM (with default Q4 quantization), "smarter"<br/>
	<strong>3B model:</strong> ~2GB RAM, leaves headroom</p>

	<p><em>(A note on quantization: Ollama models are typically 4-bit quantized by default - the Q4_K_M format. This compresses the model weights, trading a small amount of quality for significant memory savings. A full-precision 8B model would need ~16GB; quantized, it fits in ~5GB. The RAM numbers above assume default quantization.)</em></p>

	<p>My Mac Mini has 16GB total. macOS needs 4-6GB just to run. With 8B, the machine would be tight - especially if I want to run embeddings simultaneously. With 3B, there's room to breathe.</p>

	<p>I went with 3B. Conservative choice. Easy to upgrade if quality lacks.</p>

	<hr />

	<h2>Configuring Remote Access</h2>

	<p>Here's a detail that's easy to miss: Ollama only listens on localhost by default. My Mac Mini runs the models, but my main Mac (where DuckDB and my code live) needs to call the API over the network.</p>

	<pre><code>{`┌─────────────────────┐         ┌─────────────────────┐
│     Main Mac        │   API   │     Mac Mini        │
│  (DuckDB, code,     │ ──────► │  (Ollama, models)   │
│   dashboard)        │  calls  │                     │
└─────────────────────┘         └─────────────────────┘`}</code></pre>

	<p>On the Mac Mini, configure Ollama to listen on all network interfaces:</p>

	<pre><code>{`# Set Ollama to listen on all interfaces (not just localhost)
launchctl setenv OLLAMA_HOST "0.0.0.0"`}</code></pre>

	<p><strong>Important:</strong> This only works for the current session. After a reboot, Ollama reverts to localhost. For a persistent solution, add this to your shell profile:</p>

	<pre><code>{`# Add to ~/.zshrc (or ~/.bash_profile for bash)
export OLLAMA_HOST="0.0.0.0"

# Then source it
source ~/.zshrc`}</code></pre>

	<p>Then restart Ollama for the change to take effect. You can quit it from the menubar and relaunch, or:</p>

	<pre><code>osascript -e 'quit app "Ollama"' && open -a Ollama</code></pre>

	<p>Get your Mac Mini's IP address:</p>

	<pre><code>ipconfig getifaddr en0</code></pre>

	<figure class="article-image">
		<img src="/images/claude/local-llm/scr_shot_04.png" alt="Getting IP address" />
	</figure>

	<p>Now test the connection from your main Mac:</p>

	<pre><code>{`# Replace with your Mac Mini's IP
curl http://192.168.2.237:11434/api/tags`}</code></pre>

	<figure class="article-image">
		<img src="/images/claude/local-llm/scr_shot_05.png" alt="Testing API connection" />
	</figure>

	<p>JSON response listing installed models means it's working.</p>

	<p><strong>Security note:</strong> Binding to <code>0.0.0.0</code> means Ollama listens on all network interfaces. On a typical home network, your router's NAT prevents external access - only devices on your LAN can reach it. But if you're on a public network or have port forwarding enabled, Ollama could be reachable from the internet. For tighter control, use macOS firewall rules or SSH tunneling.</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/scr_shot_06.png" alt="Security configuration" />
	</figure>

	<hr />

	<h2>The Moment It Made a Mistake</h2>

	<p>Now we can finally test our new 3B model. I asked it to convert "Show me all projects from last week" to SQL.</p>

	<p>It generated:</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/scr_shot_07.png" alt="Model output with error" />
	</figure>

	<p>Looks reasonable. Except:</p>

	<pre><code class="language-sql">{`SELECT p.project_name   -- alias 'p'
FROM sessions s         -- table aliased as 's'`}</code></pre>

	<p>The table is aliased as <code>s</code>. The SELECT uses <code>p</code>. Run this query, get an error: "column p.project_name does not exist."</p>

	<p>A simple typo. The kind a human reviewer would catch instantly.</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/scr_shot_08.png" alt="Error output" />
	</figure>

	<hr />

	<h2>The Fork in the Road</h2>

	<p>Two options:</p>

	<p><strong>Option A: Use a bigger model.</strong> The 8B might make fewer mistakes. More parameters, more capability.</p>

	<p><strong>Option B: Build a validation loop.</strong> Assume the model will make mistakes. Catch them. Fix them automatically.</p>

	<p>I sat with this for a while.</p>

	<p>Option A treats the symptom. Option B addresses the root cause. Even an 8B model isn't 100% reliable. It would still occasionally make errors - just fewer of them. And when it did, the system would fail.</p>

	<p>Option B builds reliability into the system itself. The model is allowed to be imperfect because there's a safety net.</p>

	<p>This felt like an important principle: <strong>don't just throw more compute at the problem. Design around the limitation.</strong></p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/IMG_02.png" alt="Design decision illustration" />
	</figure>

	<hr />

	<h2>The Agentic Validation Loop</h2>

	<p>Here's what I built (conceptually - implementation comes in a later phase):</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/IMG_03.png" alt="Validation loop diagram" />
	</figure>

	<p>The beauty of SQL validation: just try to run it. DuckDB tells you exactly what's wrong.</p>

	<pre><code>{`Error: column p.project_name does not exist
Hint: Did you mean s.project_name?`}</code></pre>

	<p>Feed that error back to the LLM:</p>

	<blockquote>"Your SQL failed with error: 'column p.project_name does not exist'. The table 'sessions' is aliased as 's', not 'p'. Fix it."</blockquote>

	<p>The model sees its mistake, corrects it, and we try again. Usually works on the second attempt.</p>

	<hr />

	<h2>Why This Matters Beyond This Project</h2>

	<p>This isn't just about fixing a SQL typo. It's a fundamental design pattern for working with LLMs:</p>

	<p><strong>Naive approach:</strong> Trust the model → fail when it's wrong<br/>
	<strong>Agentic approach:</strong> Assume the model might be wrong → build verification → retry with feedback</p>

	<p>The 3B model with validation might actually be <em>more reliable</em> than 8B alone. The validation loop guarantees correctness through iteration. A bigger model just reduces the probability of error.</p>

	<p>I learned more from the model's mistake than I would have from it working perfectly.</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/IMG_04.png" alt="Comparison of approaches" />
	</figure>

	<hr />

	<h2>Testing the APIs</h2>

	<h3>Embedding API</h3>

	<p>The embedding model converts text to vectors - the foundation of semantic search.</p>

	<pre><code>{`curl http://192.168.2.237:11434/api/embeddings \\
  -d '{
    "model": "nomic-embed-text",
    "prompt": "How do I handle authentication errors?"
  }'`}</code></pre>

	<p>Response: a JSON object with 768 floating-point numbers representing the semantic meaning of that question.</p>

	<pre><code class="language-json">{`{"embedding":[0.311543..., 0.424029..., -3.988341..., ... (768 floats)]}`}</code></pre>

	<p><em>(The output above is simplified - actual response contains all 768 values. The 768 dimensions match BERT-base, a common standard. Higher dimensions capture more semantic nuance but require more storage and compute. For conversation search, 768 is plenty.)</em></p>

	<p>Two texts with similar meanings produce similar vectors. "authentication errors" and "login problems" would have vectors close together in this 768-dimensional space. That's how semantic search finds related content even when the words differ.</p>

	<h3>Generation API</h3>

	<p>The chat model handles text-to-SQL conversion.</p>

	<pre><code>{`curl http://192.168.2.237:11434/api/generate \\
  -d '{
    "model": "llama3.2:3b",
    "prompt": "Convert to SQL: Show all projects. Schema: sessions(session_id, project_name)",
    "stream": false
  }'`}</code></pre>

	<p>Response includes the generated text plus timing metrics:</p>

	<pre><code class="language-json">{`{
  "response": "SELECT * FROM sessions;",
  "total_duration": 11525437500,
  "load_duration": 8235500584,
  "eval_duration": 2463605841
}`}</code></pre>

	<p>The timing fields are in nanoseconds:</p>
	<ul>
		<li><strong><code>total_duration</code></strong> - Wall clock time for the entire request (~11.5 seconds here)</li>
		<li><strong><code>load_duration</code></strong> - Time to load model into memory (~8.2 seconds - only significant on cold start)</li>
		<li><strong><code>eval_duration</code></strong> - Actual inference time (~2.5 seconds - the "real" work)</li>
	</ul>

	<p>The <code>stream: false</code> flag gives you the complete response at once instead of token-by-token.</p>

	<hr />

	<h2>The Cold Start Problem</h2>

	<p>First API call to the generation model took 11 seconds. Subsequent calls: 2-3 seconds.</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/IMG_05.png" alt="Cold start timing" />
	</figure>

	<p>The difference is model loading. First call loads weights into RAM. After that, they stay loaded.</p>

	<p>For UX, this matters. If someone's first voice query takes 11 seconds, that's frustrating. Options:</p>

	<ul>
		<li>Pre-warm the model on startup</li>
		<li>Design the UI to expect delay on first query</li>
		<li>Keep the model loaded (Ollama does this automatically if there's recent activity)</li>
	</ul>

	<p><em>Tip: Ollama keeps models loaded for 5 minutes by default after the last request. You can extend this with <code>OLLAMA_KEEP_ALIVE</code>:</em></p>

	<pre><code>{`# Keep model loaded for 1 hour after last request
export OLLAMA_KEEP_ALIVE="1h"

# Or keep it loaded indefinitely (until Ollama restarts)
export OLLAMA_KEEP_ALIVE="-1"`}</code></pre>

	<p>I'll address this more thoroughly in Phase 7 (Voice Control). For now, knowing about it is enough.</p>

	<hr />

	<h2>What I Didn't Over-Engineer</h2>

	<p>I considered more elaborate solutions:</p>

	<p><strong>RAG (Retrieval-Augmented Generation)</strong> - Embed the schema and example queries, retrieve relevant context before generating SQL.</p>

	<p><strong>MCP tools</strong> - Let the model call <code>get_schema()</code> to discover table structure dynamically.</p>

	<p>Both are valid approaches for complex systems. But my schema is three tables. Total.</p>

	<p>Simple context injection works fine:</p>

	<pre><code>{`Convert to SQL: "Show me all projects from last week"

Available tables:
- sessions(session_id, project_name, first_message_at, last_message_at)
- messages(id, session_id, type, timestamp, content)
- tool_calls(id, session_id, tool_name, timestamp)

Return only SQL.`}</code></pre>

	<p>Include the schema in every prompt. Combined with the validation loop, this is robust without added complexity.</p>

	<figure class="article-image">
		<img src="/images/claude/local-llm/IMG_06.png" alt="Simple vs complex approaches" />
	</figure>

	<p>Key principle: don't create contrivance. Simple solutions first. Add complexity only when simple fails.</p>

	<hr />

	<h2>What This Enables</h2>

	<p>After Phase 3:</p>

	<div class="comparison-table">
		<table>
			<thead>
				<tr>
					<th>Capability</th>
					<th>How</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>Generate embeddings locally</td>
					<td><code>POST /api/embeddings</code> with nomic-embed-text</td>
				</tr>
				<tr>
					<td>Generate SQL from natural language</td>
					<td><code>POST /api/generate</code> with llama3.2:3b</td>
				</tr>
				<tr>
					<td>No cloud dependency</td>
					<td>Everything runs on Mac Mini</td>
				</tr>
				<tr>
					<td>No per-token cost</td>
					<td>Embed entire 500MB+ history for free</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>Phase 4 uses the embedding model to enable semantic search.<br/>
	Phase 7 uses the generation model for voice-to-SQL queries.</p>

	<p>The infrastructure is in place. Time to use it.</p>

	<hr />

	<h2>What I Learned</h2>

	<p><strong>Model errors are design opportunities.</strong> The alias mismatch bug led me to implement an agentic validation loop - a production-grade pattern I wouldn't have learned if the model worked perfectly.</p>

	<p><strong>Start conservative, upgrade if needed.</strong> 3B with validation beats 8B without. And upgrading is one command away if needed.</p>

	<p><strong>Remote access requires explicit configuration.</strong> <code>OLLAMA_HOST=0.0.0.0</code> isn't the default. Easy to miss if you're running everything on one machine.</p>

	<p><strong>Cold start vs warm performance.</strong> First inference loads the model. Design your UX around this or pre-warm.</p>

	<hr />

	<h2>The Bigger Picture</h2>

	<p>This phase plugged several leaks at once.</p>

	<p>Any cloud API dependency inherently introduces more friction and/or brittleness - not just cost, but latency, complexity, and the nagging knowledge that my conversation data was leaving my machine. Now it stays local.</p>

	<p>The 3B model error was friction too, or it could have been. Instead of throwing more compute at the problem, I designed around it. The validation loop doesn't just catch mistakes - it guarantees correctness through iteration. That's a pattern I'll use again.</p>

	<p>But the real unlock is what this enables. Right now, searching my Claude history means keyword matching. "Authentication" finds "authentication." It doesn't find "login issues" or "credential handling" or any of the dozen other ways to express the same concept. That's friction - the gap between what I mean and what I can find.</p>

	<p>Phase 4 changes that. The embeddings model I just set up will convert every message into a vector - a numerical fingerprint of meaning. Similar concepts cluster together, regardless of the exact words used. Search by meaning, not by string matching.</p>

	<p>And later on we'll take it further. Instead of writing SQL queries to analyze my collaboration patterns, I'll just ask: "What did I work on Friday?" The generation model converts natural language to database queries. Voice in, insights out.</p>

	<p>Each phase removes another barrier between question and answer. That's the point. Not the technology itself - the friction it eliminates.</p>

	<hr />

	<p class="series-note"><em>Part 3 of the Claude Memory Project. Next: semantic search with LanceDB.</em></p>
</ArticleLayout>

<style>
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

	.series-note {
		text-align: center;
		color: rgba(255, 255, 255, 0.6);
	}
</style>
