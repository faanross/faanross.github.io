<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { fade, fly } from 'svelte/transition';
	import {
		loadParquetFiles,
		getSessionById,
		getSessionMessages,
		getSessionToolCalls,
		type SessionDetail,
		type MessageDetail,
		type ToolCount
	} from '$lib/db/duckdb';
	import Breadcrumbs from '../../components/Breadcrumbs.svelte';
	import BarChart from '../../components/BarChart.svelte';

	let mounted = $state(false);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let session = $state<SessionDetail | null>(null);
	let messages = $state<MessageDetail[]>([]);
	let toolCalls = $state<ToolCount[]>([]);

	const sessionId = $derived($page.params.id);

	const crumbs = $derived([
		{ label: 'Dashboard', href: '/claude/memory' },
		{ label: 'Sessions', href: '/claude/memory/sessions' },
		{ label: session?.project_name || sessionId.slice(0, 8) + '...' }
	]);

	onMount(async () => {
		mounted = true;

		try {
			await loadParquetFiles();
			const [sessionData, messagesData, toolsData] = await Promise.all([
				getSessionById(sessionId),
				getSessionMessages(sessionId),
				getSessionToolCalls(sessionId)
			]);
			session = sessionData;
			messages = messagesData;
			toolCalls = toolsData;
		} catch (e) {
			console.error('Failed to load session:', e);
			error = e instanceof Error ? e.message : 'Unknown error';
		} finally {
			loading = false;
		}
	});

	function formatDateTime(dateStr: string): string {
		if (!dateStr) return '';
		const numVal = Number(dateStr);
		let date: Date;
		if (!isNaN(numVal) && numVal > 1000000000) {
			date = numVal > 1000000000000 ? new Date(numVal) : new Date(numVal * 1000);
		} else {
			date = new Date(dateStr);
		}
		if (isNaN(date.getTime())) return dateStr.slice(0, 19);
		return date.toLocaleString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getDuration(start: string, end: string): string {
		const startNum = Number(start);
		const endNum = Number(end);
		let startDate: Date, endDate: Date;

		if (!isNaN(startNum) && startNum > 1000000000) {
			startDate = startNum > 1000000000000 ? new Date(startNum) : new Date(startNum * 1000);
		} else {
			startDate = new Date(start);
		}

		if (!isNaN(endNum) && endNum > 1000000000) {
			endDate = endNum > 1000000000000 ? new Date(endNum) : new Date(endNum * 1000);
		} else {
			endDate = new Date(end);
		}

		const diffMs = endDate.getTime() - startDate.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		if (diffMins < 60) return `${diffMins}m`;
		const hours = Math.floor(diffMins / 60);
		const mins = diffMins % 60;
		return `${hours}h ${mins}m`;
	}

	function truncateContent(content: string, maxLen: number = 500): string {
		if (content.length <= maxLen) return content;
		return content.slice(0, maxLen) + '...';
	}

	function estimateTokens(content: string): number {
		return Math.round(content.length / 4);
	}

	const totalTokens = $derived(
		messages.reduce((sum, m) => sum + estimateTokens(m.content), 0)
	);

	const toolChartData = $derived(
		toolCalls.map((t) => ({
			label: t.tool_name,
			value: t.count
		}))
	);
</script>

<svelte:head>
	<title>{session?.project_name || 'Session'} | Memory Dashboard</title>
</svelte:head>

<section class="page-header">
	<div class="container">
		{#if mounted}
			<div in:fade={{ duration: 300 }}>
				<Breadcrumbs crumbs={crumbs} />
			</div>
			<h1 in:fly={{ y: 30, duration: 800, delay: 100 }}>Session Detail</h1>
			{#if session}
				<p class="lead" in:fly={{ y: 20, duration: 600, delay: 200 }}>
					{session.project_name}
				</p>
			{/if}
		{/if}
	</div>
</section>

<section class="page-content">
	<div class="container">
		{#if loading}
			<div class="loading-state" in:fade={{ duration: 300 }}>
				<div class="spinner"></div>
				<p>Loading session...</p>
			</div>
		{:else if error}
			<div class="error-state glass-card" in:fade={{ duration: 300 }}>
				<h3>Failed to load session</h3>
				<p>{error}</p>
			</div>
		{:else if !session}
			<div class="error-state glass-card" in:fade={{ duration: 300 }}>
				<h3>Session not found</h3>
				<p>No session with ID: {sessionId}</p>
			</div>
		{:else}
			<div class="session-detail" in:fly={{ y: 30, duration: 600 }}>
				<!-- Session Overview -->
				<div class="overview-grid">
					<div class="stat-card">
						<span class="stat-value">{session.message_count}</span>
						<span class="stat-label">Messages</span>
					</div>
					<div class="stat-card">
						<span class="stat-value">{session.user_message_count}</span>
						<span class="stat-label">User</span>
					</div>
					<div class="stat-card">
						<span class="stat-value">{session.assistant_message_count}</span>
						<span class="stat-label">Assistant</span>
					</div>
					<div class="stat-card">
						<span class="stat-value">{totalTokens.toLocaleString()}</span>
						<span class="stat-label">Est. Tokens</span>
					</div>
					<div class="stat-card">
						<span class="stat-value">{getDuration(session.first_message_at, session.last_message_at)}</span>
						<span class="stat-label">Duration</span>
					</div>
					<div class="stat-card">
						<span class="stat-value">{toolCalls.reduce((s, t) => s + t.count, 0)}</span>
						<span class="stat-label">Tool Calls</span>
					</div>
				</div>

				<!-- Metadata -->
				<div class="metadata glass-card">
					<div class="meta-row">
						<span class="meta-label">Session ID</span>
						<span class="meta-value mono">{session.session_id}</span>
					</div>
					<div class="meta-row">
						<span class="meta-label">Project Path</span>
						<span class="meta-value mono">{session.project_path}</span>
					</div>
					<div class="meta-row">
						<span class="meta-label">Started</span>
						<span class="meta-value">{formatDateTime(session.first_message_at)}</span>
					</div>
					<div class="meta-row">
						<span class="meta-label">Ended</span>
						<span class="meta-value">{formatDateTime(session.last_message_at)}</span>
					</div>
				</div>

				<!-- Tool Calls -->
				{#if toolChartData.length > 0}
					<BarChart data={toolChartData} title="Tools Used in This Session" />
				{/if}

				<!-- Message Timeline -->
				<div class="timeline-section">
					<h3>Message Timeline</h3>
					<div class="timeline">
						{#each messages as message, i}
							<div class="message {message.type}">
								<div class="message-header">
									<span class="message-role">{message.type}</span>
									<span class="message-time">{formatDateTime(message.timestamp)}</span>
									<span class="message-tokens">{estimateTokens(message.content).toLocaleString()} tokens</span>
								</div>
								{#if message.tool_name}
									<div class="tool-badge">
										Tool: {message.tool_name}
									</div>
								{/if}
								<div class="message-content">
									{truncateContent(message.content)}
								</div>
							</div>
						{/each}
					</div>
				</div>
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
		margin-bottom: 8px;
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

	.session-detail {
		max-width: 900px;
		margin: 0 auto;
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

	/* Overview Grid */
	.overview-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: 16px;
		margin-bottom: 30px;
	}

	.stat-card {
		background: rgba(61, 61, 71, 0.3);
		border-radius: 8px;
		padding: 20px;
		text-align: center;
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.stat-value {
		display: block;
		font-size: 24px;
		font-weight: 700;
		color: var(--aion-yellow);
		font-family: 'JetBrains Mono', monospace;
		margin-bottom: 4px;
	}

	.stat-label {
		font-size: 11px;
		color: rgba(255, 255, 255, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	/* Metadata */
	.metadata {
		padding: 20px;
		margin-bottom: 30px;
	}

	.meta-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 8px 0;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.meta-row:last-child {
		border-bottom: none;
	}

	.meta-label {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.meta-value {
		font-size: 13px;
		color: rgba(255, 255, 255, 0.8);
	}

	.meta-value.mono {
		font-family: 'JetBrains Mono', monospace;
		font-size: 11px;
		color: rgba(255, 255, 255, 0.6);
	}

	/* Timeline */
	.timeline-section {
		margin-top: 40px;
	}

	.timeline-section h3 {
		margin-bottom: 20px;
		font-size: 16px;
		font-weight: 600;
		color: var(--aion-purple);
	}

	.timeline {
		display: flex;
		flex-direction: column;
		gap: 16px;
		max-height: 600px;
		overflow-y: auto;
		padding-right: 8px;
	}

	.timeline::-webkit-scrollbar {
		width: 6px;
	}

	.timeline::-webkit-scrollbar-track {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 3px;
	}

	.timeline::-webkit-scrollbar-thumb {
		background: rgba(189, 147, 249, 0.3);
		border-radius: 3px;
	}

	.message {
		background: rgba(61, 61, 71, 0.2);
		border-radius: 8px;
		padding: 16px;
		border-left: 3px solid rgba(255, 255, 255, 0.2);
	}

	.message.user {
		border-left-color: var(--aion-yellow);
	}

	.message.assistant {
		border-left-color: var(--aion-purple);
	}

	.message-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 8px;
	}

	.message-role {
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 2px 8px;
		border-radius: 4px;
		background: rgba(255, 255, 255, 0.1);
	}

	.message.user .message-role {
		color: var(--aion-yellow);
		background: rgba(255, 232, 96, 0.1);
	}

	.message.assistant .message-role {
		color: var(--aion-purple-light);
		background: rgba(189, 147, 249, 0.1);
	}

	.message-time {
		font-size: 11px;
		color: rgba(255, 255, 255, 0.4);
		font-family: 'JetBrains Mono', monospace;
	}

	.message-tokens {
		font-size: 10px;
		color: rgba(255, 255, 255, 0.3);
		margin-left: auto;
	}

	.tool-badge {
		display: inline-block;
		font-size: 10px;
		color: var(--aion-purple-light);
		background: rgba(189, 147, 249, 0.15);
		padding: 2px 8px;
		border-radius: 4px;
		margin-bottom: 8px;
	}

	.message-content {
		font-size: 13px;
		line-height: 1.5;
		color: rgba(255, 255, 255, 0.7);
		white-space: pre-wrap;
		word-break: break-word;
	}

	@media (max-width: 600px) {
		.overview-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.meta-row {
			flex-direction: column;
			align-items: flex-start;
			gap: 4px;
		}
	}
</style>
