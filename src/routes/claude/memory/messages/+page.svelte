<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import {
		loadParquetFiles,
		getMessageTrend,
		getMessagesByRole,
		getMessageLengthDistribution,
		getContextStats,
		getContentTypeDistribution,
		getLargeMessages,
		type DailyCount,
		type RoleCount,
		type LengthBucket,
		type ContextStats,
		type ContentTypeCount,
		type LargeMessageInfo
	} from '$lib/db/duckdb';
	import TrendChart from '../components/TrendChart.svelte';
	import DonutChart from '../components/DonutChart.svelte';
	import BarChart from '../components/BarChart.svelte';
	import StatsGrid from '../components/StatsGrid.svelte';
	import DataTable from '../components/DataTable.svelte';
	import Breadcrumbs from '../components/Breadcrumbs.svelte';

	let mounted = $state(false);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let trendData = $state<DailyCount[]>([]);
	let roleData = $state<RoleCount[]>([]);
	let lengthData = $state<LengthBucket[]>([]);
	let contextStats = $state<ContextStats | null>(null);
	let contentTypes = $state<ContentTypeCount[]>([]);
	let largeMessages = $state<LargeMessageInfo[]>([]);

	const crumbs = [
		{ label: 'Dashboard', href: '/claude/memory' },
		{ label: 'Messages' }
	];

	onMount(async () => {
		mounted = true;

		try {
			await loadParquetFiles();
			[trendData, roleData, lengthData, contextStats, contentTypes, largeMessages] = await Promise.all([
				getMessageTrend(),
				getMessagesByRole(),
				getMessageLengthDistribution(),
				getContextStats(),
				getContentTypeDistribution(),
				getLargeMessages(10000, 15)
			]);
		} catch (e) {
			console.error('Failed to load data:', e);
			error = e instanceof Error ? e.message : 'Unknown error';
		} finally {
			loading = false;
		}
	});

	const roleChartData = $derived(
		roleData.map((r) => ({
			label: r.role,
			value: r.count
		}))
	);

	const lengthChartData = $derived(
		lengthData.map((l) => ({
			label: l.bucket,
			value: l.count
		}))
	);

	const contentTypeChartData = $derived(
		contentTypes.map((c) => ({
			label: c.content_type || 'text',
			value: c.count
		}))
	);

	const contextStatsData = $derived(
		contextStats
			? [
					{ label: 'Total Tokens (Est)', value: Math.round(contextStats.total_estimated_tokens / 1000000 * 10) / 10 + 'M' },
					{ label: 'Avg Tokens/Session', value: Math.round(contextStats.avg_tokens_per_session / 1000) + 'K' },
					{ label: 'Largest Session', value: Math.round(contextStats.largest_session_tokens / 1000) + 'K' },
					{ label: 'Sessions > 100K', value: contextStats.sessions_over_100k_tokens }
				]
			: []
	);

	const largeMessageColumns = [
		{ key: 'type', label: 'Type', sortable: true },
		{
			key: 'estimated_tokens',
			label: 'Tokens',
			sortable: true,
			align: 'right' as const,
			format: (v: unknown) => Number(v).toLocaleString()
		},
		{
			key: 'preview',
			label: 'Preview',
			format: (v: unknown) => String(v).slice(0, 60) + '...'
		}
	];
</script>

<svelte:head>
	<title>Messages | Memory Dashboard</title>
</svelte:head>

<section class="page-header">
	<div class="container">
		{#if mounted}
			<div in:fade={{ duration: 300 }}>
				<Breadcrumbs {crumbs} />
			</div>
			<h1 in:fly={{ y: 30, duration: 800, delay: 100 }}>Messages</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 200 }}>
				Patterns in conversations & context window analysis
			</p>
		{/if}
	</div>
</section>

<section class="page-content">
	<div class="container">
		{#if loading}
			<div class="loading-state" in:fade={{ duration: 300 }}>
				<div class="spinner"></div>
				<p>Loading message data...</p>
			</div>
		{:else if error}
			<div class="error-state glass-card" in:fade={{ duration: 300 }}>
				<h3>Failed to load data</h3>
				<p>{error}</p>
			</div>
		{:else}
			<div class="visualizations" in:fly={{ y: 30, duration: 600 }}>
				<!-- Context Window Stats -->
				{#if contextStatsData.length > 0}
					<StatsGrid stats={contextStatsData} title="Context Window Analysis" />
				{/if}

				{#if trendData.length > 0}
					<TrendChart data={trendData} title="Messages Over Time" />
				{/if}

				<div class="two-col">
					{#if roleChartData.length > 0}
						<DonutChart data={roleChartData} title="User vs Assistant" />
					{/if}

					{#if lengthChartData.length > 0}
						<BarChart data={lengthChartData} title="Message Length Distribution" />
					{/if}
				</div>

				<div class="two-col">
					{#if contentTypeChartData.length > 0}
						<DonutChart data={contentTypeChartData} title="Content Types" />
					{/if}
				</div>

				<!-- Large Messages Table -->
				{#if largeMessages.length > 0}
					<DataTable
						columns={largeMessageColumns}
						rows={largeMessages}
						title="Largest Messages (> 10K chars)"
						keyField="session_id"
					/>
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
		max-width: 900px;
		margin: 0 auto;
	}

	.two-col {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 30px;
		margin-top: 30px;
	}

	@media (max-width: 800px) {
		.two-col {
			grid-template-columns: 1fr;
		}
	}
</style>
