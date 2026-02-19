<script lang="ts">
	import '$lib/styles/global.css';
	import Nav from '$lib/components/Nav.svelte';
	import Background from '$lib/components/Background.svelte';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { afterNavigate } from '$app/navigation';
	import { addCopyButtons } from '$lib/utils/addCopyButtons';

	let { children } = $props();

	onMount(() => {
		setTimeout(addCopyButtons, 100);
	});

	afterNavigate(() => {
		setTimeout(addCopyButtons, 100);
	});

	// Hide footer on dashboard routes
	const hiddenRoutes = ['/agentic/memory'];
	$effect(() => {
		// This is just to make hiddenRoutes reactive with page
	});
	const showFooter = $derived(!hiddenRoutes.some(route => $page.url.pathname.startsWith(route)));

	let footerFormSubmitted = $state(false);
	let footerFormLoading = $state(false);

	async function handleFooterSubmit(event: Event) {
		event.preventDefault();
		const form = event.target as HTMLFormElement;
		const formData = new FormData(form);

		footerFormLoading = true;

		try {
			await fetch(form.action, {
				method: 'POST',
				body: formData,
				mode: 'no-cors'
			});
			footerFormSubmitted = true;
		} catch (error) {
			console.error('Form submission error:', error);
			footerFormSubmitted = true;
		} finally {
			footerFormLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Faan Rossouw | Research. Build. Teach.</title>
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

{#if showFooter}
<footer>
	<div class="container">
		<div class="footer-newsletter">
			{#if !footerFormSubmitted}
				<span class="footer-newsletter-label">Stay updated</span>
				<form
					class="footer-newsletter-form"
					action="https://assets.mailerlite.com/jsonp/1462037/forms/152012474066929011/subscribe"
					method="post"
					onsubmit={handleFooterSubmit}
				>
					<input
						type="email"
						name="fields[email]"
						placeholder="Email"
						required
						autocomplete="email"
					/>
					<input type="hidden" name="ml-submit" value="1" />
					<input type="hidden" name="anticsrf" value="true" />
					<button type="submit" disabled={footerFormLoading}>
						{footerFormLoading ? '...' : 'Subscribe'}
					</button>
				</form>
			{:else}
				<span class="footer-newsletter-success">You're in!</span>
			{/if}
		</div>
		<div class="footer-links">
			<a href="https://github.com/faanross" target="_blank" rel="noopener noreferrer" class="footer-link">GitHub</a>
			<a href="https://www.youtube.com/@FaanRoss" target="_blank" rel="noopener noreferrer" class="footer-link">YouTube</a>
			<a href="https://x.com/faanross" target="_blank" rel="noopener noreferrer" class="footer-link">X</a>
			<a href="https://www.linkedin.com/in/faan-rossouw" target="_blank" rel="noopener noreferrer" class="footer-link">LinkedIn</a>
			<a href="https://discord.gg/fdDPBnEC" target="_blank" rel="noopener noreferrer" class="footer-link">Discord</a>
		</div>
		<p>&copy; 2026 Faan Rossouw</p>
	</div>
</footer>
{/if}

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

	.footer-newsletter {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
	}

	.footer-newsletter-label {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.footer-newsletter-form {
		display: flex;
		gap: 8px;
	}

	.footer-newsletter-form input[type="email"] {
		padding: 8px 12px;
		font-size: 13px;
		border: 1px solid rgba(189, 147, 249, 0.3);
		border-radius: 6px;
		background: rgba(0, 0, 0, 0.3);
		color: var(--white);
		width: 180px;
		transition: border-color 0.2s;
	}

	.footer-newsletter-form input[type="email"]::placeholder {
		color: rgba(255, 255, 255, 0.4);
	}

	.footer-newsletter-form input[type="email"]:focus {
		outline: none;
		border-color: var(--aion-purple);
	}

	.footer-newsletter-form button {
		padding: 8px 16px;
		font-size: 12px;
		font-weight: 600;
		color: var(--black);
		background: var(--aion-yellow);
		border: none;
		border-radius: 6px;
		cursor: pointer;
		transition: background 0.2s;
	}

	.footer-newsletter-form button:hover:not(:disabled) {
		background: var(--aion-yellow-light);
	}

	.footer-newsletter-form button:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	.footer-newsletter-success {
		font-size: 13px;
		color: var(--aion-purple-light);
		font-weight: 500;
	}

	@media (max-width: 768px) {
		.footer-links {
			gap: 12px;
		}

		.footer-newsletter {
			flex-direction: column;
			gap: 8px;
		}

		.footer-newsletter-form {
			flex-direction: column;
			width: 100%;
			max-width: 250px;
		}

		.footer-newsletter-form input[type="email"] {
			width: 100%;
		}

		.footer-newsletter-form button {
			width: 100%;
		}
	}
</style>
