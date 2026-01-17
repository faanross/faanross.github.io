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
	<title>I Taught Claude to Text Me When It Needs Help | Faan Rossouw</title>
	<meta name="description" content="Setting up async notifications and bidirectional control for Claude Code via Telegram." />
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
				<span class="date">2025-01-10</span>
				<h1>I Taught Claude to Text Me When It Needs Help</h1>
			</header>

			<div class="article-content" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				<p>I've been using LLMs since day one. ChatGPT 3.5, the whole journey. Always saw the potential, found genuine use cases, stayed curious about where it was heading.</p>

				<p>But when I started using Claude Code a few weeks ago, something shifted. And once I integrated it with Obsidian into my daily workflow, I stopped seeing potential and started seeing what's coming.</p>

				<p>But there was a problem I kept ignoring.</p>

				<p>Every time Claude needed permission to run a command, or had a question, or hit any kind of impasse - it would just sit there. Waiting. And I'd be out walking my dog or handling something else entirely, while back at my desk the work had ground to a halt over a simple yes/no question.</p>

				<p>One solution is to be more liberal with permissions - pre-approve non-destructive commands so Claude doesn't have to ask. But sometimes that's just not enough.</p>

				<p>This morning I finally got annoyed enough to fix it properly.</p>

				<p>This is the full story - every wrong turn, every "why isn't this working" moment, and the final setup that actually works. If you're going to build this yourself, you deserve the real version. Not the sanitized tutorial. The messy one.</p>

				<figure class="article-image">
					<img src="/images/claude/iphone-notifications/001.png" alt="Claude Code and human collaboration" />
				</figure>

				<hr />

				<h2>The actual problem</h2>

				<p>AI assistants like Claude Code are synchronous by default. You give it a task, it works on it, then it hits something - a permission request, a clarifying question, a decision point - and stops. Waiting for you.</p>

				<p>If you're at your desk, fine. You respond, work continues.</p>

				<p>But if you're not? If you kicked off a task and walked away to stretch? You come back an hour later to find Claude blocked for 58 minutes because it wanted permission to run a simple command.</p>

				<p>That's the human bottleneck. And it breaks the promise of async AI assistance.</p>

				<figure class="article-image">
					<img src="/images/claude/iphone-notifications/003.png" alt="Sync vs Async - wizard stuck at desk versus wizard relaxed on couch" />
				</figure>

				<p>The obvious fix is notifications. Get a ping on your phone when Claude needs you.</p>

				<p>But I wanted more than that. I wanted to <em>respond</em> from my phone. Approve the permission. Answer the question. Keep work flowing without walking back to my computer.</p>

				<p>That's where things got complicated.</p>

				<hr />

				<h2>What options exist</h2>

				<p>Before building anything, I mapped out the landscape.</p>

				<p><strong>One-way notifications</strong> are the simplest path. ntfy.sh is free, open-source, and takes five minutes to set up. Subscribe to a topic, add a Claude hook that curls it, done. Your phone buzzes when Claude needs permission. You walk back to your computer to respond.</p>

				<p>Pushcut does the same thing with Apple Watch support. macOS native notifications work too, but only if you're already at your Mac. This also has its use - hence why I've already integrated that - but here I was trying to scratch a different itch.</p>

				<p><strong>SSH from your phone</strong> is the power-user approach. Termius, Blink, or Prompt let you connect to your Mac, find your tmux session, interact directly. Full control.</p>

				<p>The problem? Typing on a phone keyboard in a terminal is painful. Navigating to find the right session is friction. And there are no push notifications - you have to manually check.</p>

				<p><strong>Bidirectional messaging</strong> is the sweet spot. Get notified AND respond, all from a messaging app. Claude-Code-Remote with Telegram creates a webhook server that bridges the two. You get a message with a token, reply with your command, it gets injected into your session.</p>

				<p>That's what I chose.</p>

				<hr />

				<h2>Why Telegram</h2>

				<p>My requirements were specific:</p>

				<ul>
					<li>Start a task, walk away, get notified when blocked</li>
					<li>Reply from my phone without returning to computer</li>
					<li>Only notify when Claude is BLOCKED, not every time it completes (though this can also be added with an additional hook if you wanted)</li>
					<li>Native texting UX, not a tiny terminal keyboard</li>
				</ul>

				<p>I already use Telegram. The bot API is simple. Webhooks work. It's free.</p>

				<p>The trade-off is complexity. ntfy takes five minutes. This took over two hours, which included a whole lot of debugging. But now I can reply to Claude the same way I reply to a friend - by texting.</p>

				<p>One note: this guide is iOS/macOS-focused. If you're on Android, the core architecture is identical, but you'd use different apps - Pushover instead of Pushcut, Tasker instead of Shortcuts, JuiceSSH for terminal access.</p>

				<hr />

				<h2>How it actually works</h2>

				<p>Here's the data flow:</p>

				<figure class="article-image">
					<img src="/images/claude/iphone-notifications/002.png" alt="Data flow diagram showing Claude to Telegram to phone" />
				</figure>

				<p>Claude gets blocked → Claude hook fires → Sends Telegram message with token → Your phone buzzes</p>

				<p>You reply: <code>/cmd TOKEN123 yes please continue</code> → Telegram sends to webhook server → Server validates token → Server runs <code>tmux send-keys</code> → Claude receives your response → Work continues</p>

				<p>The insight that took me embarrassingly long to figure out:</p>

				<p><strong>Notifications don't need tmux. Responses do.</strong></p>

				<p>Sending a notification is trivial. Any script can hit the Telegram API. But receiving a response and injecting it into Claude? That requires a way to programmatically type into your terminal session.</p>

				<p>That's what tmux provides. The webhook server uses <code>tmux send-keys</code> to inject keystrokes. Without tmux, you can get notifications, but you can't respond remotely.</p>

				<hr />

				<h2>The setup</h2>

				<p>Here's what you need:</p>

				<ul>
					<li>Node.js</li>
					<li>tmux (for bidirectional communication)</li>
					<li>ngrok (to expose your local webhook to the internet)</li>
					<li>A Telegram bot (free, two minutes to create)</li>
					<li>Claude-Code-Remote (the project that ties it together)</li>
				</ul>

				<p><strong>Install the basics:</strong></p>

				<pre><code>brew install ngrok tmux</code></pre>

				<p>Sign up for a free ngrok account and add your authtoken:</p>

				<pre><code>ngrok config add-authtoken YOUR_TOKEN</code></pre>

				<p><strong>Fix tmux mouse behavior:</strong></p>

				<p>This isn't in any tutorial. tmux breaks macOS mouse defaults out of the box - scrolling doesn't work right, copy-paste is broken.</p>

				<p>Create <code>~/.tmux.conf</code>:</p>

				<pre><code>set -g mouse on
