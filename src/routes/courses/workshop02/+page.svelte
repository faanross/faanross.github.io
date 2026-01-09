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
				{ slug: "setup", title: "Setup Guide" },
				{ slug: "interfaces", title: "Project Structure and Interfaces" },
				{ slug: "yaml", title: "YAML-based Configuration Management System" }
			]
		},
		{
			title: "HTTPS Implementation",
			lessons: [
				{ slug: "https-server", title: "HTTPS Server" },
				{ slug: "https-agent", title: "HTTPS Agent" },
				{ slug: "https-loop", title: "HTTPS Loop" }
			]
		},
		{
			title: "DNS Implementation",
			lessons: [
				{ slug: "dns-server", title: "DNS Server" },
				{ slug: "dns-agent", title: "DNS Agent" },
				{ slug: "dns-loop", title: "DNS Loop" }
			]
		},
		{
			title: "Transition Using API Switch",
			lessons: [
				{ slug: "api", title: "Implement API Switch" },
				{ slug: "dual", title: "Dual-server Startup" },
				{ slug: "transition", title: "Agent Parsing + Protocol Transition" }
			]
		}
	];
</script>

<svelte:head>
	<title>Let's Build a Multi-Modal C2 Covert Channel in Golang | Faan Rossouw</title>
	<meta name="description" content="Build a C2 that can switch between HTTPS and DNS protocols on the fly." />
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
			<span class="badge" in:fly={{ y: 20, duration: 600, delay: 200 }}>AntiSyphon Workshop</span>
			<h1 in:fly={{ y: 20, duration: 600, delay: 300 }}>Let's Build a Multi-Modal C2 Covert Channel in Golang</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Build a C2 that can switch between HTTPS and DNS protocols on the fly. Learn protocol abstraction, interface-based design, and how to implement resilient multi-channel communications.
			</p>
			<div class="meta" in:fly={{ y: 20, duration: 600, delay: 500 }}>
				<span class="date">September 18, 2025</span>
				<span class="separator">|</span>
				<span class="duration">4 hours</span>
			</div>
			<div class="actions" in:fly={{ y: 20, duration: 600, delay: 600 }}>
				<a href="https://github.com/faanross/workshop_antisyphon_18092025" target="_blank" rel="noopener noreferrer" class="btn-secondary">
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
					Below are the lecture notes for my AntiSyphon workshop presented on September 18, 2025. Though the notes are in
					general more descriptive than the actual lectures, they are not expanded, meaning the content from the lectures roughly map
					1:1 onto these notes.
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
								<a href="/courses/workshop02/{lesson.slug}">{lesson.title}</a>
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
		color: var(--aion-purple);
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
