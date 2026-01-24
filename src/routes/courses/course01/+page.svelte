<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});

	const sections = [
		{
			title: "Foundation",
			lessons: [
				{ slug: "welcome", title: "Welcome" },
				{ slug: "setup", title: "Setup Guide" },
				{ slug: "what-we-build", title: "What We'll Build" }
			]
		},
		{
			title: "Part 1: Architecture",
			lessons: [
				{ slug: "lesson-01", title: "Lesson 1: Interfaces and Factory Functions" }
			]
		},
		{
			title: "Part 2: HTTPS Communication",
			lessons: [
				{ slug: "lesson-02", title: "Lesson 2: HTTPS Server" },
				{ slug: "lesson-03", title: "Lesson 3: HTTPS Agent" },
				{ slug: "lesson-04", title: "Lesson 4: Run Loop with HTTPS" }
			]
		},
		{
			title: "Part 3: DNS Communication",
			lessons: [
				{ slug: "lesson-05", title: "Lesson 5: DNS Server" },
				{ slug: "lesson-06", title: "Lesson 6: DNS Agent" },
				{ slug: "lesson-07", title: "Lesson 7: Add DNS to Run Loop" }
			]
		},
		{
			title: "Part 4: Protocol Switching",
			lessons: [
				{ slug: "lesson-08", title: "Lesson 8: API Switch Client" },
				{ slug: "lesson-09", title: "Lesson 9: Dual Server Startup" },
				{ slug: "lesson-10", title: "Lesson 10: Protocol Transition" }
			]
		},
		{
			title: "Part 5: Security Layer",
			lessons: [
				{ slug: "lesson-11", title: "Lesson 11: HMAC Authentication" },
				{ slug: "lesson-12", title: "Lesson 12: Payload Encryption" }
			]
		},
		{
			title: "Part 6: Command Infrastructure",
			lessons: [
				{ slug: "lesson-13", title: "Lesson 13: Command Endpoint" },
				{ slug: "lesson-14", title: "Lesson 14: Command Validation and Processing" },
				{ slug: "lesson-15", title: "Lesson 15: Queue Commands" },
				{ slug: "lesson-16", title: "Lesson 16: Dequeue and Send Commands" }
			]
		},
		{
			title: "Part 7: Agent Execution Framework",
			lessons: [
				{ slug: "lesson-17", title: "Lesson 17: Agent Execution Framework" },
				{ slug: "lesson-18", title: "Lesson 18: Shellcode Orchestrator" },
				{ slug: "lesson-19", title: "Lesson 19: Shellcode Doer Interface" },
				{ slug: "lesson-20", title: "Lesson 20: Windows Shellcode Doer" },
				{ slug: "lesson-21", title: "Lesson 21: Server Receives Results" }
			]
		},
		{
			title: "Part 8: Extending the Framework",
			lessons: [
				{ slug: "lesson-22", title: "Lesson 22: Download Command" },
				{ slug: "lesson-23", title: "Lesson 23: Persistence" }
			]
		},
		{
			title: "Wrap-Up",
			lessons: [
				{ slug: "review", title: "Review" },
				{ slug: "where-to-from-here", title: "Where to From Here" },
				{ slug: "final-thoughts", title: "Final Thoughts" }
			]
		}
	];
</script>

<svelte:head>
	<title>Building a C2 Framework in Go | Faan Rossouw</title>
	<meta name="description" content="A comprehensive 16-hour course on building a Command and Control framework from scratch using Go." />
</svelte:head>

