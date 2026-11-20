<script lang="ts">
	interface Lesson {
		slug: string;
		title: string;
	}

	interface Section {
		title: string;
		lessons: Lesson[];
	}

	const sections: Section[] = [
		{
			title: 'Foundation',
			lessons: [
				{ slug: 'welcome', title: 'Welcome' },
				{ slug: 'setup', title: 'Setup Guide' },
				{ slug: 'what-we-build', title: "What We'll Build" }
			]
		},
		{
			title: 'Architecture',
			lessons: [{ slug: 'lesson-01', title: 'Lesson 1: Interfaces and Factory Functions' }]
		},
		{
			title: 'HTTPS Communication',
			lessons: [
				{ slug: 'lesson-02', title: 'Lesson 2: HTTPS Server' },
				{ slug: 'lesson-03', title: 'Lesson 3: HTTPS Agent' },
				{ slug: 'lesson-04', title: 'Lesson 4: Run Loop with HTTPS' }
			]
		},
		{
			title: 'DNS Communication',
			lessons: [
				{ slug: 'lesson-05', title: 'Lesson 5: DNS Server' },
				{ slug: 'lesson-06', title: 'Lesson 6: DNS Agent' },
				{ slug: 'lesson-07', title: 'Lesson 7: Add DNS to Run Loop' }
			]
		},
		{
			title: 'Protocol Switching',
			lessons: [
				{ slug: 'lesson-08', title: 'Lesson 8: API Switch Client' },
				{ slug: 'lesson-09', title: 'Lesson 9: Dual Server Startup' },
				{ slug: 'lesson-10', title: 'Lesson 10: Protocol Transition' }
			]
		},
		{
			title: 'Security Layer',
			lessons: [
				{ slug: 'lesson-11', title: 'Lesson 11: HMAC Authentication' },
				{ slug: 'lesson-12', title: 'Lesson 12: Payload Encryption' }
			]
		},
		{
			title: 'Command Infrastructure',
			lessons: [
				{ slug: 'lesson-13', title: 'Lesson 13: Command Endpoint' },
				{ slug: 'lesson-14', title: 'Lesson 14: Command Validation and Processing' },
				{ slug: 'lesson-15', title: 'Lesson 15: Queue Commands' },
				{ slug: 'lesson-16', title: 'Lesson 16: Dequeue and Send Commands' }
			]
		},
		{
			title: 'Agent Execution Framework',
			lessons: [
				{ slug: 'lesson-17', title: 'Lesson 17: Agent Execution Framework' },
				{ slug: 'lesson-18', title: 'Lesson 18: Shellcode Orchestrator' },
				{ slug: 'lesson-19', title: 'Lesson 19: Shellcode Doer Interface' },
				{ slug: 'lesson-20', title: 'Lesson 20: Windows Shellcode Doer' },
				{ slug: 'lesson-21', title: 'Lesson 21: Server Receives Results' }
			]
		},
		{
			title: 'Extending the Framework',
			lessons: [
				{ slug: 'lesson-22', title: 'Lesson 22: Download Command' },
				{ slug: 'lesson-23', title: 'Lesson 23: Persistence' }
			]
		},
		{
			title: 'Wrap-Up',
			lessons: [
				{ slug: 'review', title: 'Review' },
				{ slug: 'where-to-from-here', title: 'Where to From Here' },
				{ slug: 'final-thoughts', title: 'Final Thoughts' }
			]
		}
	];

	const pad = (n: number) => String(n).padStart(2, '0');

	/** "Lesson 5: DNS Server" -> ["05", "DNS Server"]; others get a quiet marker. */
	function split(title: string): [string, string] {
		const m = title.match(/^Lesson (\d+):\s*(.*)$/);
		return m ? [pad(Number(m[1])), m[2]] : ['··', title];
	}
</script>

<svelte:head>
	<title>Building a C2 Framework in Go · Course Archive</title>
	<meta
		name="description"
		content="Course archive for Building a C2 Framework in Go: multi-protocol communication, protocol switching, encrypted channels and shellcode execution."
	/>
</svelte:head>

