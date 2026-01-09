<script lang="ts">
	import { page } from '$app/stores';

	const navItems = [
		{ href: '/', label: 'Home' },
		{ href: '/courses', label: 'Courses' },
		{ href: '/articles', label: 'Articles' },
		{ href: '/talks', label: 'Talks' },
		{ href: '/projects', label: 'Projects' },
		{ href: '/about', label: 'About' }
	];

	let mobileMenuOpen = $state(false);

	function toggleMenu() {
		mobileMenuOpen = !mobileMenuOpen;
	}

	function closeMenu() {
		mobileMenuOpen = false;
	}
</script>

<nav class="nav">
	<div class="nav-container">
		<a href="/" class="logo" onclick={closeMenu}>
			<img src="/images/moi.png" alt="Faan Rossouw" class="logo-img" />
			<span>Faan Rossouw</span>
		</a>

		<button class="mobile-toggle" onclick={toggleMenu} aria-label="Toggle menu">
			<span class="bar" class:open={mobileMenuOpen}></span>
			<span class="bar" class:open={mobileMenuOpen}></span>
			<span class="bar" class:open={mobileMenuOpen}></span>
		</button>

		<ul class="nav-links" class:open={mobileMenuOpen}>
			{#each navItems as item}
				<li>
					<a
						href={item.href}
						class:active={$page.url.pathname === item.href || ($page.url.pathname.startsWith(item.href) && item.href !== '/')}
						onclick={closeMenu}
					>
						{item.label}
					</a>
				</li>
			{/each}
			<li>
				<a
					href="https://aionsec.ai"
					target="_blank"
					rel="noopener noreferrer"
					class="cta-link"
					onclick={closeMenu}
				>
					AionSec
				</a>
			</li>
		</ul>
	</div>
</nav>

<style>
	.nav {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		z-index: 100;
		background: linear-gradient(180deg, rgba(30, 30, 38, 0.6) 0%, rgba(30, 30, 38, 0) 100%);
		backdrop-filter: blur(8px);
	}

	.nav-container {
		max-width: 1400px;
		margin: 0 auto;
		padding: 0 48px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.logo {
		display: flex;
		align-items: center;
		gap: 10px;
		text-decoration: none;
	}

	.logo-img {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		object-fit: cover;
		border: 2px solid rgba(189, 147, 249, 0.4);
	}

	.logo span:not(.logo-text) {
		font-weight: 600;
		font-size: 14px;
		color: var(--white);
		letter-spacing: 0.02em;
	}

	.nav-links {
		display: flex;
		align-items: center;
		gap: 24px;
		list-style: none;
	}

	.nav-links a {
		font-size: 14px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.7);
		text-decoration: none;
		transition: all 0.3s ease;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}

	.nav-links a:hover,
	.nav-links a.active {
		color: var(--aion-purple);
	}

	.cta-link {
		padding: 6px 14px !important;
		background: var(--aion-yellow);
		border-radius: 16px;
		color: var(--aion-grey-dark) !important;
		font-weight: 600 !important;
	}

	.cta-link:hover {
		background: var(--aion-yellow-light);
		color: var(--aion-grey-dark) !important;
		transform: translateY(-1px);
	}

	.mobile-toggle {
		display: none;
		flex-direction: column;
		gap: 4px;
		background: none;
		border: none;
		cursor: pointer;
		padding: 8px;
	}

	.bar {
		width: 20px;
		height: 2px;
		background: var(--white);
		transition: all 0.3s ease;
	}

	.bar.open:nth-child(1) {
		transform: rotate(45deg) translate(4px, 4px);
	}

	.bar.open:nth-child(2) {
		opacity: 0;
	}

	.bar.open:nth-child(3) {
		transform: rotate(-45deg) translate(4px, -4px);
	}

	@media (max-width: 768px) {
		.nav-container {
			padding: 0 24px;
		}

		.mobile-toggle {
			display: flex;
		}

		.nav-links {
			position: fixed;
			top: 44px;
			left: 0;
			right: 0;
			bottom: 0;
			height: calc(100vh - 44px);
			flex-direction: column;
			justify-content: center;
			align-items: center;
			background: rgba(0, 0, 0, 0.85);
			backdrop-filter: blur(20px);
			-webkit-backdrop-filter: blur(20px);
			padding: 32px 24px;
			gap: 44px;
			transform: translateX(100%);
			transition: transform 0.3s ease;
		}

		.nav-links.open {
			transform: translateX(0);
		}

		.nav-links a {
			font-size: 28px;
		}
	}
</style>
