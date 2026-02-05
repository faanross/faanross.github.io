<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});

	const courses = [
		{
			title: "Let's Build a Simple C2 in Golang and Vue.js",
			badge: "AntiSyphon",
			description: "4-hour hands-on workshop where we build a complete C2 infrastructure from scratch. Learn to create a server with HTTP handlers, a Golang agent, and a Vue.js client interface.",
			date: "April 25, 2025",
			duration: "4 hours",
			githubUrl: "https://github.com/faanross/workshop_antisyphon_25042025",
			courseUrl: "/courses/workshop"
		},
		{
			title: "Let's Build a Multi-Modal C2 Covert Channel in Golang",
			badge: "AntiSyphon",
			description: "Build a C2 that can switch between HTTPS and DNS protocols on the fly. Learn protocol abstraction, interface-based design, and how to implement resilient multi-channel communications.",
			date: "September 18, 2025",
			duration: "4 hours",
			githubUrl: "https://github.com/faanross/workshop_antisyphon_18092025",
			courseUrl: "/courses/workshop02"
		},
		{
			title: "Let's Build a Reflective Loader + C2 Channel in Golang",
			badge: "AntiSyphon",
			description: "Build a reflective DLL loader integrated with a C2 channel. Learn PE parsing, memory mapping, relocations, and how to execute payloads without touching disk.",
			date: "January 23, 2026",
			duration: "4 hours",
			githubUrl: "https://github.com/faanross/workshop_antisyphon_23012026",
			courseUrl: "/courses/workshop03"
		},
		{
			title: "Let's Build a Reflective Loader in Golang",
			badge: "Faan Ross",
			badgeColor: "yellow",
			description: "Comprehensive self-paced course covering DLLs, PE format, reflective loading, relocations, imports, and obfuscation techniques. Build a reflective DLL loader from the ground up.",
			date: "Self-Paced",
			duration: "8 Modules",
			courseUrl: "/courses/reflective"
		},
		{
			title: "Building a C2 Framework in Go",
			badge: "AntiSyphon",
			description: "Comprehensive 16-hour course building a complete C2 framework from scratch. Multi-protocol communication (HTTPS/DNS), dynamic protocol switching, encrypted channels, command infrastructure, and shellcode execution.",
			date: "January 2026",
			duration: "16 hours",
			githubUrl: "https://github.com/faanross/antisyphon_course_c2_golang",
			courseUrl: "/courses/course01"
		}
	];
</script>

<svelte:head>
	<title>Courses | Faan Rossouw</title>
	<meta name="description" content="Workshops and courses on C2 development, reflective loading, and malware techniques in Golang." />
</svelte:head>

<section class="courses-hero">
	<div class="container">
		{#if mounted}
			<h1 in:fly={{ y: 30, duration: 800, delay: 200 }}>Courses</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Hands-on workshops and courses on offensive tooling development and threat hunting
			</p>
		{/if}
	</div>
</section>

<section class="courses-grid-section">
	<div class="container">
		{#if mounted}
			<div class="courses-grid">
				{#each courses as course, index}
					<div class="course-card glass-card" in:fly={{ y: 30, duration: 600, delay: 500 + index * 100 }}>
						<div class="course-header">
							<span class="badge" class:badge-yellow={course.badgeColor === 'yellow'}>{course.badge}</span>
							<span class="duration">{course.duration}</span>
						</div>
						<h3>{course.title}</h3>
						<p class="description">{course.description}</p>
						<div class="course-meta">
							<span class="date">{course.date}</span>
						</div>
						<div class="course-links">
							{#if course.courseUrl}
								<a href={course.courseUrl} class="btn-primary">
									View Course
								</a>
							{/if}
							{#if course.githubUrl}
								<a href={course.githubUrl} target="_blank" rel="noopener noreferrer" class="btn-secondary">
									<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
										<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
									</svg>
									Solutions
								</a>
							{/if}
							{#if !course.courseUrl && !course.githubUrl}
								<span class="coming-soon">Coming Soon</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</section>

<section class="aionsec-cta">
	<div class="container">
		{#if mounted}
			<div class="cta-card glass-card" in:fly={{ y: 30, duration: 600, delay: 1000 }}>
				<h2>Looking for More?</h2>
				<p>AionSec offers structured courses on agentic AI for security — from building C2 frameworks to designing AI-powered workflows</p>
				<a href="https://aionsec.ai/courses" target="_blank" rel="noopener noreferrer" class="btn-primary">
					Browse AionSec Courses
				</a>
			</div>
		{/if}
	</div>
</section>

<style>
	.courses-hero {
		padding: 80px 0 60px;
		text-align: center;
	}

	.courses-hero h1 {
		margin-bottom: 16px;
	}

	.lead {
		font-size: clamp(16px, 2vw, 20px);
		color: rgba(255, 255, 255, 0.7);
		max-width: 600px;
		margin: 0 auto;
	}

	.courses-grid-section {
		padding: 0 0 60px;
	}

	.courses-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
		gap: 24px;
	}

	.course-card {
		display: flex;
		flex-direction: column;
		padding: 28px;
	}

	.course-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.badge {
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		padding: 4px 10px;
		background: rgba(189, 147, 249, 0.2);
		border: 1px solid rgba(189, 147, 249, 0.4);
		border-radius: 12px;
		color: var(--aion-purple-light);
	}

	.badge-yellow {
		background: rgba(245, 230, 99, 0.2);
		border-color: rgba(245, 230, 99, 0.4);
		color: var(--aion-yellow);
	}

	.duration {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.5);
	}

	.course-card h3 {
		font-size: 18px;
		font-weight: 600;
		color: var(--white);
		line-height: 1.4;
		margin-bottom: 12px;
	}

	.description {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.7);
		line-height: 1.6;
		flex-grow: 1;
		margin-bottom: 16px;
	}

	.course-meta {
		margin-bottom: 20px;
	}

	.date {
		font-size: 13px;
		color: var(--aion-purple);
		font-weight: 500;
	}

	.course-links {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}

	.btn-secondary {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 10px 18px;
		background: transparent;
		border: 1px solid rgba(189, 147, 249, 0.4);
		border-radius: 20px;
		color: var(--aion-purple-light);
		font-size: 13px;
		font-weight: 600;
		text-decoration: none;
		transition: all 0.3s ease;
	}

	.btn-secondary:hover {
		background: rgba(189, 147, 249, 0.1);
		border-color: rgba(189, 147, 249, 0.6);
	}

	.coming-soon {
		font-size: 12px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: rgba(255, 255, 255, 0.4);
		padding: 10px 0;
	}

	.aionsec-cta {
		padding: 0 0 100px;
	}

	.cta-card {
		max-width: 500px;
		margin: 0 auto;
		text-align: center;
		padding: 40px;
		animation: ctaGlow 5s ease-in-out infinite;
	}

	@keyframes ctaGlow {
		0%, 100% {
			box-shadow: 0 0 15px rgba(245, 230, 99, 0.1), 0 0 30px rgba(245, 230, 99, 0.05);
		}
		50% {
			box-shadow: 0 0 20px rgba(245, 230, 99, 0.18), 0 0 40px rgba(245, 230, 99, 0.08);
		}
	}

	.cta-card h2 {
		font-size: 24px;
		margin-bottom: 12px;
	}

	.cta-card p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.7);
		margin-bottom: 24px;
	}

	@media (max-width: 768px) {
		.courses-grid {
			grid-template-columns: 1fr;
		}

		.course-card {
			padding: 24px;
		}

		.cta-card {
			padding: 32px 24px;
		}
	}
</style>
