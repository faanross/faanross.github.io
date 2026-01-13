<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});
</script>

<svelte:head>
	<title>Faan Rossouw | Research. Build. Teach.</title>
</svelte:head>

<section class="hero">
	{#if mounted}
		<div class="hero-content" in:fade={{ duration: 800 }}>
			<img src="/images/potential.png" alt="Faan Rossouw" class="hero-avatar" in:fly={{ y: 20, duration: 600, delay: 100 }} />
			<span class="tagline" in:fly={{ y: 20, duration: 600, delay: 200 }}><em>Panta rhei</em> (πάντα ῥεῖ)</span>
			<h1 in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Research. Build. Teach.
			</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 600 }}>
				Agentic Design + Threat Hunting
			</p>
		</div>
	{/if}
</section>

<section class="sections-grid">
	{#if mounted}
		<!-- Triangle - Courses -->
		<a href="/courses" class="section-card" in:fly={{ y: 30, duration: 600, delay: 800 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
					<polygon points="12,3 22,21 2,21" />
				</svg>
			</div>
			<h3>Courses</h3>
			<p>Hands-on workshops and courses on offensive tooling development and threat hunting</p>
			<span class="card-link">View courses</span>
		</a>

		<!-- Square - Articles -->
		<a href="/articles" class="section-card" in:fly={{ y: 30, duration: 600, delay: 900 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
					<rect x="3" y="3" width="18" height="18" />
				</svg>
			</div>
			<h3>Articles</h3>
			<p>Research on C2 techniques, threat hunting methodology, and malware analysis</p>
			<span class="card-link">Read articles</span>
		</a>

		<!-- Pentagon - Talks -->
		<a href="/talks" class="section-card" in:fly={{ y: 30, duration: 600, delay: 1000 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
					<polygon points="12,2 22,9 18.5,21 5.5,21 2,9" />
				</svg>
			</div>
			<h3>Talks</h3>
			<p>Webcasts and podcast appearances on C2, threat detection, and security research</p>
			<span class="card-link">Watch talks</span>
		</a>

		<!-- Hexagon - Projects -->
		<a href="/projects" class="section-card" in:fly={{ y: 30, duration: 600, delay: 1100 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
					<polygon points="12,2 21,7 21,17 12,22 3,17 3,7" />
				</svg>
			</div>
			<h3>Projects</h3>
			<p>Open-source C2 tools and research projects for the security community</p>
			<span class="card-link">View projects</span>
		</a>

		<!-- Heptagon - Claude -->
		<a href="/claude" class="section-card" in:fly={{ y: 30, duration: 600, delay: 1200 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
					<polygon points="12,2 19.8,5.5 22,13 17.5,20 6.5,20 2,13 4.2,5.5" />
				</svg>
			</div>
			<h3>Claude</h3>
			<p>Chronicling my journey optimizing workflows with Claude Code and Obsidian</p>
			<span class="card-link">Explore</span>
		</a>
	{/if}
</section>

<section class="cta-section">
	{#if mounted}
		<div class="cta-card glass-card" in:fly={{ y: 30, duration: 600, delay: 1300 }}>
			<h2>Let's Connect</h2>
			<p>Questions, ideas, or just want to chat about threat hunting and AI? I'm always happy to connect.</p>
			<a href="mailto:moi@faanross.com" class="btn-primary">Get in Touch</a>
		</div>
	{/if}
</section>

<style>
	.hero-avatar {
		width: 234px;
		height: 234px;
		border-radius: 50%;
		object-fit: cover;
		border: 3px solid rgba(189, 147, 249, 0.4);
		margin: 0 auto 24px;
		display: block;
	}

	.hero {
		min-height: 60vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 80px 24px 60px;
	}

	.hero-content {
		text-align: center;
		max-width: 800px;
	}

	.tagline {
		display: inline-block;
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.15em;
		color: var(--aion-yellow);
		margin-bottom: 16px;
	}

	h1 {
		font-size: clamp(36px, 5vw, 56px);
		font-weight: 700;
		line-height: 1.1;
		margin-bottom: 24px;
		color: var(--white);
	}

	.gradient-text {
		background: linear-gradient(135deg, var(--aion-purple), var(--aion-purple-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.lead {
		font-size: clamp(16px, 2vw, 20px);
		color: rgba(255, 255, 255, 0.7);
		line-height: 1.6;
		max-width: 600px;
		margin: 0 auto;
	}

	.sections-grid {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 24px;
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 24px 60px;
	}

	.section-card {
		flex: 0 1 280px;
		max-width: 320px;
		background: rgba(61, 61, 71, 0.4);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(189, 147, 249, 0.2);
		border-radius: 16px;
		padding: 32px;
		text-decoration: none;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		display: flex;
		flex-direction: column;
		text-align: center;
		align-items: center;
	}

	.section-card:hover {
		transform: translateY(-4px);
		border-color: rgba(189, 147, 249, 0.5);
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
	}

	.card-icon {
		width: 56px;
		height: 56px;
		border-radius: 12px;
		background: linear-gradient(135deg, var(--aion-purple), var(--aion-purple-light));
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 20px;
		color: var(--black);
	}

	.section-card h3 {
		font-size: 20px;
		font-weight: 600;
		color: var(--white);
		margin-bottom: 12px;
	}

	.section-card p {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.6);
		line-height: 1.6;
		flex-grow: 1;
		margin-bottom: 16px;
	}

	.card-link {
		font-size: 13px;
		font-weight: 600;
		color: var(--aion-yellow);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.section-card:hover .card-link {
		color: var(--aion-yellow-light);
	}

	.cta-section {
		max-width: 800px;
		margin: 0 auto;
		padding: 0 24px 80px;
	}

	.cta-card {
		text-align: center;
		padding: 48px;
	}

	.cta-card h2 {
		font-size: clamp(24px, 3vw, 32px);
		font-weight: 600;
		color: var(--white);
		margin-bottom: 16px;
	}

	.cta-card p {
		font-size: 16px;
		color: rgba(255, 255, 255, 0.7);
		margin-bottom: 28px;
		max-width: 500px;
		margin-left: auto;
		margin-right: auto;
	}

	@media (max-width: 768px) {
		.hero {
			min-height: 50vh;
			padding: 60px 24px 40px;
		}

		.sections-grid {
			grid-template-columns: 1fr;
		}

		.cta-card {
			padding: 32px 24px;
		}
	}
</style>
