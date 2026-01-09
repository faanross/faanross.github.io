<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});
</script>

<svelte:head>
	<title>Faan Rossouw | Security Research</title>
</svelte:head>

<section class="hero">
	{#if mounted}
		<div class="hero-content" in:fade={{ duration: 800 }}>
			<img src="/images/moi.png" alt="Faan Rossouw" class="hero-avatar" in:fly={{ y: 20, duration: 600, delay: 100 }} />
			<span class="tagline" in:fly={{ y: 20, duration: 600, delay: 200 }}>Threat Hunting + Agentic AI</span>
			<h1 in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Never Stop <span class="gradient-text">Hunting</span>
			</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 600 }}>
				Research. Build. Teach.
			</p>
		</div>
	{/if}
</section>

<section class="sections-grid">
	{#if mounted}
		<a href="/courses" class="section-card" in:fly={{ y: 30, duration: 600, delay: 800 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
					<path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
				</svg>
			</div>
			<h3>Courses</h3>
			<p>Workshops on C2 development, reflective loading, and malware techniques in Golang</p>
			<span class="card-link">View courses</span>
		</a>

		<a href="/articles" class="section-card" in:fly={{ y: 30, duration: 600, delay: 900 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"></path>
					<polyline points="14 2 14 8 20 8"></polyline>
					<line x1="16" y1="13" x2="8" y2="13"></line>
					<line x1="16" y1="17" x2="8" y2="17"></line>
					<line x1="10" y1="9" x2="8" y2="9"></line>
				</svg>
			</div>
			<h3>Articles</h3>
			<p>Research on C2 techniques, threat hunting methodology, and malware analysis</p>
			<span class="card-link">Read articles</span>
		</a>

		<a href="/talks" class="section-card" in:fly={{ y: 30, duration: 600, delay: 1000 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<polygon points="23 7 16 12 23 17 23 7"></polygon>
					<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
				</svg>
			</div>
			<h3>Talks</h3>
			<p>Webcasts and podcast appearances on C2, threat detection, and security research</p>
			<span class="card-link">Watch talks</span>
		</a>

		<a href="/projects" class="section-card" in:fly={{ y: 30, duration: 600, delay: 1100 }}>
			<div class="card-icon">
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"></path>
				</svg>
			</div>
			<h3>Projects</h3>
			<p>Open-source C2 tools and research projects for the security community</p>
			<span class="card-link">View projects</span>
		</a>
	{/if}
</section>

<section class="cta-section">
	{#if mounted}
		<div class="cta-card glass-card" in:fly={{ y: 30, duration: 600, delay: 1200 }}>
			<h2>Let's Connect</h2>
			<p>Questions, ideas, or just want to chat about threat hunting and AI? I'm always happy to connect.</p>
			<a href="mailto:moi@faanross.com" class="btn-primary">Get in Touch</a>
		</div>
	{/if}
</section>

<style>
	.hero-avatar {
		width: 180px;
		height: 180px;
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