<section class="course-hero">
	<div class="container">
		{#if mounted}
			<a href="/courses" class="back-link" in:fly={{ y: -10, duration: 400, delay: 100 }}>
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<polyline points="15 18 9 12 15 6"></polyline>
				</svg>
				Back to Courses
			</a>
			<span class="badge badge-blue" in:fly={{ y: 20, duration: 600, delay: 200 }}>AntiSyphon Course</span>
			<h1 in:fly={{ y: 20, duration: 600, delay: 300 }}>Building a C2 Framework in Go</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Build a fully functional Command and Control framework from scratch. Learn multi-protocol communication, dynamic protocol switching, encrypted channels, and shellcode execution.
			</p>
			<div class="meta" in:fly={{ y: 20, duration: 600, delay: 500 }}>
				<span class="date">January 2026</span>
				<span class="separator">|</span>
				<span class="duration">16 hours</span>
			</div>
			<div class="actions" in:fly={{ y: 20, duration: 600, delay: 600 }}>
				<a href="https://github.com/faanross/antisyphon_course_c2_golang" target="_blank" rel="noopener noreferrer" class="btn-secondary">
					<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
						<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
					</svg>
					View Solutions on GitHub
				</a>
			</div>
		{/if}
	</div>
</section>

<section class="course-overview">
	<div class="container">
		{#if mounted}
			<div class="overview-card glass-card" in:fly={{ y: 30, duration: 600, delay: 700 }}>
				<h2>Overview</h2>
				<p>
					In this comprehensive 16-hour course, you'll build a fully functional C2 framework that can communicate over multiple protocols (HTTPS and DNS), switch protocols dynamically on operator command, execute shellcode via reflective DLL loading, download files from target systems, persist through system reboots, and authenticate and encrypt all communications. By the end, you'll have a working C2 system and deep understanding of the Go patterns that make it possible.
				</p>
			</div>
		{/if}
	</div>
</section>

<section class="course-toc">
	<div class="container">
		{#if mounted}
			{#each sections as section, sectionIndex}
				<div class="toc-section" in:fly={{ y: 30, duration: 600, delay: 800 + sectionIndex * 100 }}>
					<h3>{section.title}</h3>
					<ul>
						{#each section.lessons as lesson}
							<li>
								<a href="/courses/course01/{lesson.slug}">{lesson.title}</a>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		{/if}
	</div>
</section>

<style>
	.course-hero {
		padding: 80px 0 40px;
		text-align: center;
	}

	.back-link {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: rgba(255, 255, 255, 0.6);
		text-decoration: none;
		margin-bottom: 24px;
		transition: color 0.3s ease;
	}

	.back-link:hover {
		color: var(--aion-purple);
	}

	.badge {
		display: inline-block;
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		padding: 4px 12px;
		background: rgba(189, 147, 249, 0.2);
		border: 1px solid rgba(189, 147, 249, 0.4);
		border-radius: 12px;
		color: var(--aion-purple-light);
		margin-bottom: 16px;
	}

	.badge-blue {
		background: rgba(99, 179, 237, 0.2);
		border-color: rgba(99, 179, 237, 0.4);
		color: #63b3ed;
	}

	.course-hero h1 {
		font-size: clamp(28px, 4vw, 42px);
		font-weight: 700;
		margin-bottom: 16px;
		max-width: 800px;
		margin-left: auto;
		margin-right: auto;
	}

	.lead {
		font-size: clamp(16px, 2vw, 18px);
		color: rgba(255, 255, 255, 0.7);
		max-width: 600px;
		margin: 0 auto 20px;
	}

	.meta {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.5);
		margin-bottom: 24px;
	}

	.separator {
		margin: 0 12px;
	}

	.date {
		color: #63b3ed;
	}

	.actions {
		display: flex;
		justify-content: center;
		gap: 12px;
	}

	.btn-secondary {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 12px 20px;
		background: transparent;
		border: 1px solid rgba(189, 147, 249, 0.4);
		border-radius: 20px;
		color: var(--aion-purple-light);
		font-size: 14px;
		font-weight: 600;
		text-decoration: none;
		transition: all 0.3s ease;
	}

	.btn-secondary:hover {
		background: rgba(189, 147, 249, 0.1);
		border-color: rgba(189, 147, 249, 0.6);
	}

	.course-overview {
		padding: 20px 0 40px;
	}

	.overview-card {
		max-width: 800px;
		margin: 0 auto;
		padding: 32px;
	}

	.overview-card h2 {
		font-size: 20px;
		margin-bottom: 12px;
	}

	.overview-card p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.7);
		line-height: 1.7;
	}

	.course-toc {
		padding: 20px 0 80px;
		max-width: 800px;
		margin: 0 auto;
	}

	.toc-section {
		margin-bottom: 32px;
	}

	.toc-section h3 {
		font-size: 18px;
		font-weight: 600;
		color: var(--aion-purple);
		margin-bottom: 12px;
		padding-bottom: 8px;
		border-bottom: 1px solid rgba(189, 147, 249, 0.2);
	}

	.toc-section ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.toc-section li {
		margin-bottom: 8px;
	}

	.toc-section a {
		display: block;
		padding: 12px 16px;
		background: rgba(61, 61, 71, 0.3);
		border: 1px solid rgba(189, 147, 249, 0.1);
		border-radius: 8px;
		color: rgba(255, 255, 255, 0.8);
		text-decoration: none;
		font-size: 14px;
		transition: all 0.3s ease;
	}

	.toc-section a:hover {
		background: rgba(61, 61, 71, 0.5);
		border-color: rgba(189, 147, 249, 0.3);
		color: var(--white);
		transform: translateX(4px);
	}

	@media (max-width: 768px) {
		.course-hero {
			padding: 60px 0 30px;
		}

		.overview-card {
			padding: 24px;
		}

		.course-toc {
			padding: 20px 24px 60px;
		}
	}
</style>
