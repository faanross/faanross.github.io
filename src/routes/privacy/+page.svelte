<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});
</script>

<svelte:head>
	<title>Privacy Policy | Faan Rossouw</title>
	<meta name="description" content="Privacy policy for faanross.com" />
</svelte:head>

<section class="privacy">
	<div class="container">
		{#if mounted}
			<header class="privacy-header" in:fly={{ y: 30, duration: 800, delay: 200 }}>
				<h1>Privacy Policy</h1>
				<p class="meta">Last updated: January 2026</p>
			</header>

			<div class="privacy-content" in:fly={{ y: 30, duration: 600, delay: 400 }}>
				<h2>Overview</h2>
				<p>This is a personal website. I respect your privacy and collect minimal information necessary for the site to function.</p>

				<h2>Information Collection</h2>
				<p>This website:</p>
				<ul>
					<li>Does not use cookies for tracking</li>
					<li>Does not collect personal information unless you email me directly</li>
					<li>Does not use third-party analytics services</li>
					<li>Is hosted on GitHub Pages</li>
				</ul>

				<h2>External Links</h2>
				<p>This site contains links to external websites (GitHub, YouTube, Active Countermeasures, AntiSyphon, etc.). These external sites have their own privacy policies that govern how they collect and use your information.</p>

				<h2>Contact</h2>
				<p>If you email me at <a href="mailto:moi@faanross.com">moi@faanross.com</a>, I'll use your email only to respond to your message. I won't add you to any mailing lists or share your contact information.</p>

				<h2>Changes</h2>
				<p>This policy may be updated occasionally. Any changes will be reflected on this page with an updated date.</p>
			</div>
		{/if}
	</div>
</section>

<style>
	.privacy {
		padding: 80px 0 120px;
	}

	.container {
		max-width: 800px;
	}

	.privacy-header {
		margin-bottom: 48px;
	}

	.privacy-header h1 {
		margin-bottom: 16px;
	}

	.meta {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.5);
	}

	.privacy-content {
		font-size: 16px;
		line-height: 1.8;
	}

	.privacy-content h2 {
		margin-top: 40px;
		margin-bottom: 16px;
		font-size: 24px;
	}

	.privacy-content p {
		margin-bottom: 16px;
		color: rgba(255, 255, 255, 0.85);
	}

	.privacy-content ul {
		margin: 16px 0;
		padding-left: 0;
		list-style: none;
	}

	.privacy-content li {
		padding: 6px 0 6px 24px;
		position: relative;
		color: rgba(255, 255, 255, 0.85);
	}

	.privacy-content li::before {
		content: '>';
		position: absolute;
		left: 0;
		color: var(--aion-purple);
		font-weight: 600;
	}

	.privacy-content a {
		color: var(--aion-purple);
	}

	@media (max-width: 768px) {
		.privacy {
			padding: 60px 0 80px;
		}

		.privacy-content {
			font-size: 15px;
		}
	}
</style>
