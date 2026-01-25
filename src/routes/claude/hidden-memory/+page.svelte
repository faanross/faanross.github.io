<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="I Discovered Claude Code Has a Hidden Memory"
	date="2025-01-11"
	description="Claude Code stores complete conversation transcripts locally. Here's how to unlock persistent memory across sessions with one addition to your CLAUDE.md."
>
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

	<hr />

	<h2>Quick primer: What is JSONL?</h2>

	<p>JSONL (JSON Lines) is a simple format where each line of a file is a complete, valid JSON object. Unlike regular JSON which wraps everything in one structure, JSONL lets you append new records without parsing the entire file.</p>

	<pre><code>{`{"type": "user", "message": "first message", "timestamp": "..."}
{"type": "assistant", "message": "response", "timestamp": "..."}
{"type": "user", "message": "second message", "timestamp": "..."}`}</code></pre>

	<p>One object per line. No commas between lines. No wrapping array.</p>

	<p>When you <code>cat</code> a JSONL file and it looks like a wall of text, that's because each JSON object can be thousands of characters long - they wrap visually in your terminal but are still single lines. A typical Claude response might be 70,000+ characters on one line.</p>

	<p>This format is perfect for logs and conversation history - new messages just append to the file, and you can stream through gigabytes of data line by line without loading everything into memory.</p>

	<hr />

	<h2>What's actually in there</h2>

	<p>Each session creates a JSONL file named with its session UUID. But it's not just messages - there are <strong>five distinct record types</strong>:</p>

	<div class="data-table">
		<table>
			<thead>
				<tr>
					<th>Type</th>
					<th>What It Captures</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td><code>user</code></td>
					<td>Your messages (typed or voice-transcribed)</td>
				</tr>
				<tr>
					<td><code>assistant</code></td>
					<td>Claude's responses (including thinking blocks)</td>
				</tr>
				<tr>
					<td><code>progress</code></td>
					<td>Tool execution progress (MCP calls, file operations)</td>
				</tr>
				<tr>
					<td><code>system</code></td>
					<td>Internal events (hooks firing, stop reasons)</td>
				</tr>
				<tr>
					<td><code>file-history-snapshot</code></td>
					<td>Snapshots of file state during edits</td>
				</tr>
			</tbody>
		</table>
	</div>

	<p>In a typical session, you'll see roughly 2x as many assistant messages as user messages, plus progress events for every tool call, system events for hooks, and file snapshots when Claude edits files.</p>

	<p><em>Note: The examples below are sanitized and simplified to illustrate the structure. Real entries contain additional metadata fields and much longer content.</em></p>

	<h3>1. User messages</h3>

	<p>Every message you send, whether typed or spoken via voice mode:</p>

	<pre><code>{`{
  "type": "user",
  "message": {
    "role": "user",
    "content": "Can you help me refactor this function?"
  },
  "timestamp": "2026-01-17T10:05:53.497Z",
  "sessionId": "c4bdaf0d-e808-42ce-90d0-6536fbf7983b",
  "cwd": "/Users/you/your-project",
  "uuid": "56f6c539-729a-4ed5-95cb-47aaeb10af20",
  "gitBranch": "main",
  "version": "2.1.11"
}`}</code></pre>

	<h3>2. Assistant messages</h3>

	<p>Claude's responses, including the full content and thinking blocks (if extended thinking is enabled):</p>

	<pre><code>{`{
  "type": "assistant",
  "message": {
    "role": "assistant",
    "content": [
      {
        "type": "thinking",
        "thinking": "Let me analyze the function structure..."
      },
      {
        "type": "text",
        "text": "I can see several opportunities to improve this..."
      }
    ]
  },
  "timestamp": "2026-01-17T10:05:57.413Z",
  "sessionId": "c4bdaf0d-e808-42ce-90d0-6536fbf7983b",
  "usage": {
    "input_tokens": 9500,
    "output_tokens": 1200
  }
}`}</code></pre>

	<h3>3. Progress events</h3>

	<p>Every tool call generates progress events - when MCP servers are invoked, when files are read, when bash commands run:</p>

	<pre><code>{`{
  "type": "progress",
  "data": {
    "type": "mcp_progress",
    "status": "started",
    "serverName": "voicemode",
    "toolName": "converse"
  },
  "toolUseID": "toolu_01DRAfS2ccqm8zYenJT6PR2R",
  "timestamp": "2026-01-17T10:06:10.830Z",
  "sessionId": "c4bdaf0d-e808-42ce-90d0-6536fbf7983b"
}`}</code></pre>

	<h3>4. System events</h3>

	<p>Internal events like hooks firing, session stops, and other system-level signals:</p>

	<pre><code>{`{
  "type": "system",
  "subtype": "stop_hook_summary",
  "hookCount": 1,
  "hookInfos": [
    { "command": "sh /path/to/hook-wrapper.sh handle-hook Stop" }
  ],
  "hookErrors": [],
  "stopReason": "",
  "timestamp": "2026-01-17T10:07:56.454Z",
  "sessionId": "c4bdaf0d-e808-42ce-90d0-6536fbf7983b"
}`}</code></pre>

	<h3>5. File history snapshots</h3>

	<p>When Claude edits files, snapshots capture the state - enabling the undo functionality:</p>

	<pre><code>{`{
  "type": "file-history-snapshot",
  "messageId": "7e91a3ea-e6b5-4ddb-a944-4faad3eb24ec",
  "isSnapshotUpdate": false,
  "snapshot": {
    "messageId": "7e91a3ea-e6b5-4ddb-a944-4faad3eb24ec",
    "timestamp": "2026-01-17T10:06:30.000Z",
    "trackedFileBackups": {
      "src/components/Button.tsx": { "content": "..." }
    }
  }
}`}</code></pre>

	<hr />

	<h2>File organization</h2>

	<p>The files are organized by the directory you're working in. So all your sessions in <code>/Users/you/project-a/</code> live in one folder, sessions in <code>/Users/you/project-b/</code> in another. The folder names use dashes instead of slashes:</p>

	<pre><code>{`~/.claude/projects/
├── -Users-you-project-a/
│   ├── c4bdaf0d-e808-42ce-90d0-6536fbf7983b.jsonl
│   ├── a1b2c3d4-e5f6-7890-abcd-ef1234567890.jsonl
│   ├── subagents/          ← spawned sub-agents
│   └── ...
├── -Users-you-project-b/
│   └── ...
└── -Users-you-Documents-work/
    └── ...`}</code></pre>

	<p>You'll also notice <code>subagents/</code> folders within project directories. These contain conversations from spawned sub-agents - separate Claude instances that handle delegated tasks. For most queries, you'll want to exclude these (hence <code>! -path "*/subagents/*"</code> in the commands below) to focus on your direct conversations.</p>

	<hr />

	<h2>Privacy and security considerations</h2>

	<p>Before you get too excited about this data goldmine, some things to keep in mind:</p>

	<p><strong>No encryption at rest.</strong> These are plain text JSON files. Anyone with access to your machine - or your backups - can read your complete conversation history.</p>

	<p><strong>File snapshots contain actual file contents.</strong> The <code>file-history-snapshot</code> records store the content of files Claude edits. If you're working with sensitive code, API keys, credentials, or proprietary information, that content lives in your JSONL files too.</p>

	<p><strong>Your prompts reveal your thinking.</strong> Every question you've asked, every problem you've described, every piece of context you've provided - it's all there. Consider what that history reveals about your projects, your knowledge gaps, and your workflow.</p>

	<p>This isn't a reason not to use the feature - just be aware of what you're storing and where. If you're on a shared machine or backing up to cloud storage, factor this into your security posture.</p>

	<hr />

	<h2>Why this matters</h2>

	<p>I'd been wanting something like this since I started using AI assistants. A way to go back and find "that thing Claude explained last week." A searchable record of decisions and solutions.</p>

	<p>But I assumed I'd have to build it. Export conversations manually. Set up logging. Create some elaborate capture system.</p>

	<p>Turns out it already existed. I just didn't know where to look.</p>

	<p>And it's better than what I would have built:</p>

	<ul>
		<li><strong>Complete</strong> - both sides of every conversation, not just my input</li>
		<li><strong>Automatic</strong> - no export step, no manual logging</li>
		<li><strong>Local</strong> - stored on your disk, no cloud retention policy eating your history after 30 days (though future Claude Code updates could change the storage format or location)</li>
		<li><strong>Parseable</strong> - standard JSONL format, easy to query with basic tools</li>
	</ul>

	<hr />

	<h2>What I realized when comparing approaches</h2>

	<p>I'd seen <a href="https://www.linkedin.com/in/artemxtech" target="_blank" rel="noopener noreferrer">Artem Zhutov</a> posting about analyzing his Claude conversations. He uses Wispr Flow - a voice dictation tool that captures everything he speaks into any app. 956K words across all applications. 147K words to Claude Desktop alone.</p>

	<p>That's a lot of data. But look at what each approach actually captures:</p>

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
					<td>Local disk, no auto-expiry</td>
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

	<p>The data was already there. I just needed to build a system to actually use it.</p>

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

	<p>Filenames are UUIDs (e.g., <code>c4bdaf0d-e808-42ce-90d0-6536fbf7983b.jsonl</code>). Timestamps are inside the files, not in filenames - use modification time or parse internal timestamps to find sessions by date.</p>

	<h3>Find sessions by date</h3>

	<p>Looking for what you worked on Tuesday?</p>

	<pre><code>{`find ~/.claude/projects -name "*.jsonl" -type f ! -path "*/subagents/*" -newermt "2026-01-07" ! -newermt "2026-01-08" -exec ls -la {} \\;`}</code></pre>

	<h3>Search across all conversations</h3>

	<p>That solution you can't quite remember? Grep it:</p>

	<pre><code>{`grep -r "MITRE" ~/.claude/projects/ --include="*.jsonl" | head -20`}</code></pre>

	<h3>Parse a session into readable format</h3>

	<p>Note: The <code>.message.content</code> structure varies - sometimes it's a string (user messages), sometimes an array of objects (assistant messages with thinking blocks). This command extracts just user/assistant messages:</p>

	<pre><code>{`cat session-file.jsonl | jq -s '[.[] | select(.type == "user" or .type == "assistant")] | .[] | {type, time: .timestamp}'`}</code></pre>

	<p>For a more detailed view that handles both content formats:</p>

	<pre><code>{`cat session-file.jsonl | jq -s '.[] | select(.type == "user") | {type, time: .timestamp, content: .message.content}'`}</code></pre>

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