unbind -T copy-mode MouseDragEnd1Pane
bind-key -T copy-mode MouseDragEnd1Pane send-keys -X copy-pipe-and-cancel "pbcopy"
unbind -T copy-mode-vi MouseDragEnd1Pane
bind-key -T copy-mode-vi MouseDragEnd1Pane send-keys -X copy-pipe-and-cancel "pbcopy"</code></pre>

				<p>Now scrolling works and select-drag-release copies to clipboard automatically (bonus win imo).</p>

				<p><strong>Create a Telegram bot:</strong></p>

				<p>Open Telegram, search for <code>@BotFather</code>, send <code>/newbot</code>, give it a name and username (must end in <code>bot</code>), save the API token.</p>

				<p>Get your Chat ID by messaging your bot, then visiting:</p>

				<pre><code>https://api.telegram.org/bot&lt;YOUR_TOKEN&gt;/getUpdates</code></pre>

				<p>Find <code>"chat":{"{"}id":123456789{"}"}</code> - that number is your Chat ID.</p>

				<p><strong>Clone and configure Claude-Code-Remote:</strong></p>

				<pre><code>cd ~/tools
git clone https://github.com/JessyTsui/Claude-Code-Remote.git
cd Claude-Code-Remote
npm install</code></pre>

				<p>Now create <code>.env</code>:</p>

				<pre><code>TELEGRAM_ENABLED=true
