<script lang="ts">
	import '$lib/styles/global.css';
	import Nav from '$lib/components/Nav.svelte';
	import Background from '$lib/components/Background.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<title>Faan Rossouw | Security Research</title>
	<meta name="description" content="Building the future of autonomous threat hunting at the intersection of agentic AI and security." />
	<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
	<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
	<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
</svelte:head>

<Background />
<Nav />

<main>
	{@render children()}
</main>

<footer>
	<div class="container">
		<p>&copy; 2026 Faan Rossouw</p>
		<div class="footer-links">
			<a href="https://github.com/faanross" target="_blank" rel="noopener noreferrer" class="footer-link">GitHub</a>
			<a href="https://www.youtube.com/@FaanRoss" target="_blank" rel="noopener noreferrer" class="footer-link">YouTube</a>
			<a href="https://x.com/faanross" target="_blank" rel="noopener noreferrer" class="footer-link">X</a>
			<a href="https://www.linkedin.com/in/faan-rossouw" target="_blank" rel="noopener noreferrer" class="footer-link">LinkedIn</a>
			<a href="https://discord.gg/fdDPBnEC" target="_blank" rel="noopener noreferrer" class="footer-link">Discord</a>
		</div>
	</div>
</footer>

<style>
	main {
		position: relative;
		z-index: 10;
		min-height: calc(100vh - 140px);
		padding-top: 48px;
	}

	footer {
		position: relative;
		z-index: 10;
		padding: 20px 0;
		text-align: center;
		border-top: 1px solid rgba(189, 147, 249, 0.1);
	}

	footer .container {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 12px;
	}

	.footer-links {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 16px;
	}

	footer p {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.5);
	}

	:global(.footer-link) {
		font-size: 11px;
		color: var(--aion-yellow);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	:global(.footer-link:hover) {
		color: var(--aion-yellow-light);
	}

	@media (max-width: 768px) {
		.footer-links {
			gap: 12px;
		}
	}
</style>
