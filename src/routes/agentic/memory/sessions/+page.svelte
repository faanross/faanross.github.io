<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { fade, fly } from 'svelte/transition';
	import {
		loadParquetFiles,
		getActivityHeatmap,
		getSessionList,
		getSessionDurationStats,
		type HeatmapData,
		type SessionSummary,
		type SessionDurationStats
	} from '$lib/db/duckdb';
	import Heatmap from '../components/Heatmap.svelte';
	import DataTable from '../components/DataTable.svelte';
	import StatsGrid from '../components/StatsGrid.svelte';
	import Breadcrumbs from '../components/Breadcrumbs.svelte';

	let mounted = $state(false);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let heatmapData = $state<HeatmapData[]>([]);
	let sessions = $state<SessionSummary[]>([]);
	let durationStats = $state<SessionDurationStats | null>(null);

	const crumbs = [
		{ label: 'Dashboard', href: '/claude/memory' },
		{ label: 'Sessions' }
	];

	onMount(async () => {
		mounted = true;

		try {
			await loadParquetFiles();
			[heatmapData, sessions, durationStats] = await Promise.all([
				getActivityHeatmap(),
				getSessionList(50),
				getSessionDurationStats()
			]);
		} catch (e) {
			console.error('Failed to load data:', e);
			error = e instanceof Error ? e.message : 'Unknown error';
		} finally {
			loading = false;
		}
	});

	const statsData = $derived(
		durationStats
			? [
					{ label: 'Total Hours', value: durationStats.total_hours },
					{ label: 'Avg Session', value: `${durationStats.avg_duration_minutes}m` },
					{ label: 'Longest Session', value: `${Math.round(durationStats.max_duration_minutes / 60)}h` },
					{ label: 'Total Sessions', value: durationStats.total_sessions }
				]
			: []
	);

	// Table columns
	const tableColumns = [
		{
			key: 'project_name',
			label: 'Project',
			sortable: true,
			format: (v: unknown) => {
				const name = String(v ?? '');
				const parts = name.split(/[\\/]/);
				return parts[parts.length - 1] || parts[parts.length - 2] || name;
			}
		},
		{
			key: 'start_time',
			label: 'Date',
			sortable: true,
			format: (v: unknown) => {
				const dateStr = String(v ?? '');
				const numVal = Number(dateStr);
				let date: Date;
				if (!isNaN(numVal) && numVal > 1000000000) {
					date = numVal > 1000000000000 ? new Date(numVal) : new Date(numVal * 1000);
				} else {
					date = new Date(dateStr);
				}
				if (isNaN(date.getTime())) return dateStr.slice(0, 10);
				return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
			}
		},
		{ key: 'message_count', label: 'Messages', sortable: true, align: 'right' as const },
		{
			key: 'duration_minutes',
			label: 'Duration',
			sortable: true,
			align: 'right' as const,
			format: (v: unknown) => {
				const mins = Number(v ?? 0);
				if (mins < 60) return `${mins}m`;
				const hours = Math.floor(mins / 60);
				const m = mins % 60;
				return m > 0 ? `${hours}h ${m}m` : `${hours}h`;
			}
		}
	];

	function handleRowClick(row: Record<string, unknown>) {
		const sessionId = row.session_id as string;
		goto(`/claude/memory/sessions/${sessionId}`);
	}
</script>

<svelte:head>
	<title>Sessions | Memory Dashboard</title>
</svelte:head>

<section class="page-header">
	<div class="container">
		{#if mounted}
			<div in:fade={{ duration: 300 }}>
				<Breadcrumbs {crumbs} />
			</div>
			<h1 in:fly={{ y: 30, duration: 800, delay: 100 }}>Sessions</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 200 }}>
				When and how I work with Claude
			</p>
		{/if}
	</div>
</section>

<section class="page-content">
	<div class="container">
		{#if loading}
			<div class="loading-state" in:fade={{ duration: 300 }}>
				<div class="spinner"></div>
				<p>Loading session data...</p>
			</div>
		{:else if error}
			<div class="error-state glass-card" in:fade={{ duration: 300 }}>
				<h3>Failed to load data</h3>
				<p>{error}</p>
			</div>
		{:else}
			<div class="visualizations" in:fly={{ y: 30, duration: 600 }}>
				{#if statsData.length > 0}
					<StatsGrid stats={statsData} title="Session Duration Stats" />
				{/if}

				{#if heatmapData.length > 0}
					<Heatmap data={heatmapData} />
				{/if}

				{#if sessions.length > 0}
					<DataTable
						columns={tableColumns}
						rows={sessions}
						title="Recent Sessions (click to view details)"
						onRowClick={handleRowClick}
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
</style>