TELEGRAM_BOT_TOKEN=your-token-here
TELEGRAM_CHAT_ID=your-chat-id-here
INJECTION_MODE=tmux</code></pre>

				<p><strong>Add the tmux wrapper to your shell:</strong></p>

				<p>In <code>~/.zshrc</code>:</p>

				<pre><code>{`claude() {
    if [[ -n "$TMUX" ]]; then
        command claude "$@"
    else
        local session_name="claude-$(date +%s)"
        export CLAUDE_TMUX_SESSION="$session_name"
        tmux new-session -s "$session_name" \\
            "export CLAUDE_TMUX_SESSION=$session_name; command claude $*; exec zsh"
    fi
}`}</code></pre>

				<p>If you're using bash, add this to <code>~/.bashrc</code> instead and change <code>exec zsh</code> to <code>exec bash</code> on the last line.</p>

				<p>Now when you type <code>claude</code> in your terminal, it doesn't just open Claude - it wraps it in a tmux session with a unique timestamp-based name.</p>

				<p><strong>Configure Claude hooks:</strong></p>

				<p>Now that Claude is wrapped in tmux for response injection, we need to tell Claude when to send notifications. Claude Code has a hooks system - shell commands that fire on specific events. We'll use <code>PermissionRequest</code> (fires when Claude needs permission) and <code>Notification</code> with <code>idle_prompt</code> (fires when Claude is waiting for input).</p>

				<p>In <code>~/.claude/settings.json</code>:</p>

				<pre><code>{`{
  "hooks": {
    "PermissionRequest": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "node ~/tools/Claude-Code-Remote/claude-hook-permission.js",
            "timeout": 5
          }
        ]
      }
    ],
    "Notification": [
      {
        "matcher": "idle_prompt",
        "hooks": [
          {
            "type": "command",
            "command": "node ~/tools/Claude-Code-Remote/claude-hook-notify.js waiting",
            "timeout": 5
          }
        ]
      }
    ]
  }
}`}</code></pre>

				<p>Why <code>PermissionRequest</code> instead of a notification hook? Because it provides full JSON via stdin - the command Claude wants to run, the working directory, everything. Richer notifications.</p>

				<p><strong>Start the services:</strong></p>

				<pre><code>~/tools/Claude-Code-Remote/start-telegram-services.sh</code></pre>

				<p>This starts ngrok, gets the tunnel URL, registers the webhook with Telegram, and starts the server.</p>

				<hr />

				<h2>Everything that went wrong</h2>

				<p>If the setup above worked perfectly for you on the first try, you're luckier than I was.</p>

				<p><strong>"PTY session 'default' not found"</strong></p>

				<p>This error cost me 15 minutes.</p>

				<p>The cause: <code>INJECTION_MODE=pty</code> in my <code>.env</code> file. The project supports two modes - PTY and tmux. PTY mode doesn't work with Claude Code's architecture.</p>

				<p>The fix: Change to <code>INJECTION_MODE=tmux</code>.</p>

				<p>The gotcha: you have to restart the webhook server after changing <code>.env</code>. Node reads config at startup. I changed the setting, tested again, same error. Spent another few minutes confused before realizing the server was running with old config.</p>

				<p><strong>Trying to scrape the tmux screen for permission details</strong></p>

				<p>My first approach was janky screen scraping to get permission details. Fragile and ultimately pointless.</p>

				<p>The discovery: Claude's <code>PermissionRequest</code> hook receives full JSON via stdin. Just read from stdin. No scraping needed.</p>

				<pre><code>{`{
  "tool_name": "Bash",
  "tool_input": {"command": "rm -rf /tmp/test"},
  "cwd": "/Users/me/project"
}`}</code></pre>

				<p><strong>Parallel Claude sessions conflicting</strong></p>

				<p>My original tmux wrapper used the <code>-A</code> flag, which attaches to an existing session if one exists. Only one Claude session could run at a time.</p>

				<p>The fix: Timestamp-based unique session names. Every <code>claude</code> invocation gets its own session.</p>

				<p><strong>ngrok URL changes on restart</strong></p>

				<p>Free tier ngrok gives you a random URL every restart. Your Telegram webhook registration becomes stale.</p>

				<p>The startup script handles this automatically - gets the new URL, re-registers with Telegram. But if you're debugging manually, this will bite you.</p>

				<p>If the random URL annoys you, alternatives exist:</p>
				<ul>
					<li>ngrok paid ($8/month): Static subdomain</li>
					<li>Cloudflare Tunnel (free): Static URL if you have a domain on Cloudflare</li>
					<li>Tailscale Funnel (free): Static URL if you use Tailscale</li>
				</ul>

				<p>I'm sticking with ngrok free. The startup script handles the URL dance.</p>

				<p><strong>Notifications sent but phone never buzzed</strong></p>

				<p>Everything configured correctly. Telegram API confirmed message sent. Phone silent.</p>

				<p>The cause: iOS Focus mode was blocking Telegram notifications. No error anywhere. Just silence.</p>

				<p>The fix: Add Telegram to "Allowed Apps" in Focus settings.</p>

				<hr />

				<h2>What I specifically didn't want</h2>

				<p>Important clarification: I did NOT want completion notifications.</p>

				<p>Many people set up notifications for when Claude finishes a task. Every turn completion, ping.</p>

				<p>That would drive me insane.</p>

				<p>I wanted impasse notifications. I want to know when Claude is blocked - needs permission, waiting for input, can't continue without me. I don't need a ping every time it successfully writes a file.</p>

				<p>The signal is "Claude needs you." Not "Claude is done."</p>

				<p>This affects your hook configuration. Completion notifications use a <code>Stop</code> hook. Impasse notifications use <code>PermissionRequest</code> and <code>Notification</code> with <code>idle_prompt</code>.</p>

				<hr />

				<h2>The daily reality</h2>

				<p>Here's what my day looks like now.</p>

				<p>I start Claude. The wrapper creates a tmux session automatically. First permission prompt triggers a Telegram notification.</p>

				<p>I give Claude a task and walk away. Stretch. Take my dog out.</p>

				<p>Phone buzzes: "Claude needs permission: Bash(npm install bcrypt)"</p>

				<p>I reply: <code>/cmd XK7M2P yes</code></p>

				<p>Claude continues. I don't stand up.</p>

				<figure class="article-image">
					<img src="/images/claude/iphone-notifications/screenshot.jpeg" alt="Telegram notification on phone" />
				</figure>

				<p>One more thing: the last thing I want is to walk away from my desk and find out the tunnel or webhook server was down the whole time. So I added an auto-check to my project's <code>CLAUDE.md</code> file. Every time Claude starts a session, it checks if ngrok and the webhook server are running. If either is down, it offers to start them. And since free ngrok gives you a new URL every restart, it automatically re-registers the webhook with Telegram using the new URL.</p>

				<p>This means I never have to think about it. Start Claude, it handles the plumbing. When I leave my desk, I know Telegram is connected.</p>

				<p>NOTE: I'm sharing that with you below, just C+P into your root-level CLAUDE.md.</p>

				<hr />

				<h2>What I learned</h2>

				<p><strong>Notifications and responses are different problems.</strong> Sending a ping is trivial. Injecting a response requires terminal access. Completely different architectures.</p>

				<p><strong>Config details matter more than you think.</strong> <code>pty</code> vs <code>tmux</code> is one word. It determines whether everything works or fails cryptically.</p>

				<p><strong>Restart after config changes.</strong> Node reads <code>.env</code> at startup. Change the file, restart the server. I lost 20 minutes to this.</p>

				<p><strong>Mobile OS is part of your system.</strong> Focus mode, notification settings, background app limits - all can silently break your setup with no error anywhere.</p>

				<p><strong>Test each layer independently.</strong> Telegram API, ngrok tunnel, webhook server, hook scripts, tmux injection. Isolate failures by testing components separately.</p>

				<p><strong>Unique session names for parallel work.</strong> If you want multiple Claude instances, each needs its own tmux session.</p>

				<hr />

				<h2>The bigger picture</h2>

				<p>This fix took two hours. But it wasn't really about notifications.</p>

				<p>My mantra has become simple: reduce friction to the minimum required to fully manifest an idea. Every unnecessary step between thought and creation is a leak in the system.</p>

				<figure class="article-image">
					<img src="/images/claude/iphone-notifications/004.png" alt="Optimized flow versus congestion - wizard channeling streamlined workflow" />
				</figure>

				<p>That's what this was about. The human bottleneck - me sitting at my desk waiting for Claude to need me - was friction. Now it's gone. I can start a task, walk away, and stay in the loop from my phone.</p>

				<p>This is the first article in what I expect to be a longer journey. There's more friction to eliminate, more workflows to optimize, more ways to make the collaboration between human and AI feel less like tool usage and more like partnership.</p>

				<p>If that resonates, follow along.</p>

				<hr />

				<h2>Bonus: The CLAUDE.md Auto-Check</h2>

				<p>Here's the logic I added to my project's <code>CLAUDE.md</code> file. Claude reads this on session start and runs the health check automatically:</p>

				<pre><code>{`## Session Startup Checks

**On your FIRST response in any new session, BEFORE addressing the user's request, run these checks:**

### Telegram Remote Services Health Check

**Run these commands:**

pgrep -x ngrok > /dev/null 2>&1 && echo "ngrok:UP" || echo "ngrok:DOWN"

lsof -i :3001 > /dev/null 2>&1 && echo "webhook:UP" || echo "webhook:DOWN"

**Then act based on results:**

| ngrok | webhook | Action |
|-------|---------|--------|
| UP | UP | Say nothing, proceed with user's request |
| DOWN | * | Ask: "Telegram remote not running. Start services? (y/n)" |
| * | DOWN | Ask: "Telegram remote not running. Start services? (y/n)" |

**If user says yes:** Run:

~/tools/Claude-Code-Remote/start-telegram-services.sh`}</code></pre>

				<p>Now I never think about it. Claude checks, offers to start if needed, and handles the ngrok URL dance automatically.</p>

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
	}
</style>
