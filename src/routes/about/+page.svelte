<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});
</script>

<svelte:head>
	<title>About | Faan Rossouw</title>
	<meta name="description" content="Building the future of threat hunting at the intersection of agentic AI and security. Researcher at Active Countermeasures, instructor at AntiSyphon." />
</svelte:head>

<section class="about-hero">
	<div class="container">
		{#if mounted}
			<h1 in:fly={{ y: 30, duration: 800, delay: 200 }}>hi.</h1>
		{/if}
	</div>
</section>

<section class="about-content">
	<div class="container">
		{#if mounted}
			<div class="content-block" in:fly={{ y: 30, duration: 600, delay: 400 }}>
				<p>
					I'm a security researcher focused on the intersection of <strong>threat hunting</strong>
					and <strong>agentic AI</strong>. The convergence of these fields is unlocking what was
					impossible before: continuous hunting at machine speed with human intuition.
				</p>
				<p>
					I'm currently building <a href="https://aionsec.ai" target="_blank" rel="noopener noreferrer">aionsec.ai</a>,
					an open-source platform to make elite threat hunting accessible to everyone.
				</p>
				<p>
					I also do research at <a href="https://www.activecountermeasures.com" target="_blank" rel="noopener noreferrer">Active Countermeasures</a>
					and instruct at <a href="https://www.antisyphontraining.com/instructor/faan-ross/" target="_blank" rel="noopener noreferrer">AntiSyphon</a>,
					teaching threat hunting and offensive security tooling.
				</p>
				<p>
					If you have any questions or feedback, feel free to
					<a href="mailto:moi@faanross.com">connect</a>.
				</p>
				<p class="signoff">
					<strong>Live long and prosper.</strong>
				</p>
				<p class="name">Faan</p>
			</div>

			<div class="profile-image" in:fly={{ y: 30, duration: 600, delay: 600 }}>
				<img src="/images/site/profpic.png" alt="Faan Rossouw" />
			</div>

			<div class="quote-block glass-card" in:fly={{ y: 30, duration: 600, delay: 800 }}>
				<p><em>People don't have ideas, ideas have people.</em></p>
				<span class="attribution">— CGJ</span>
			</div>
		{/if}
	</div>
</section>

<section class="connect-section">
	<div class="container">
		{#if mounted}
			<h2 in:fly={{ y: 30, duration: 600, delay: 900 }}>Connect</h2>
			<div class="social-links" in:fly={{ y: 30, duration: 600, delay: 1000 }}>
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
	</div>
</section>

<style>
	.about-hero {
		padding: 80px 0 40px;
		text-align: center;
	}

	.about-hero h1 {
		font-size: clamp(48px, 6vw, 72px);
		font-weight: 700;
		color: var(--white);
	}

	.about-content {
		padding: 40px 0 60px;
	}

	.about-content .container {
		max-width: 700px;
	}

	.content-block {
		margin-bottom: 48px;
	}

	.content-block p {
		font-size: 18px;
		line-height: 1.8;
		margin-bottom: 24px;
		color: rgba(255, 255, 255, 0.85);
	}

	.content-block strong {
		color: var(--aion-purple);
		font-weight: 600;
	}

	.content-block a {
		color: var(--aion-yellow);
		text-decoration: none;
		border-bottom: 1px solid transparent;
		transition: border-color 0.3s ease;
	}

	.content-block a:hover {
		border-bottom-color: var(--aion-yellow);
	}

	.signoff {
		margin-top: 32px;
	}

	.signoff strong {
		color: var(--aion-yellow) !important;
	}

	.name {
		font-size: 20px;
		font-weight: 600;
		color: var(--white);
		margin-top: 8px;
	}

	.profile-image {
		text-align: center;
		margin-bottom: 48px;
	}

	.profile-image img {
		max-width: 560px;
		border-radius: 16px;
		border: 2px solid rgba(189, 147, 249, 0.3);
	}

	.quote-block {
		text-align: center;
		padding: 32px;
	}

	.quote-block p {
		font-size: 18px;
		color: rgba(255, 255, 255, 0.9);
		margin-bottom: 12px;
	}

	.quote-block em {
		font-style: italic;
	}

	.attribution {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.5);
	}

	.connect-section {
		padding: 60px 0 100px;
	}

	.connect-section h2 {
		text-align: center;
		margin-bottom: 32px;
		font-size: clamp(24px, 3vw, 32px);
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

	@media (max-width: 768px) {
		.content-block p {
			font-size: 16px;
		}

		.profile-image img {
			max-width: 200px;
		}

		.social-links {
			grid-template-columns: repeat(3, 1fr);
		}

		.social-card {
			padding: 16px 12px;
		}
	}
</style>
