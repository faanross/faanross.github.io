<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { loadParquetFiles, getStats } from '$lib/db/duckdb';

	let mounted = $state(false);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let stats = $state<{ sessions: number; messages: number; toolCalls: number } | null>(null);

	onMount(async () => {
		mounted = true;

		try {
			await loadParquetFiles();
			stats = await getStats();
		} catch (e) {
			console.error('Failed to initialize database:', e);
			error = e instanceof Error ? e.message : 'Unknown error';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Memory Dashboard | Faan Rossouw</title>
	<meta name="description" content="Visual exploration of Claude conversation history and patterns." />
</svelte:head>

<section class="memory-hero">
	<div class="container">
		{#if mounted}
			<h1 in:fly={{ y: 30, duration: 800, delay: 200 }}>Memory Dashboard</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Visual exploration of conversation history
			</p>
		{/if}
	</div>
</section>

<section class="memory-content">
	<div class="container">
		{#if loading}
			<div class="loading-state" in:fade={{ duration: 300 }}>
				<div class="spinner"></div>
				<p>Loading conversation data...</p>
			</div>
		{:else if error}
			<div class="error-state glass-card" in:fade={{ duration: 300 }}>
				<h3>Failed to load data</h3>
				<p>{error}</p>
			</div>
		{:else if stats}
			<div class="stats-grid" in:fly={{ y: 30, duration: 600 }}>
				<a href="/claude/memory/sessions" class="stat-card glass-card">
					<span class="stat-value">{stats.sessions.toLocaleString()}</span>
					<span class="stat-label">Sessions</span>
					<span class="stat-hint">Activity patterns, session browser</span>
				</a>
				<a href="/claude/memory/messages" class="stat-card glass-card">
					<span class="stat-value">{stats.messages.toLocaleString()}</span>
					<span class="stat-label">Messages</span>
					<span class="stat-hint">Frequency, ratios, trends</span>
				</a>
				<a href="/claude/memory/tools" class="stat-card glass-card">
					<span class="stat-value">{stats.toolCalls.toLocaleString()}</span>
					<span class="stat-label">Tool Calls</span>
					<span class="stat-hint">Most used tools, patterns</span>
				</a>
			</div>
		{/if}
	</div>
</section>

<style>
	.memory-hero {
		padding: 80px 0 40px;
		text-align: center;
	}

	.memory-hero h1 {
		margin-bottom: 16px;
		background: linear-gradient(135deg, var(--aion-purple), var(--aion-purple-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.lead {
		font-size: clamp(16px, 2vw, 20px);
		color: rgba(255, 255, 255, 0.7);
	}

	.memory-content {
		padding: 40px 0 80px;
	}

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
		padding: 60px 0;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid rgba(189, 147, 249, 0.2);
		border-top-color: var(--aion-purple);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.loading-state p {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.6);
	}

	.error-state {
		max-width: 500px;
		margin: 0 auto;
		padding: 32px;
		text-align: center;
	}

	.error-state h3 {
		margin-bottom: 12px;
		color: #ff6b6b;
	}

	.error-state p {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.6);
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 24px;
		max-width: 900px;
		margin: 0 auto;
	}

	.stat-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 40px 24px 32px;
		text-align: center;
		text-decoration: none;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.stat-card:hover {
		transform: translateY(-4px);
		border-color: var(--aion-yellow);
		box-shadow: 0 12px 32px rgba(245, 230, 99, 0.15);
	}

	.stat-card:hover .stat-value {
		color: var(--aion-yellow-light);
	}

	.stat-card:hover .stat-hint {
		color: rgba(255, 255, 255, 0.6);
	}

	.stat-value {
		font-size: clamp(32px, 5vw, 48px);
		font-weight: 700;
		color: var(--aion-yellow);
		line-height: 1;
		margin-bottom: 8px;
		transition: color 0.3s;
	}

	.stat-label {
		font-size: 14px;
		font-weight: 600;
		color: rgba(255, 255, 255, 0.8);
		text-transform: uppercase;
		letter-spacing: 0.1em;
		margin-bottom: 12px;
	}

	.stat-hint {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.4);
		transition: color 0.3s;
	}

	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: 1fr;
			gap: 16px;
		}

		.stat-card {
			padding: 32px 20px 24px;
		}
	}
</style>