<main class="archive">
	<nav class="crumbs" aria-label="Breadcrumb">
		<a href="/">faanross</a><span class="slash">/</span><span>courses</span><span class="slash">/</span><span
			class="here">c2-framework-go</span
		>
	</nav>

	<header class="head">
		<p class="kicker">course archive <span class="dot">·</span> antisyphon training</p>
		<h1>Building a C2 Framework in Go</h1>
		<p class="lead">
			Build a fully functional Command and Control framework from scratch. Multi-protocol communication,
			dynamic protocol switching, encrypted channels, and shellcode execution.
		</p>
		<p class="meta">
			<span>january 2026</span><span class="dot">·</span><span>16 hours</span><span class="dot">·</span><a
				href="https://github.com/faanross/antisyphon_course_c2_golang"
				target="_blank"
				rel="noopener noreferrer">solutions on github ↗</a
			>
		</p>
	</header>

	<section class="overview">
		<h2 class="label">overview</h2>
		<p>
			Over 16 hours you'll build a C2 framework that communicates over HTTPS and DNS, switches protocols on
			operator command, executes shellcode via reflective DLL loading, downloads files from target systems,
			persists through reboots, and authenticates and encrypts all of its traffic. By the end you'll have a
			working C2 system and a deep understanding of the Go patterns that make it possible.
		</p>
	</section>

	<ol class="toc">
		{#each sections as section, i}
			<li class="part">
				<h2 class="label"><span class="idx">{pad(i)}</span>{section.title.toLowerCase()}</h2>
				<ul>
					{#each section.lessons as lesson}
						{@const [num, name] = split(lesson.title)}
						<li>
							<a href="/courses/course01/{lesson.slug}">
								<span class="num">{num}</span>
								<span class="name">{name}</span>
								<span class="go" aria-hidden="true">→</span>
							</a>
						</li>
					{/each}
				</ul>
			</li>
		{/each}
	</ol>

	<footer class="foot">
		<a href="/">← faanross.com</a>
		<span class="dot">·</span>
		<a href="https://aionsec.ai">aionsec.ai</a>
	</footer>
</main>

<style>
	.archive {
		max-width: 46rem;
		margin: 0 auto;
		padding: 56px 24px 96px;
	}

	.crumbs {
		font-size: 12px;
		color: var(--fg-faint);
		margin-bottom: 72px;
	}

	.crumbs a {
		color: var(--fg-dim);
		text-decoration-color: var(--line);
	}

	.crumbs a:hover {
		color: var(--fg);
		text-decoration-color: var(--purple);
	}

	.slash {
		margin: 0 0.6ch;
		opacity: 0.7;
	}

	.here {
		color: var(--fg-dim);
	}

	.kicker {
		font-size: 12px;
		color: var(--purple);
		opacity: 0.85;
		margin-bottom: 14px;
	}

	.dot {
		margin: 0 0.8ch;
		color: var(--fg-faint);
	}

	.head h1 {
		font-size: clamp(22px, 3.2vw, 28px);
		font-weight: 400;
		line-height: 1.3;
		letter-spacing: -0.01em;
		margin-bottom: 18px;
	}

	.lead {
		color: var(--fg-dim);
		max-width: 38rem;
		margin-bottom: 18px;
	}

	.meta {
		font-size: 12px;
		color: var(--fg-faint);
	}

	.meta a {
		color: var(--yellow);
		text-decoration-color: rgba(241, 250, 140, 0.35);
	}

	.overview {
		margin-top: 56px;
		padding-top: 28px;
		border-top: 1px dashed var(--line);
	}

	.overview p {
		color: var(--fg-dim);
		font-size: 14px;
		line-height: 1.8;
	}

	.label {
		font-size: 12px;
		font-weight: 400;
		color: var(--fg-faint);
		margin-bottom: 14px;
	}

	.toc {
		list-style: none;
		margin-top: 56px;
	}

	.part {
		padding-top: 28px;
		border-top: 1px dashed var(--line);
		margin-bottom: 36px;
	}

	.idx {
		color: var(--purple);
		margin-right: 1.5ch;
	}

	.part ul {
		list-style: none;
	}

	.part a {
		display: grid;
		grid-template-columns: 4ch 1fr auto;
		align-items: baseline;
		gap: 0 1ch;
		padding: 7px 10px;
		margin: 0 -10px;
		border-radius: 4px;
		text-decoration: none;
		color: var(--fg-dim);
		font-size: 14px;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.num {
		color: var(--fg-faint);
		font-variant-numeric: tabular-nums;
		transition: color 0.15s ease;
	}

	.go {
		color: var(--yellow);
		opacity: 0;
		transform: translateX(-4px);
		transition:
			opacity 0.15s ease,
			transform 0.15s ease;
	}

	.part a:hover,
	.part a:focus-visible {
		color: var(--fg);
		background: rgba(236, 235, 230, 0.04);
	}

	.part a:hover .num,
	.part a:focus-visible .num {
		color: var(--yellow);
	}

	.part a:hover .go,
	.part a:focus-visible .go {
		opacity: 1;
		transform: none;
	}

	.foot {
		margin-top: 72px;
		padding-top: 24px;
		border-top: 1px dashed var(--line);
		font-size: 12px;
		color: var(--fg-faint);
	}

	.foot a {
		color: var(--fg-dim);
		text-decoration-color: var(--line);
	}

	.foot a:hover {
		color: var(--fg);
		text-decoration-color: var(--purple);
	}

	@media (max-width: 640px) {
		.archive {
			padding: 32px 20px 72px;
		}

		.crumbs {
			margin-bottom: 48px;
		}

		.part a {
			font-size: 13px;
		}
	}
</style>
