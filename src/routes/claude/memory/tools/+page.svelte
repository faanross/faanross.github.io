<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { loadParquetFiles, getTopTools, getToolTrend, type ToolCount, type ToolTrend } from '$lib/db/duckdb';
	import BarChart from '../components/BarChart.svelte';
	import TrendChart from '../components/TrendChart.svelte';
	import Breadcrumbs from '../components/Breadcrumbs.svelte';

	let mounted = $state(false);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let topTools = $state<ToolCount[]>([]);
	let toolTrend = $state<ToolTrend[]>([]);

	const crumbs = [
		{ label: 'Dashboard', href: '/claude/memory' },
		{ label: 'Tool Calls' }
	];

	onMount(async () => {
		mounted = true;

		try {
			await loadParquetFiles();
			[topTools, toolTrend] = await Promise.all([getTopTools(12), getToolTrend()]);
		} catch (e) {
			console.error('Failed to load data:', e);
			error = e instanceof Error ? e.message : 'Unknown error';
		} finally {
			loading = false;
		}
	});

	const barData = $derived(
		topTools.map((t) => ({
			label: t.tool_name,
			value: t.count
		}))
	);

	const trendData = $derived(
		toolTrend.map((t) => ({
			date: t.date,
			count: t.count
		}))
	);
</script>

<svelte:head>
	<title>Tool Calls | Memory Dashboard</title>
</svelte:head>

<section class="page-header">
	<div class="container">
		{#if mounted}
			<div in:fade={{ duration: 300 }}>
				<Breadcrumbs {crumbs} />
			</div>
			<h1 in:fly={{ y: 30, duration: 800, delay: 100 }}>Tool Calls</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 200 }}>
				How Claude interacts with the system
			</p>
		{/if}
	</div>
</section>

<section class="page-content">
	<div class="container">
		{#if loading}
			<div class="loading-state" in:fade={{ duration: 300 }}>
				<div class="spinner"></div>
				<p>Loading tool data...</p>
			</div>
		{:else if error}
			<div class="error-state glass-card" in:fade={{ duration: 300 }}>
				<h3>Failed to load data</h3>
				<p>{error}</p>
			</div>
		{:else}
			<div class="visualizations" in:fly={{ y: 30, duration: 600 }}>
				{#if trendData.length > 0}
					<TrendChart data={trendData} title="Tool Usage Over Time" />
				{/if}

				{#if barData.length > 0}
					<BarChart data={barData} title="Most Used Tools" />
				{/if}
			</div>
		{/if}
	</div>
</section>

<style>
	.page-header {
		padding: 40px 0 30px;
		text-align: center;
	}

	.page-header h1 {
		margin-bottom: 12px;
		background: linear-gradient(135deg, var(--aion-purple), var(--aion-purple-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.lead {
		font-size: clamp(14px, 2vw, 18px);
		color: rgba(255, 255, 255, 0.7);
	}

	.page-content {
		padding: 20px 0 80px;
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
		to { transform: rotate(360deg); }
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

	.visualizations {
		max-width: 800px;
		margin: 0 auto;
	}
</style>
