<script lang="ts">
	import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
	title="The Bug That Proved the Point"
	date="2026-02-04"
	description="My Zeek parser looked right. The tests said otherwise. How verification catches what specifications miss."
>
	<figure class="article-image">
		<img src="/images/claude/verification-catches-misinterpretations/IMG-001-HERO.png" alt="Verification catches misinterpretations" />
	</figure>

	<p>
		I was building a Zeek log parser with Claude Code to harmonize field names to a universal schema. The prompt was clear—or so I thought:
	</p>

	<pre><code>- conn.log: Map id.orig_h→src_ip, id.orig_p→src_port, id.resp_h→dest_ip,
  id.resp_p→dest_port, proto→protocol, orig_bytes→bytes_sent,</code></pre>

	<p>Claude generated the parser. It looked reasonable. I could have moved on.</p>

	<p>Instead, I ran the unit tests against sample Zeek data I'd prepared earlier.</p>

	<pre><code>TestParseZeekConn: Expected SrcIP to not be empty
TestParseZeekConn: Expected DestIP to not be empty</code></pre>

	<p>Wait. The fields were empty? The parser compiled. The logic seemed right. What happened?</p>

	<hr />

	<h2>The Misinterpretation</h2>

	<figure class="article-image">
		<img src="/images/claude/verification-catches-misinterpretations/IMG-002-MISINTERPRETATION.png" alt="Two valid interpretations of the same spec" />
	</figure>

	<p>
		Claude had read my field mapping <code>id.orig_h→src_ip</code> and interpreted it as nested JSON structure: <code>id: {'{'}orig_h: ...{'}'}</code>.
	</p>

	<p>
		But Zeek's actual format uses flat dotted keys—literally the string <code>"id.orig_h"</code> as a key name, not nested objects.
	</p>

	<p>
		My specification was ambiguous. Claude made a reasonable interpretation. It just happened to be wrong.
	</p>

	<p>
		This is the nature of working with AI agents. They're remarkably capable at understanding intent and generating code. But "remarkably capable" still means they'll sometimes misunderstand, especially when your specification has any ambiguity.
	</p>

	<p>The question isn't whether misinterpretations will happen. They will.</p>

	<p>The question is whether you catch them immediately—or discover them later in production.</p>

	<hr />

	<h2>What Saved Me</h2>

	<p>Two things:</p>

	<p>
		<strong>First, I had prepared test data with the actual Zeek format.</strong> The test data knew the truth about how Zeek structures its fields.
	</p>

	<p>
		<strong>Second, I wrote unit tests before moving on.</strong> The tests compared expected output against actual output. When the parser produced empty fields instead of IP addresses, the discrepancy was immediately obvious.
	</p>

	<p>
		Without those two things, this bug would have silently broken correlation between endpoint and network telemetry in my threat hunting system. I would have discovered it much later—when investigations weren't working—and spent far more time tracing the root cause back to a parser that "looked right."
	</p>

	<hr />

	<h2>The Input/Output Framework</h2>

	<figure class="article-image">
		<img src="/images/claude/verification-catches-misinterpretations/IMG-003-FRAMEWORK.png" alt="Input, Process, Output - your power is at the bookends" />
	</figure>

	<p>Here's the insight that changes how you work with AI agents:</p>

	<p>
		<strong>When the agent handles implementation, your power moves to the bookends—the inputs you provide and the outputs you verify.</strong>
	</p>

	<p>
		Your inputs shape what gets built: specifications, reference documents, architecture decisions, test data. The clearer and more precise your inputs, the better the agent's first attempt.
	</p>

	<p>
		Your output verification catches the inevitable misinterpretations. Tests, reviews, validation against expected behavior. This is where you catch ambiguities that slipped through your inputs.
	</p>

	<p>
		The agent is remarkably good at the middle part—turning your intent into working code. But it can only work with what you give it, and it can only be as correct as your verification allows.
	</p>

	<hr />

	<h2>This Is Not Vibe Coding</h2>

	<p>
		"Vibe coding" has come to mean blindly accepting whatever an AI produces. Ignoring errors. Shipping code you don't understand. Hoping for the best.
	</p>

	<p>What I'm describing is the opposite.</p>

	<p>
		You still use AI to write code—that capability is real and valuable. But you invest in the bookends: clear inputs that reduce ambiguity, and systematic verification that catches what slips through.
	</p>

	<p>
		The Zeek parser bug took minutes to find because I'd built verification into my workflow. Without tests, it might have taken hours or days to surface, buried under layers of subsequent work.
	</p>

	<p>That's the difference between building systems that work and generating slop that appears to work.</p>

	<hr />

	<h2>The Practical Takeaway</h2>

	<figure class="article-image">
		<img src="/images/claude/verification-catches-misinterpretations/IMG-004-TAKEAWAY.png" alt="You control the bookends" />
	</figure>

	<p>When you're building with AI agents:</p>

	<p>
		<strong>Invest in your inputs.</strong> Reference documents with exact field names. Test data reflecting real-world formats. Specifications that minimize ambiguity.
	</p>

	<p>
		<strong>Invest in your outputs.</strong> Unit tests that verify behavior. Sample data that exposes incorrect assumptions. Automated checks that catch misinterpretations immediately.
	</p>

	<p>The agent handles implementation. You handle the bookends.</p>

	<p>That's how you build things that actually work.</p>
</ArticleLayout>