Files are organized by working directory path. Each session creates a JSONL file (one JSON object per line) named with a UUID. Subagent conversations are stored in \`subagents/\` subdirectories - exclude these with \`! -path "*/subagents/*"\` to focus on direct conversations.

### Record Types

Five types of records are stored:
- **user** - Your messages (typed or voice-transcribed)
- **assistant** - Claude's responses (including thinking blocks)
- **progress** - Tool execution events (MCP calls, file reads)
- **system** - Internal events (hooks, stop reasons)
- **file-history-snapshot** - File state for undo functionality (contains actual file contents)

### Data Retention

Stored locally on disk, no cloud retention policy. Files persist until manually deleted (though future Claude Code updates could change storage format/location).

### Privacy Note

These are unencrypted plain text files. File snapshots contain actual file contents. Be mindful of what's stored if on shared machines or cloud backups.

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

2. **Find sessions by date (excludes subagents):**
   \`\`\`bash
   find ~/.claude/projects -name "*.jsonl" -type f ! -path "*/subagents/*" -newermt "2026-01-10" -exec ls -la {} \\;
   \`\`\`

3. **Search for keywords across all sessions:**
   \`\`\`bash
   grep -r "keyword" ~/.claude/projects/ --include="*.jsonl"
   \`\`\`

4. **Extract user messages from a session:**
   \`\`\`bash
   cat [session-file].jsonl | jq -s '.[] | select(.type == "user") | {time: .timestamp, content: .message.content}'
   \`\`\`

5. **Count records by type in a session:**
   \`\`\`bash
   cat [session-file].jsonl | jq -s 'group_by(.type) | .[] | {type: .[0].type, count: length}'
   \`\`\`

### Use Cases

- Recall specific decisions or explanations from past sessions
- Find code snippets or solutions discussed previously
- Track patterns in what topics you work on
- Search for that "thing we talked about" without remembering exactly when
- Analyze tool usage patterns and MCP call frequency`}</code></pre>
</ArticleLayout>
