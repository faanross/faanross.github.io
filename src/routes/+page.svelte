<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	let mounted = $state(false);

	// Typewriter effect state
	let displayedText = $state('');
	const fullText = 'Research. Build. Teach.';
	let typewriterComplete = $state(false);

	onMount(() => {
		mounted = true;

		// Start typewriter after initial fade
		setTimeout(() => {
			let i = 0;
			const typeInterval = setInterval(() => {
				if (i < fullText.length) {
					displayedText = fullText.slice(0, i + 1);
					i++;
				} else {
					clearInterval(typeInterval);
					typewriterComplete = true;
				}
			}, 80);
		}, 500);
	});
</script>

<svelte:head>
	<title>Faan Rossouw | Research. Build. Teach.</title>
</svelte:head>

<section class="hero">
	{#if mounted}
		<div class="hero-content" in:fade={{ duration: 800 }}>
			<div class="avatar-wrapper image-glow" in:fly={{ y: 20, duration: 600, delay: 100 }}>
				<div class="rotating-ring"></div>
				<img src="/images/site/potential.png" alt="Faan Rossouw" class="hero-avatar" />
			</div>
			<span class="tagline" in:fly={{ y: 20, duration: 600, delay: 200 }}><em>Panta rhei</em> (πάντα ῥεῖ)</span>
			<h1 in:fly={{ y: 20, duration: 600, delay: 400 }}>
				{displayedText}<span class="cursor" class:hidden={typewriterComplete}>|</span>
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

<section class="connect-section">
	{#if mounted}
		<h2 in:fly={{ y: 30, duration: 600, delay: 1300 }}>Connect</h2>
		<div class="social-links" in:fly={{ y: 30, duration: 600, delay: 1400 }}>
			<a href="https://github.com/faanross" target="_blank" rel="noopener noreferrer" class="social-card glass-card">
				<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
					<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
				</svg>
				<span>GitHub</span>
			</a>
			<a href="https://www.youtube.com/@FaanRoss" target="_blank" rel="noopener noreferrer" class="social-card glass-card">
				<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
					<path d="M19.615 3.184c-3.604-.246-11.631-.245-15.23 0-3.897.266-4.356 2.62-4.385 8.816.029 6.185.484 8.549 4.385 8.816 3.6.245 11.626.246 15.23 0 3.897-.266 4.356-2.62 4.385-8.816-.029-6.185-.484-8.549-4.385-8.816zm-10.615 12.816v-8l8 3.993-8 4.007z"/>
				</svg>
				<span>YouTube</span>
			</a>
			<a href="https://x.com/faanross" target="_blank" rel="noopener noreferrer" class="social-card glass-card">
				<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
					<path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
				</svg>
				<span>X</span>
			</a>
			<a href="https://www.linkedin.com/in/faan-rossouw" target="_blank" rel="noopener noreferrer" class="social-card glass-card">
				<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
					<path d="M19 0h-14c-2.761 0-5 2.239-5 5v14c0 2.761 2.239 5 5 5h14c2.762 0 5-2.239 5-5v-14c0-2.761-2.238-5-5-5zm-11 19h-3v-11h3v11zm-1.5-12.268c-.966 0-1.75-.79-1.75-1.764s.784-1.764 1.75-1.764 1.75.79 1.75 1.764-.783 1.764-1.75 1.764zm13.5 12.268h-3v-5.604c0-3.368-4-3.113-4 0v5.604h-3v-11h3v1.765c1.396-2.586 7-2.777 7 2.476v6.759z"/>
				</svg>
				<span>LinkedIn</span>
			</a>
			<a href="https://discord.gg/fdDPBnEC" target="_blank" rel="noopener noreferrer" class="social-card glass-card">
				<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
					<path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.4189-2.1568 2.4189Z"/>
				</svg>
				<span>Discord</span>
			</a>
			<a href="mailto:moi@faanross.com" class="social-card glass-card">
				<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
					<polyline points="22,6 12,13 2,6"></polyline>
				</svg>
				<span>Email</span>
			</a>
		</div>
	{/if}
</section>

<section class="aionsec-cta">
	{#if mounted}
		<div class="cta-card glass-card" in:fly={{ y: 30, duration: 600, delay: 1500 }}>
			<h2>Agentic AI for Security Professionals</h2>
			<p>Learn how agentic design patterns can transform your security workflows — from threat hunting to detection engineering</p>
			<a href="https://aionsec.ai/courses" target="_blank" rel="noopener noreferrer" class="btn-primary">
				Explore AionSec
			</a>
		</div>
	{/if}
</section>

<style>
	/* Avatar with rotating ring */
	.avatar-wrapper {
		position: relative;
		width: 234px;
		height: 234px;
		margin: 0 auto 40px;
	}

	.rotating-ring {
		position: absolute;
		inset: -8px;
		border-radius: 50%;
		background: conic-gradient(
			from 0deg,
			var(--aion-yellow) 0deg,
			var(--aion-yellow-light) 90deg,
			var(--aion-yellow-dark) 180deg,
			var(--aion-yellow-light) 270deg,
			var(--aion-yellow) 360deg
		);
		animation: rotateRing 8s linear infinite;
		-webkit-mask: radial-gradient(farthest-side, transparent calc(100% - 3px), black calc(100% - 3px));
		mask: radial-gradient(farthest-side, transparent calc(100% - 3px), black calc(100% - 3px));
	}

	@keyframes rotateRing {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	/* Golden glow pulse */
	.image-glow {
		animation: goldenGlow 5s ease-in-out infinite;
	}

	@keyframes goldenGlow {
		0%, 100% {
			filter: drop-shadow(0 0 20px rgba(245, 230, 99, 0.35));
		}
		50% {
			filter: drop-shadow(0 0 30px rgba(245, 230, 99, 0.5)) drop-shadow(0 0 45px rgba(245, 230, 99, 0.25));
		}
	}

	.hero-avatar {
		width: 234px;
		height: 234px;
		border-radius: 50%;
		object-fit: cover;
		border: 3px solid rgba(245, 230, 99, 0.4);
		display: block;
	}

	/* Typewriter cursor */
	.cursor {
		display: inline-block;
		color: var(--aion-yellow);
		animation: blink 0.8s infinite;
		margin-left: 2px;
	}

	.cursor.hidden {
		visibility: hidden;
	}

	@keyframes blink {
		0%, 50% { opacity: 1; }
		51%, 100% { opacity: 0; }
	}

	.hero {
		min-height: 80vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 100px 24px 120px;
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
		margin-bottom: 24px;
	}

	h1 {
		font-size: clamp(36px, 5vw, 56px);
		font-weight: 700;
		line-height: 1.1;
		margin-bottom: 20px;
		color: var(--white);
		white-space: nowrap;
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
		padding: 40px 24px 80px;
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

	.connect-section {
		padding: 60px 24px 100px;
		max-width: 800px;
		margin: 0 auto;
	}

	.connect-section h2 {
		text-align: center;
		margin-bottom: 32px;
		font-size: clamp(24px, 3vw, 32px);
		color: var(--white);
	}

	.social-links {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
		max-width: 600px;
		margin: 0 auto;
	}

	.social-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 24px 16px;
		text-decoration: none;
		color: var(--white);
		transition: all 0.3s ease;
	}

	.social-card:hover {
		transform: translateY(-3px);
		border-color: rgba(189, 147, 249, 0.5);
	}

	.social-card svg {
		color: var(--aion-purple);
	}

	.social-card:hover svg {
		color: var(--aion-purple-light);
	}

	.social-card span {
		font-size: 13px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.aionsec-cta {
		padding: 20px 24px 60px;
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
		.hero {
			min-height: 50vh;
			padding: 60px 24px 40px;
		}

		h1 {
			white-space: normal;
			font-size: clamp(28px, 10vw, 36px);
		}

		.sections-grid {
			grid-template-columns: 1fr;
		}

		.cta-card {
			padding: 32px 24px;
		}

		.social-card {
			padding: 16px 12px;
		}
	}
</style>
