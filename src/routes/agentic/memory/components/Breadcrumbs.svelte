<script lang="ts">
	type Crumb = {
		label: string;
		href?: string;
	};

	let { crumbs = [] }: { crumbs: Crumb[] } = $props();
</script>

<nav class="breadcrumbs" aria-label="Breadcrumb">
	<ol>
		{#each crumbs as crumb, i}
			<li>
				{#if crumb.href && i < crumbs.length - 1}
					<a href={crumb.href}>{crumb.label}</a>
				{:else}
					<span class="current">{crumb.label}</span>
				{/if}
				{#if i < crumbs.length - 1}
					<span class="separator">/</span>
				{/if}
			</li>
		{/each}
	</ol>
</nav>

<style>
	.breadcrumbs {
		margin-bottom: 24px;
	}

	ol {
		display: flex;
		align-items: center;
		gap: 8px;
		list-style: none;
		padding: 0;
		margin: 0;
		font-size: 13px;
	}

	li {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	a {
		color: rgba(255, 255, 255, 0.5);
		text-decoration: none;
		transition: color 0.2s;
	}

	a:hover {
		color: var(--aion-purple);
	}

	.current {
		color: var(--aion-purple-light);
		font-weight: 500;
	}

	.separator {
		color: rgba(255, 255, 255, 0.3);
	}
</style>
