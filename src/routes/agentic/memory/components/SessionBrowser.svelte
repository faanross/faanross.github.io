<script lang="ts">
	import type { SessionSummary } from '$lib/db/duckdb';

	let { sessions = [] }: { sessions: SessionSummary[] } = $props();

	// Format date/time
	function formatDateTime(dateStr: string): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		if (isNaN(date.getTime())) {
			const parts = String(dateStr).match(/(\d{4})-(\d{2})-(\d{2})/);
			if (parts) {
				return `${parts[2]}/${parts[3]}`;
			}
			return String(dateStr).slice(0, 10);
		}
		return date.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatDuration(minutes: number): string {
		if (minutes < 60) return `${minutes}m`;
		const hours = Math.floor(minutes / 60);
		const mins = minutes % 60;
		return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
	}

	function shortenProjectName(name: string): string {
		if (!name) return 'Unknown';
		// Extract the last meaningful part of the path
		const parts = name.split(/[\\/]/);
		return parts[parts.length - 1] || parts[parts.length - 2] || name;
	}
</script>

<div class="session-browser">
	<h3>Recent Sessions</h3>

	<div class="sessions-list">
		{#each sessions as session}
			<div class="session-card">
				<div class="session-header">
					<span class="session-project">{shortenProjectName(session.project_name)}</span>
					<span class="session-date">{formatDateTime(session.start_time)}</span>
				</div>
				<div class="session-stats">
					<div class="stat">
						<span class="stat-value">{session.message_count}</span>
						<span class="stat-label">messages</span>
					</div>
					<div class="stat">
						<span class="stat-value">{formatDuration(session.duration_minutes)}</span>
						<span class="stat-label">duration</span>
					</div>
				</div>
				<div class="session-id">{session.session_id.slice(0, 8)}...</div>
			</div>
		{:else}
			<p class="no-sessions">No sessions found</p>
		{/each}
	</div>
</div>

<style>
	.session-browser {
		margin-bottom: 40px;
	}

	h3 {
		margin-bottom: 20px;
		font-size: 16px;
		font-weight: 600;
		color: var(--aion-purple);
	}

	.sessions-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
		max-height: 500px;
		overflow-y: auto;
		padding-right: 8px;
	}

	.sessions-list::-webkit-scrollbar {
		width: 6px;
	}

	.sessions-list::-webkit-scrollbar-track {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 3px;
	}

	.sessions-list::-webkit-scrollbar-thumb {
		background: rgba(189, 147, 249, 0.3);
		border-radius: 3px;
	}

	.session-card {
		background: rgba(61, 61, 71, 0.2);
		border-radius: 8px;
		padding: 16px;
		border: 1px solid rgba(255, 255, 255, 0.05);
		transition: border-color 0.2s;
	}

	.session-card:hover {
		border-color: rgba(189, 147, 249, 0.3);
	}

	.session-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.session-project {
		font-size: 14px;
		font-weight: 600;
		color: var(--aion-purple-light);
	}

	.session-date {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.5);
		font-family: 'JetBrains Mono', monospace;
	}

	.session-stats {
		display: flex;
		gap: 24px;
		margin-bottom: 8px;
	}

	.stat {
		display: flex;
		align-items: baseline;
		gap: 6px;
	}

	.stat-value {
		font-size: 18px;
		font-weight: 700;
		color: var(--aion-yellow);
		font-family: 'JetBrains Mono', monospace;
	}

	.stat-label {
		font-size: 11px;
		color: rgba(255, 255, 255, 0.4);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.session-id {
		font-size: 10px;
		color: rgba(255, 255, 255, 0.3);
		font-family: 'JetBrains Mono', monospace;
	}

	.no-sessions {
		text-align: center;
		color: rgba(255, 255, 255, 0.5);
		padding: 40px 0;
	}
</style>
